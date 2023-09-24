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
	Id           int       `db:"id"`
	UserId       int       `db:"user_id"`
	LivestreamId int       `db:"livestream_id"`
	Comment      string    `db:"comment"`
	Tip          int       `db:"tip"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type SuperchatReport struct {
	Id          int       `db:"id"`
	UserId      int       `db:"user_id"`
	SuperchatId int       `db:"superchat_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
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
		return echo.NewHTTPError(http.StatusUnauthorized)
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

	createdAt := time.Now()
	return c.JSON(http.StatusCreated, &Superchat{
		Id:           int(superchatId),
		UserId:       userId,
		LivestreamId: livestreamId,
		Comment:      superchat.Comment,
		Tip:          superchat.Tip,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	})
}

func reportSuperchatHandler(c echo.Context) error {
	ctx := c.Request().Context()

	superchatId, err := strconv.Atoi(c.Param("superchat_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	userId, ok := sess.Values[defaultUserIDKey].(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO superchat_reports(user_id, superchat_id) VALUES (:user_id, :superchat_id)", &SuperchatReport{
		UserId:      userId,
		SuperchatId: superchatId,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reportId, err := rs.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx.Commit()

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"report_id": reportId,
	})
}
