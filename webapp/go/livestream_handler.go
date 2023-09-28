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
	ID            int       `db:"id"`
	UserID        int       `db:"user_id"`
	Title         string    `db:"title"`
	Description   string    `db:"description"`
	PrivacyStatus string    `db:"privacy_status"`
	StartAt       time.Time `db:"start_at"`
	EndAt         time.Time `db:"end_at"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type LivestreamViewer struct {
	UserID       int `db:"user_id"`
	LivestreamID int `db:"livestream_id"`
}

// FIXME: リアクション

func reserveLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	userID, ok := sess.Values[defaultUserIDKey].(int)
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
			UserID:        userID,
			Title:         req.Title,
			Description:   req.Description,
			PrivacyStatus: req.PrivacyStatus,
			StartAt:       startAt,
			EndAt:         endAt,
		}
	)
	rs, err := tx.NamedExecContext(ctx, "INSERT INTO livestreams (user_id, title, description, privacy_status, start_at, end_at) VALUES(:user_id, :title, :description, :privacy_status, :start_at, :end_at)", livestream)
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	createdAt := time.Now()

	livestreamID, err := rs.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx.Commit()

	livestream.ID = int(livestreamID)
	livestream.CreatedAt = createdAt
	livestream.UpdatedAt = createdAt
	return c.JSON(http.StatusCreated, livestream)
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

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, livestreams)
}

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

	userID, ok := sess.Values[defaultUserIDKey].(int)
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

	viewer := LivestreamViewer{
		UserID:       userID,
		LivestreamID: livestreamID,
	}

	if _, err := tx.NamedExecContext(ctx, "INSERT INTO livestream_viewers (user_id, livestream_id) VALUES(:user_id, :livestream_id)", viewer); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
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

	userID, ok := sess.Values[defaultUserIDKey].(int)
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

	viewer := LivestreamViewer{
		UserID:       userID,
		LivestreamID: livestreamID,
	}

	if _, err := tx.NamedExecContext(ctx, "DELETE FROM livestream_viewers WHERE user_id = :user_id AND livestream_id = :livestream_id", viewer); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func getLivestreamHandler(c echo.Context) error {
	return nil
}
