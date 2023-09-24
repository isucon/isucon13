package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type Reaction struct {
	ID           string    `db:"id"`
	EmojiName    string    `db:"emoji_name"`
	UserID       string    `db:"user_id"`
	LivestreamID string    `db:"livestream_id"`
	CreatedAt    time.Time `db:"created_at"`
}

type PostReactionRequest struct {
	EmojiName string `json:"emoji_name"`
}

type PostReactionResponse struct {
	ReactionID int64 `json:"reaction_id"`
}

func postReactionHandler(c echo.Context) error {
	ctx := c.Request().Context()
	livestreamID := c.Param("livestream_id")

	if err := verifyUserSession(c); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	userID, ok := sess.Values[defaultUserIDKey].(int)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
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

	return c.JSON(http.StatusCreated, &PostReactionResponse{
		ReactionID: reactionID,
	})
}
