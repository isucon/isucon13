package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type PostLivecommentRequest struct {
	Comment string `json:"comment"`
	Tip     int    `json:"tip"`
}

type Livecomment struct {
	Id           int       `json:"id" db:"id"`
	UserId       int       `json:"user_id" db:"user_id"`
	LivestreamId int       `json:"livestream_id" db:"livestream_id"`
	Comment      string    `json:"comment" db:"comment"`
	Tip          int       `json:"tip" db:"tip"`
	ReportCount  int       `json:"report_count" db:"report_count"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type LivecommentReport struct {
	Id            int       `db:"id"`
	UserId        int       `db:"user_id"`
	LivestreamId  int       `db:"livestream_id"`
	LivecommentId int       `db:"livecomment_id"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type ModerateRequest struct {
	NGWord string `json:"ng_word"`
}

type NGWord struct {
	UserId       int    `db:"user_id"`
	LivestreamId int    `db:"livestream_id"`
	Word         string `db:"word"`
}

func getLivecommentsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	livestreamId := c.Param("livestream_id")

	livecomments := []Livecomment{}
	if err := dbConn.SelectContext(ctx, &livecomments, "SELECT * FROM livecomments WHERE livestream_id = ?", livestreamId); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
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
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if req.Tip < 0 || 20000 < req.Tip {
		return echo.NewHTTPError(http.StatusBadRequest, "the tips in a live comment must be 1 <= tips <= 20000")
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var hitSpam int
	query := `
	SELECT COUNT(*) AS cnt
	FROM ng_words AS w
	CROSS JOIN
	(SELECT ? AS text) AS t
	WHERE t.text LIKE CONCAT('%', w.word, '%');
	`
	if err := tx.GetContext(ctx, &hitSpam, query, req.Comment); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	c.Logger().Infof("[hitSpam=%d] comment = %s", hitSpam, req.Comment)
	if hitSpam >= 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "このコメントがスパム判定されました")
	}

	livecomment := Livecomment{
		UserId:       userId,
		LivestreamId: livestreamId,
		Comment:      req.Comment,
		Tip:          req.Tip,
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO livecomments (user_id, livestream_id, comment, tip) VALUES (:user_id, :livestream_id, :comment, :tip)", livecomment)
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	livecommentId, err := rs.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	livecomment.Id = int(livecommentId)
	createdAt := time.Now()
	livecomment.CreatedAt = createdAt
	livecomment.UpdatedAt = createdAt
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

	// 配信者自身の配信に対するGETなのかを検証
	var ownedLivestreams []*Livestream
	if err := tx.SelectContext(ctx, &ownedLivestreams, "SELECT * FROM livestreams WHERE id = ? AND user_id = ?", livestreamId, userId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(ownedLivestreams) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "A streamer can't get livecomment reports that other streamers own")
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id) VALUES (:user_id, :livestream_id, :livecomment_id)", &LivecommentReport{
		UserId:        userId,
		LivestreamId:  livestreamId,
		LivecommentId: livecommentId,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if _, err := tx.ExecContext(ctx, "UPDATE livecomments SET report_count = report_count + 1 WHERE id = ?", livecommentId); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reportId, err := rs.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	createdAt := time.Now()
	report := &LivecommentReport{
		Id:            int(reportId),
		UserId:        userId,
		LivestreamId:  livestreamId,
		LivecommentId: livecommentId,
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}

	tx.Commit()

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
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// 配信者自身の配信に対するmoderateなのかを検証
	var ownedLivestreams []*Livestream
	if err := tx.SelectContext(ctx, &ownedLivestreams, "SELECT * FROM livestreams WHERE id = ? AND user_id = ?", livestreamId, userId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(ownedLivestreams) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "A streamer can't moderate livestreams that other streamers own")
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO ng_words(user_id, livestream_id, word) VALUES (:user_id, :livestream_id, :word)", &NGWord{
		UserId:       userId,
		LivestreamId: livestreamId,
		Word:         req.NGWord,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	wordId, err := rs.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx.Commit()

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"word_id": wordId,
	})
}
