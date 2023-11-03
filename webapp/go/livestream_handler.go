package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type ReserveLivestreamRequest struct {
	Tags        []int64 `json:"tags"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	// NOTE: コラボ配信の際に便利な自動スケジュールチェック機能
	// DBに記録しないが、コラボレーターがスケジュール的に問題ないか調べて、エラーを返す
	Collaborators []int64 `json:"collaborators"`
	StartAt       int64   `json:"start_at"`
	EndAt         int64   `json:"end_at"`
}

type LivestreamViewerModel struct {
	UserID       int64 `db:"user_id"`
	LivestreamID int64 `db:"livestream_id"`
	CreatedAt    int64 `db:"created_at"`
}

type LivestreamModel struct {
	ID           int64  `db:"id"`
	UserID       int64  `db:"user_id"`
	Title        string `db:"title"`
	Description  string `db:"description"`
	PlaylistUrl  string `db:"playlist_url"`
	ThumbnailUrl string `db:"thumbnail_url"`
	StartAt      int64  `db:"start_at"`
	EndAt        int64  `db:"end_at"`
}

type Livestream struct {
	ID           int64  `json:"id"`
	Owner        User   `json:"owner"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	PlaylistUrl  string `json:"playlist_url"`
	ThumbnailUrl string `json:"thumbnail_url"`
	Tags         []Tag  `json:"tags"`
	StartAt      int64  `json:"start_at"`
	EndAt        int64  `json:"end_at"`
}

type LivestreamTagModel struct {
	ID           int64 `db:"id"`
	LivestreamID int64 `db:"livestream_id"`
	TagID        int64 `db:"tag_id"`
}

type ReservationSlotModel struct {
	ID      int64 `db:"id"`
	Slot    int64 `db:"slot"`
	StartAt int64 `db:"start_at"`
	EndAt   int64 `db:"end_at"`
}

func reserveLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	userID, ok := sess.Values[defaultUserIDKey].(int64)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	var req *ReserveLivestreamRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	// 2024/04/01からの１年間の期間内であるかチェック
	c.Logger().Info("check term")
	var (
		termStartAt    = time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
		termEndAt      = time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
		reserveStartAt = time.Unix(req.StartAt, 0)
		reserveEndAt   = time.Unix(req.EndAt, 0)
	)
	if !(reserveEndAt.Equal(termEndAt) || reserveEndAt.Before(termEndAt)) && (reserveStartAt.Equal(termStartAt) || reserveStartAt.After(termStartAt)) {
		return echo.NewHTTPError(http.StatusBadRequest, "bad reservation time range")
	}

	c.Logger().Info("check collaborators")
	// 各ユーザについて、予約時間帯とかぶるような予約が存在しないか調べる (ある人は同時に複数の配信に物理的に出れない)
	var users []int64
	users = append(users, int64(userID))
	users = append(users, req.Collaborators...)
	for _, user := range users {
		var founds int
		if err := tx.GetContext(ctx, &founds, "SELECT COUNT(*) FROM livestreams WHERE user_id = ? AND  ? >= start_at AND ? <= end_at", user, reserveStartAt, reserveEndAt); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		if founds >= numReservationSlot {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ユーザ%dが予約できません", user))
		}
	}

	// 予約枠をみて、予約が可能か調べる
	var slots []*ReservationSlotModel
	if err := tx.SelectContext(ctx, &slots, "SELECT * FROM reservation_slots WHERE start_at >= ? AND end_at <= ?", req.StartAt, req.EndAt); err != nil {
		c.Logger().Warnf("予約枠一覧取得でエラー発生: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	for _, slot := range slots {
		var count int
		if err := tx.GetContext(ctx, &count, "SELECT slot FROM reservation_slots WHERE start_at = ? AND end_at = ?", slot.StartAt, slot.EndAt); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		c.Logger().Infof("%d ~ %d予約枠の残数 = %d\n", slot.StartAt, slot.EndAt, slot.Slot)
		if count < 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("予約区間 %d ~ %dが予約できません", req.StartAt, req.EndAt))
		}
	}

	livestreamModel := &LivestreamModel{
		UserID:      int64(userID),
		Title:       req.Title,
		Description: req.Description,
		// FIXME: プレイリスト、サムネイルは配信環境より配信されるので、それらのURLをPOSTできるようにする
		PlaylistUrl:  "https://d2jpkt808jogxx.cloudfront.net/BigBuckBunny/playlist.m3u8",
		ThumbnailUrl: "https://picsum.photos/200/300",
		StartAt:      req.StartAt,
		EndAt:        req.EndAt,
	}

	c.Logger().Info("insert reservation slot")
	if _, err := tx.ExecContext(ctx, "UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= ? AND end_at <= ?", req.StartAt, req.EndAt); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.Logger().Info("insert livestream")
	rs, err := tx.NamedExecContext(ctx, "INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(:user_id, :title, :description, :playlist_url, :thumbnail_url, :start_at, :end_at)", livestreamModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.Logger().Info("get inserted id")
	livestreamID, err := rs.LastInsertId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	livestreamModel.ID = livestreamID

	c.Logger().Info("insert tags")
	// タグ追加
	for _, tagID := range req.Tags {
		if _, err := tx.NamedExecContext(ctx, "INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (:livestream_id, :tag_id)", &LivestreamTagModel{
			LivestreamID: livestreamID,
			TagID:        tagID,
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	livestream, err := fillLivestreamResponse(ctx, tx, *livestreamModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, livestream)
}

func searchLivestreamsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	keyTagName := c.QueryParam("tag")

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	var livestreamModels []*LivestreamModel
	if c.QueryParam("tag") != "" {
		// タグによる取得
		var tagIDList []int
		if err := tx.SelectContext(ctx, &tagIDList, "SELECT id FROM tags WHERE name = ?", keyTagName); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		query, params, err := sqlx.In("SELECT * FROM livestream_tags WHERE id IN (?)", tagIDList)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		var keyTaggedLivestreams []*LivestreamTagModel
		if err := tx.SelectContext(ctx, &keyTaggedLivestreams, query, params...); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		for _, keyTaggedLivestream := range keyTaggedLivestreams {
			ls := LivestreamModel{}
			if err := tx.GetContext(ctx, &ls, "SELECT * FROM livestreams WHERE id = ?", keyTaggedLivestream.LivestreamID); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}

			livestreamModels = append(livestreamModels, &ls)
		}
	} else {
		// 検索条件なし
		query := `SELECT * FROM livestreams`
		if c.QueryParam("limit") != "" {
			limit, err := strconv.Atoi(c.QueryParam("limit"))
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			query += fmt.Sprintf(" LIMIT %d", limit)
		}

		if err := tx.SelectContext(ctx, &livestreamModels, query); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	livestreams := make([]Livestream, len(livestreamModels))
	for i := range livestreamModels {
		livestream, err := fillLivestreamResponse(ctx, tx, *livestreamModels[i])
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		livestreams[i] = livestream
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, livestreams)
}

func getUserLivestreamsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	userID, ok := sess.Values[defaultUserIDKey].(int64)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to find user-id from session")
	}

	var livestreams []*LivestreamModel
	if err := tx.SelectContext(ctx, &livestreams, "SELECT * FROM livestreams WHERE user_id = ?", userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, livestreams)
}

// viewerテーブルの廃止
func enterLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	userID, ok := sess.Values[defaultUserIDKey].(int64)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to find user-id from session")
	}

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	viewer := LivestreamViewerModel{
		UserID:       int64(userID),
		LivestreamID: int64(livestreamID),
		CreatedAt:    time.Now().Unix(),
	}

	if _, err := tx.NamedExecContext(ctx, "INSERT INTO livestream_viewers_history (user_id, livestream_id, created_at) VALUES(:user_id, :livestream_id, :created_at)", viewer); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func leaveLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}
	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	userID, ok := sess.Values[defaultUserIDKey].(int64)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to find user-id from session")
	}

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM livestream_viewers_history WHERE user_id = ? AND livestream_id = ?", userID, livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func getLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	livestreamModel := LivestreamModel{}
	err = tx.GetContext(ctx, &livestreamModel, "SELECT * FROM livestreams WHERE id = ?", livestreamID)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	livestream, err := fillLivestreamResponse(ctx, tx, livestreamModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, livestream)
}

func getLivecommentReportsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	var livestreamModel LivestreamModel
	if err := tx.GetContext(ctx, &livestreamModel, "SELECT * FROM livestreams WHERE id = ?", livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	userID, ok := sess.Values[defaultUserIDKey].(int64)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if livestreamModel.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "can't get other streamer's livecomment reports")
	}

	var reportModels []*LivecommentReportModel
	if err := tx.SelectContext(ctx, &reportModels, "SELECT * FROM livecomment_reports WHERE livestream_id = ?", livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reports := make([]LivecommentReport, len(reportModels))
	for i := range reportModels {
		report, err := fillLivecommentReportResponse(ctx, tx, *reportModels[i])
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		reports[i] = report
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, reports)
}

func fillLivestreamResponse(ctx context.Context, tx *sqlx.Tx, livestreamModel LivestreamModel) (Livestream, error) {
	ownerModel := UserModel{}
	if err := tx.GetContext(ctx, &ownerModel, "SELECT * FROM users WHERE id = ?", livestreamModel.UserID); err != nil {
		return Livestream{}, err
	}
	owner, err := fillUserResponse(ctx, tx, ownerModel)
	if err != nil {
		return Livestream{}, err
	}

	var livestreamTagModels []*LivestreamTagModel
	if err := tx.SelectContext(ctx, &livestreamTagModels, "SELECT * FROM livestream_tags WHERE livestream_id = ?", livestreamModel.ID); err != nil {
		return Livestream{}, err
	}

	tags := make([]Tag, len(livestreamTagModels))
	for i := range livestreamTagModels {
		tagModel := TagModel{}
		if err := tx.GetContext(ctx, &tagModel, "SELECT * FROM tags WHERE id = ?", livestreamTagModels[i].TagID); err != nil {
			return Livestream{}, err
		}

		tags[i] = Tag{
			ID:   tagModel.ID,
			Name: tagModel.Name,
		}
	}

	livestream := Livestream{
		ID:           livestreamModel.ID,
		Owner:        owner,
		Title:        livestreamModel.Title,
		Tags:         tags,
		Description:  livestreamModel.Description,
		PlaylistUrl:  livestreamModel.PlaylistUrl,
		ThumbnailUrl: livestreamModel.ThumbnailUrl,
		StartAt:      livestreamModel.StartAt,
		EndAt:        livestreamModel.EndAt,
	}
	return livestream, nil
}
