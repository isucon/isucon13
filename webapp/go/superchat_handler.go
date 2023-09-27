package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type PostSuperchatRequest struct {
	Comment string `json:"comment"`
	Tip     int    `json:"tip"`
}

type Superchat struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	LivestreamID int       `json:"livestream_id" db:"livestream_id"`
	Comment      string    `json:"comment" db:"comment"`
	Tip          int       `json:"tip" db:"tip"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type SuperchatReport struct {
	ID          int       `db:"id"`
	UserID      int       `db:"user_id"`
	SuperchatID int       `db:"superchat_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func getSuperchatsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	livestreamID := c.Param("livestream_id")

	superchats := []Superchat{}
	if err := dbConn.SelectContext(ctx, &superchats, "SELECT * FROM superchats WHERE livestream_id = ?", livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, superchats)
}

func postSuperchatHandler(c echo.Context) error {
	ctx := c.Request().Context()

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	userID, ok := sess.Values[defaultUserIDKey].(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to find user-id from session")
	}

	var req *PostSuperchatRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	superchat := Superchat{
		UserID:       userID,
		LivestreamID: livestreamID,
		Comment:      req.Comment,
		Tip:          req.Tip,
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO superchats (user_id, livestream_id, comment, tip) VALUES (:user_id, :livestream_id, :comment, :tip)", superchat)
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	superchatID, err := rs.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	superchat.ID = int(superchatID)
	createdAt := time.Now()
	superchat.CreatedAt = createdAt
	superchat.UpdatedAt = createdAt
	return c.JSON(http.StatusCreated, superchat)
}

func reportSuperchatHandler(c echo.Context) error {
	ctx := c.Request().Context()

	superchatID, err := strconv.Atoi(c.Param("superchat_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	userID, ok := sess.Values[defaultUserIDKey].(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO superchat_reports(user_id, superchat_id) VALUES (:user_id, :superchat_id)", &SuperchatReport{
		UserID:      userID,
		SuperchatID: superchatID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reportID, err := rs.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx.Commit()

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"report_id": reportID,
	})
}
