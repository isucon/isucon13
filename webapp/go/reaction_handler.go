package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type Reaction struct {
	ID           int       `json:"id" db:"id"`
	EmojiName    string    `json:"emoji_name" db:"emoji_name"`
	UserID       string    `json:"user_id" db:"user_id"`
	LivestreamID string    `json:"livestream_id" db:"livestream_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type PostReactionRequest struct {
	EmojiName string `json:"emoji_name"`
}

func getReactionsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	livestreamID := c.Param("livestream_id")

	reactions := []Reaction{}
	if err := dbConn.SelectContext(ctx, &reactions, "SELECT * FROM reactions WHERE livestream_id = ?", livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, reactions)
}

func postReactionHandler(c echo.Context) error {
	ctx := c.Request().Context()
	livestreamID := c.Param("livestream_id")

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	log.Printf("%+v\n", sess.Values)
	userID, ok := sess.Values[defaultUserIDKey].(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to find user-id from session")
	}

	var req *PostReactionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reaction := Reaction{
		UserID:       fmt.Sprintf("%d", userID),
		LivestreamID: livestreamID,
		EmojiName:    req.EmojiName,
	}

	result, err := tx.NamedExecContext(ctx, "INSERT INTO reactions (user_id, livestream_id, emoji_name) VALUES (:user_id, :livestream_id, :emoji_name)", reaction)
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reactionID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reaction.ID = int(reactionID)
	return c.JSON(http.StatusCreated, reaction)
}
