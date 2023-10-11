package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type ReactionModel struct {
	Id           int       `db:"id"`
	EmojiName    string    `db:"emoji_name"`
	UserId       int       `db:"user_id"`
	LivestreamId int       `db:"livestream_id"`
	CreatedAt    time.Time `db:"created_at"`
}

type Reaction struct {
	Id           int    `json:"id"`
	EmojiName    string `json:"emoji_name"`
	UserId       int    `json:"user_id"`
	LivestreamId int    `json:"livestream_id"`
	CreatedAt    int    `json:"created_at"`
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

	livestreamId := c.Param("livestream_id")

	reactionModels := []ReactionModel{}
	if err := dbConn.SelectContext(ctx, &reactionModels, "SELECT * FROM reactions WHERE livestream_id = ?", livestreamId); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	reactions := make([]Reaction, len(reactionModels))
	for i := range reactionModels {
		reactions[i] = Reaction{
			Id:           reactionModels[i].Id,
			EmojiName:    reactionModels[i].EmojiName,
			UserId:       reactionModels[i].UserId,
			LivestreamId: reactionModels[i].LivestreamId,
			CreatedAt:    int(reactionModels[i].CreatedAt.Unix()),
		}
	}
	return c.JSON(http.StatusOK, reactions)
}

func postReactionHandler(c echo.Context) error {
	ctx := c.Request().Context()
	livestreamId, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

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

	var req *PostReactionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reactionModel := ReactionModel{
		UserId:       userId,
		LivestreamId: livestreamId,
		EmojiName:    req.EmojiName,
	}

	result, err := tx.NamedExecContext(ctx, "INSERT INTO reactions (user_id, livestream_id, emoji_name) VALUES (:user_id, :livestream_id, :emoji_name)", reactionModel)
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reactionId, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reaction := Reaction{
		Id:           int(reactionId),
		EmojiName:    reactionModel.EmojiName,
		UserId:       reactionModel.UserId,
		LivestreamId: reactionModel.LivestreamId,
		CreatedAt:    int(reactionModel.CreatedAt.Unix()),
	}
	return c.JSON(http.StatusCreated, reaction)
}
