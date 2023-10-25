package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type PostLivecommentRequest struct {
	Comment string `json:"comment"`
	Tip     int64  `json:"tip"`
}

type LivecommentModel struct {
	Id           int64     `db:"id"`
	UserId       int64     `db:"user_id"`
	LivestreamId int64     `db:"livestream_id"`
	Comment      string    `db:"comment"`
	Tip          int64     `db:"tip"`
	ReportCount  int64     `db:"report_count"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type Livecomment struct {
	Id          int64      `json:"id"`
	User        User       `json:"user"`
	Livestream  Livestream `json:"livestream"`
	Comment     string     `json:"comment"`
	Tip         int64      `json:"tip"`
	ReportCount int64      `json:"report_count"`
	CreatedAt   int64      `json:"created_at"`
	UpdatedAt   int64      `json:"updated_at"`
}

type LivecommentReport struct {
	Id          int64       `json:"id"`
	Reporter    User        `json:"reporter"`
	Livecomment Livecomment `json:"livecomment"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
}

type LivecommentReportModel struct {
	Id            int64     `db:"id"`
	UserId        int64     `db:"user_id"`
	LivestreamId  int64     `db:"livestream_id"`
	LivecommentId int64     `db:"livecomment_id"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type ModerateRequest struct {
	NGWord string `json:"ng_word"`
}

type NGWord struct {
	UserId       int64  `json:"user_id" db:"user_id"`
	LivestreamId int64  `json:"livestream_id" db:"livestream_id"`
	Word         string `json:"word" db:"word"`
}

func getLivecommentsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	livestreamId, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	livecommentModels := []LivecommentModel{}
	err = tx.SelectContext(ctx, &livecommentModels, "SELECT * FROM livecomments WHERE livestream_id = ?", livestreamId)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	livecomments := make([]Livecomment, len(livecommentModels))
	for i := range livecommentModels {
		livecomment, err := fillLivecommentResponse(ctx, tx, livecommentModels[i])
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		livecomments[i] = livecomment
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, livecomments)
}

func postLivecommentHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamId, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	sess, err := session.Get(defaultSessionIdKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	userId, ok := sess.Values[defaultUserIdKey].(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to find user-id from session")
	}

	var req *PostLivecommentRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	var hitSpam int
	query := `
	SELECT COUNT(*) AS cnt
	FROM ng_words AS w
	CROSS JOIN
	(SELECT ? AS text) AS t
	WHERE t.text LIKE CONCAT('%', w.word, '%');
	`
	if err = tx.GetContext(ctx, &hitSpam, query, req.Comment); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	c.Logger().Infof("[hitSpam=%d] comment = %s", hitSpam, req.Comment)
	if hitSpam >= 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "このコメントがスパム判定されました")
	}

	now := time.Now()
	livecommentModel := LivecommentModel{
		UserId:       int64(userId),
		LivestreamId: int64(livestreamId),
		Comment:      req.Comment,
		Tip:          req.Tip,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at, updated_at) VALUES (:user_id, :livestream_id, :comment, :tip, :created_at, :updated_at)", livecommentModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	livecommentId, err := rs.LastInsertId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	livecommentModel.Id = livecommentId

	livecomment, err := fillLivecommentResponse(ctx, tx, livecommentModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, livecomment)
}

func reportLivecommentHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamId, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	livecommentId, err := strconv.Atoi(c.Param("livecomment_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	sess, err := session.Get(defaultSessionIdKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	userId, ok := sess.Values[defaultUserIdKey].(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	// 配信者自身の配信に対するGETなのかを検証
	var ownedLivestreams []*LivestreamModel
	if err := tx.SelectContext(ctx, &ownedLivestreams, "SELECT * FROM livestreams WHERE id = ? AND user_id = ?", livestreamId, userId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(ownedLivestreams) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "A streamer can't get livecomment reports that other streamers own")
	}

	now := time.Now()
	reportModel := LivecommentReportModel{
		UserId:        int64(userId),
		LivestreamId:  int64(livestreamId),
		LivecommentId: int64(livecommentId),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	rs, err := tx.NamedExecContext(ctx, "INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id, created_at, updated_at) VALUES (:user_id, :livestream_id, :livecomment_id, :created_at, :updated_at)", &reportModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if _, err := tx.ExecContext(ctx, "UPDATE livecomments SET report_count = report_count + 1 WHERE id = ?", livecommentId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reportId, err := rs.LastInsertId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	reportModel.Id = reportId

	report, err := fillLivecommentReportResponse(ctx, tx, reportModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, report)
}

// NGワードを登録
func moderateNGWordHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamId, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	sess, err := session.Get(defaultSessionIdKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	userId, ok := sess.Values[defaultUserIdKey].(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	var req *ModerateRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	// 配信者自身の配信に対するmoderateなのかを検証
	var ownedLivestreams []*LivestreamModel
	if err := tx.SelectContext(ctx, &ownedLivestreams, "SELECT * FROM livestreams WHERE id = ? AND user_id = ?", livestreamId, userId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(ownedLivestreams) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "A streamer can't moderate livestreams that other streamers own")
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO ng_words(user_id, livestream_id, word) VALUES (:user_id, :livestream_id, :word)", &NGWord{
		UserId:       int64(userId),
		LivestreamId: int64(livestreamId),
		Word:         req.NGWord,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	wordId, err := rs.LastInsertId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"word_id": wordId,
	})
}

func fillLivecommentResponse(ctx context.Context, tx *sqlx.Tx, livecommentModel LivecommentModel) (Livecomment, error) {
	commentOwnerModel := UserModel{}
	if err := tx.GetContext(ctx, &commentOwnerModel, "SELECT * FROM users WHERE id = ?", livecommentModel.UserId); err != nil {
		return Livecomment{}, err
	}
	commentOwner, err := fillUserResponse(ctx, tx, commentOwnerModel)
	if err != nil {
		return Livecomment{}, err
	}

	livestreamModel := LivestreamModel{}
	if err := tx.GetContext(ctx, &livestreamModel, "SELECT * FROM livestreams WHERE id = ?", livecommentModel.LivestreamId); err != nil {
		return Livecomment{}, err
	}
	livestream, err := fillLivestreamResponse(ctx, tx, livestreamModel)
	if err != nil {
		return Livecomment{}, err
	}

	livecomment := Livecomment{
		Id:          livecommentModel.Id,
		User:        commentOwner,
		Livestream:  livestream,
		Comment:     livecommentModel.Comment,
		Tip:         livecommentModel.Tip,
		ReportCount: livecommentModel.ReportCount,
		CreatedAt:   livecommentModel.CreatedAt.Unix(),
		UpdatedAt:   livecommentModel.UpdatedAt.Unix(),
	}

	return livecomment, nil
}

func fillLivecommentReportResponse(ctx context.Context, tx *sqlx.Tx, reportModel LivecommentReportModel) (LivecommentReport, error) {
	reporterModel := UserModel{}
	if err := tx.GetContext(ctx, &reporterModel, "SELECT * FROM users WHERE id = ?", reportModel.UserId); err != nil {
		return LivecommentReport{}, err
	}
	reporter, err := fillUserResponse(ctx, tx, reporterModel)
	if err != nil {
		return LivecommentReport{}, err
	}

	livecommentModel := LivecommentModel{}
	if err := tx.GetContext(ctx, &livecommentModel, "SELECT * FROM livecomments WHERE id = ?", reportModel.LivecommentId); err != nil {
		return LivecommentReport{}, err
	}
	livecomment, err := fillLivecommentResponse(ctx, tx, livecommentModel)
	if err != nil {
		return LivecommentReport{}, err
	}

	report := LivecommentReport{
		Id:          reportModel.Id,
		Reporter:    reporter,
		Livecomment: livecomment,
		CreatedAt:   reportModel.CreatedAt.Unix(),
		UpdatedAt:   reportModel.UpdatedAt.Unix(),
	}
	return report, nil
}
