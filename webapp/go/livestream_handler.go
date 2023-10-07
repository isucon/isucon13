package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type ReserveLivestreamRequest struct {
	Tags        []int  `json:"tags"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// NOTE: コラボ配信の際に便利な自動スケジュールチェック機能
	// DBに記録しないが、コラボレーターがスケジュール的に問題ないか調べて、エラーを返す
	Collaborators []int `json:"collaborators"`
	StartAt       int64 `json:"start_at"`
	EndAt         int64 `json:"end_at"`
}

type LivestreamViewer struct {
	Id           int `json:"id" db:"id"`
	UserId       int `json:"user_id" db:"user_id"`
	LivestreamId int `json:"livestream_id" db:"livestream_id"`
}

type Livestream struct {
	Id           int       `json:"id" db:"id"`
	UserId       int       `json:"user_id" db:"user_id"`
	Title        string    `json:"title" db:"title"`
	Description  string    `json:"description" db:"description"`
	PlaylistUrl  string    `json:"playlist_url" db:"playlist_url"`
	ThumbnailUrl string    `json:"thumbnail_url" db:"thumbnail_url"`
	ViewersCount int       `json:"viewers_count" db:"viewers_count"`
	StartAt      time.Time `json:"start_at" db:"start_at"`
	EndAt        time.Time `json:"end_at" db:"end_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type LivestreamTag struct {
	Id           int `json:"id" db:"id"`
	LivestreamId int `json:"livestream_id" db:"livestream_id"`
	TagId        int `json:"tag_id" db:"tag_id"`
}

func reserveLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	sess, err := session.Get(defaultSessionIdKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	userId, ok := sess.Values[defaultUserIdKey].(int)
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

	// 2024/04/01 - 2025/03/31までの期間かチェック
	var (
		termStartAt    = time.Date(2024, 4, 1, 0, 0, 0, 0, time.Local)
		termEndAt      = time.Date(2025, 3, 31, 0, 0, 0, 0, time.Local)
		reserveStartAt = time.Unix(req.StartAt, 0)
		reserveEndAt   = time.Unix(req.EndAt, 0)
	)
	if !(reserveEndAt.Equal(termEndAt) || reserveEndAt.Before(termEndAt)) && (reserveStartAt.Equal(termStartAt) || reserveStartAt.After(termStartAt)) {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, "bad reservation time range")
	}

	// 各ユーザについて、予約時間帯とかぶるような予約が存在しないか調べる
	var users []int
	users = append(users, userId)
	users = append(users, req.Collaborators...)
	for _, user := range users {
		var founds int
		if err := tx.SelectContext(ctx, &founds, "SELECT COUNT(*) FROM livestreams WHERE user_id = ? AND  ? >= start_at AND ? <= end_at", user, reserveStartAt, reserveEndAt); err != nil {
			// FIXME: スケジューラ実装ができてきたら、ちゃんとエラーを返すように
			// tx.Rollback()
			// return echo.NewHTTPError(http.StatusConflict, "schedule conflict")
			c.Logger().Warn("schedule conflict")
		}
	}

	var (
		startAt    = time.Unix(req.StartAt, 0)
		endAt      = time.Unix(req.EndAt, 0)
		livestream = &Livestream{
			UserId:       userId,
			Title:        req.Title,
			Description:  req.Description,
			PlaylistUrl:  "https://d2jpkt808jogxx.cloudfront.net/BigBuckBunny/playlist.m3u8",
			ThumbnailUrl: "https://picsum.photos/200/300",
			StartAt:      startAt,
			EndAt:        endAt,
		}
	)
	rs, err := tx.NamedExecContext(ctx, "INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(:user_id, :title, :description, :playlist_url, :thumbnail_url, :start_at, :end_at)", livestream)
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	createdAt := time.Now()

	livestreamId, err := rs.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// タグ追加
	for _, tagId := range req.Tags {
		if _, err := tx.NamedExecContext(ctx, "INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (:livestream_id, :tag_id)", &LivestreamTag{
			LivestreamId: int(livestreamId),
			TagId:        tagId,
		}); err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	tx.Commit()

	livestream.Id = int(livestreamId)
	livestream.CreatedAt = createdAt
	livestream.UpdatedAt = createdAt
	return c.JSON(http.StatusCreated, livestream)
}

func getLivestreamsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	keyTagName := c.QueryParam("tag")

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// 複数件取得
	var livestreams []*Livestream
	if keyTagName == "" {
		if err := tx.SelectContext(ctx, &livestreams, "SELECT * FROM livestreams"); err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	} else {
		var tagIdList []int
		if err := dbConn.SelectContext(ctx, &tagIdList, "SELECT id FROM tags WHERE name = ?", keyTagName); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		query, params, err := sqlx.In("SELECT * FROM livestream_tags WHERE id IN (?)", tagIdList)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		var keyTaggedLivestreams []*LivestreamTag
		if err := tx.SelectContext(ctx, &keyTaggedLivestreams, query, params...); err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		for _, keyTaggedLivestream := range keyTaggedLivestreams {
			ls := Livestream{}
			if err := tx.GetContext(ctx, &ls, "SELECT * FROM livestreams WHERE id = ?", keyTaggedLivestream.LivestreamId); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}

			livestreams = append(livestreams, &ls)
		}
	}

	if err := tx.Commit(); err != nil {
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

	sess, err := session.Get(defaultSessionIdKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	userId, ok := sess.Values[defaultUserIdKey].(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to find user-id from session")
	}

	livestreamId, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	viewer := LivestreamViewer{
		UserId:       userId,
		LivestreamId: livestreamId,
	}

	if _, err := tx.NamedExecContext(ctx, "INSERT INTO livestream_viewers_history (user_id, livestream_id) VALUES(:user_id, :livestream_id)", viewer); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if _, err := tx.ExecContext(ctx, "UPDATE livestreams SET viewers_count = viewers_count + 1 WHERE id = ?", livestreamId); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, nil)
}

func leaveLivestreamHandler(c echo.Context) error {
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

	if _, err := tx.ExecContext(ctx, "UPDATE livestreams SET viewers_count = viewers_count - 1 WHERE id = ?", livestreamId); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, nil)
}

func getLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamId := c.Param("livestream_id")
	livestream := Livestream{}
	if err := dbConn.GetContext(ctx, &livestream, "SELECT * FROM livestreams WHERE id = ?", livestreamId); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, livestream)
}

func getLivecommentReportsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamId := c.Param("livestream_id")

	var reports []*LivecommentReport
	if err := dbConn.SelectContext(ctx, &reports, "SELECT * FROM livecomment_reports WHERE livestream_id = ?", livestreamId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, reports)
}
