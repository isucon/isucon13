package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type ReserveLivestreamRequest struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	PrivacyStatus string `json:"privacy_status"`
	StartAt       int64  `json:"start_at"`
	EndAt         int64  `json:"end_at"`
}

type Livestream struct {
	Id            int       `db:"id"`
	UserId        int       `db:"user_id"`
	Title         string    `db:"title"`
	Description   string    `db:"description"`
	PrivacyStatus string    `db:"privacy_status"`
	StartAt       time.Time `db:"start_at"`
	EndAt         time.Time `db:"end_at"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type PostSuperchatRequest struct {
	Comment string `json:"comment"`
	Tip     int    `json:"tip"`
}

type PostSuperchatResponse struct {
	SuperchatId int64  `json:"superchat_id"`
	Comment     string `json:"comment"`
	Tip         int    `json:"tip"`
}

type Superchat struct {
	Id           int       `db:"id"`
	UserId       int       `db:"user_id"`
	LivestreamId int       `db:"livestream_id"`
	Comment      string    `db:"comment"`
	Tip          int       `db:"tip"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// FIXME: リアクション
// FIXME: isucon13-mysql-1  | ERROR 1064 (42000) at line 55: You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near ') ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_bin' at line 17

func reserveLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	userId, ok := sess.Values[defaultUserIDKey].(int)
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

	// FIXME: 2024/04/01 - 2025/03/31までの期間かチェック

	var (
		startAt    = time.Unix(req.StartAt, 0)
		endAt      = time.Unix(req.EndAt, 0)
		livestream = &Livestream{
			UserId:        userId,
			Title:         req.Title,
			Description:   req.Description,
			PrivacyStatus: req.PrivacyStatus,
			StartAt:       startAt,
			EndAt:         endAt,
		}
	)
	if _, err := tx.NamedExecContext(ctx, "INSERT INTO livestreams (user_id, title, description, privacy_status, start_at, end_at) VALUES(:user_id, :title, :description, :privacy_status, :start_at, :end_at)", livestream); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx.Commit()

	// FIXME: PK補完
	return c.JSON(http.StatusCreated, livestream)
}

func postSuperchatHandler(c echo.Context) error {
	ctx := c.Request().Context()

	livestreamId, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	userId, ok := sess.Values[defaultUserIDKey].(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	var req *PostSuperchatRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var (
		superchat = &Superchat{
			UserId:       userId,
			LivestreamId: livestreamId,
			Comment:      req.Comment,
			Tip:          req.Tip,
		}
	)
	rs, err := tx.NamedExecContext(ctx, "INSERT INTO superchats (user_id, livestream_id, comment, tip) VALUES (:user_id, :livestream_id, :comment, :tip)", superchat)
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	superchatId, err := rs.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx.Commit()

	return c.JSON(http.StatusCreated, &PostSuperchatResponse{
		SuperchatId: superchatId,
		Comment:     superchat.Comment,
		Tip:         superchat.Tip,
	})
}

func getLivestreamsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// 複数件取得
	var livestreams []*Livestream
	if err := tx.SelectContext(ctx, &livestreams, "SELECT * FROM livestreams"); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	c.Logger().Debugf("livestreams = %+v\n", livestreams)

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, livestreams)
}

func getLivestreamHandler(c echo.Context) error {
	return nil
}

func getLivestreamCommentsHandler(c echo.Context) error {
	return nil
}
