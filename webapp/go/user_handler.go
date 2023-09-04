package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"crypto/sha512"

	"github.com/labstack/echo/v4"
)

type User struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	DisplayName string `db:"display_name"`
	Description string `db:"description"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt int `db:"created_at"`
	UpdatedAt int `db:"updated_at"`
}

type PasswordHash struct {
	ID       int    `db:"id"`
	UserID   int    `db:"user_id"`
	Password string `db:"password"`
}

type PostUserRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Password    string `json:"password"`
}

// ユーザ登録API
// POST /user
func userRegisterHandler(c echo.Context) error {
	ctx := c.Request().Context()

	req := PostUserRequest{}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	hashedPassword := sha512.Sum512([]byte(req.Password))
	user := User{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	result, err := tx.NamedExecContext(ctx, "INSERT INTO users (name, display_name, description) VALUES(:name, :display_name, :description)", user)
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	userID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	passwordHash := PasswordHash{
		UserID:   int(userID),
		Password: fmt.Sprintf("%x", hashedPassword),
	}
	if _, err := tx.NamedExecContext(ctx, "INSERT INTO password_hash (user_id, password) VALUES(:user_id, :password)", passwordHash); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, user)
}

// ユーザログインAPI
// POST /login
func loginHandler(c echo.Context) error {
	return nil
}

// ユーザ詳細API
// GET /user/:userid
func userHandler(c echo.Context) error {
	return nil
}

// ユーザ
// XXX セッション情報返すみたいな？
// GET /user
func userSessionHandler(c echo.Context) error {
	return nil
}

// ユーザが登録しているチャンネル一覧
// GET /user/:user_id/channel
func userChannelHandler(c echo.Context) error {
	return nil
}

// チャンネル登録
// POST /user/:user_id/channel/:channelid/subscribe
func subscribeChannelHandler(c echo.Context) error {
	return nil
}

// チャンネル登録解除
// POST /user/:user_id/channel/:channelid/unsubscribe
func unsubscribeChannelHandler(c echo.Context) error {
	return nil
}

// チャンネル情報
// GET /channel/:channel_id
func channelHandler(c echo.Context) error {
	return nil
}

// チャンネル登録者数
// GET /channel/:channel_id/subscribers
func channelSubscribersHandler(c echo.Context) error {
	return nil
}

// チャンネルの動画一覧
// GET /channel/:channel_id/movie
func channelMovieHandler(c echo.Context) error {
	return nil
}

// チャンネル作成
// POST /channel
func createChannelHandler(c echo.Context) error {
	return nil
}

// チャンネル編集
// PUT /channel/:channel_id
func updateChannelHandler(c echo.Context) error {
	return nil
}

// チャンネル削除
// DELETE /channel/:channel_id
func deleteChannelHandler(c echo.Context) error {
	return nil
}
