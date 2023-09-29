package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultSessionIDKey      = "SESSIONID"
	defaultSessionExpiresKey = "EXPIRES"
	defaultUserIDKey         = "USERID"
	bcryptDefaultCost        = 10
)

type User struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	DisplayName string `db:"display_name"`
	Description string `db:"description"`
	// HashedPassword is hashed password.
	HashedPassword string `db:"password"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Theme struct {
	UserID   int  `db:"user_id"`
	DarkMode bool `db:"dark_mode"`
}

type Session struct {
	// ID is an identifier that forms an UUIDv4.
	ID     string `db:"id"`
	UserID int    `db:"user_id"`
	// Expires is the UNIX timestamp that the sesison will be expired.
	Expires int `db:"expires"`
}

type PostUserRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	// Password is non-hashed password.
	Password string `json:"password"`
	Theme    Theme  `json:"theme"`
}

type LoginRequest struct {
	UserName string `json:"username"`
	// Password is non-hashed password.
	Password string `json:"password"`
}

// ユーザ登録API
// POST /user
func userRegisterHandler(c echo.Context) error {
	ctx := c.Request().Context()

	req := PostUserRequest{}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptDefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	user := User{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		HashedPassword: string(hashedPassword),
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	result, err := tx.NamedExecContext(ctx, "INSERT INTO users (name, display_name, description, password) VALUES(:name, :display_name, :description, :password)", user)
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	userID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	user.ID = int(userID)

	theme := Theme{
		UserID:   int(userID),
		DarkMode: req.Theme.DarkMode,
	}
	if _, err := tx.NamedExecContext(ctx, "INSERT INTO themes (user_id, dark_mode) VALUES(:user_id, :dark_mode)", theme); err != nil {
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
	ctx := c.Request().Context()

	req := LoginRequest{}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	user := User{}
	// usernameはUNIQUEなので、whereで一意に特定できる
	if err := dbConn.GetContext(ctx, &user, "SELECT * FROM users WHERE name = ?", req.UserName); err != nil {
		c.Logger().Printf("failed to get: username='%s', err=%+v", req.UserName, err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid username or password")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sessionEndAt := time.Now().Add(10 * time.Minute)

	sessionID := uuid.NewString()
	userSession := Session{
		ID:      sessionID,
		UserID:  user.ID,
		Expires: int(sessionEndAt.Unix()),
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sess.Options = &sessions.Options{
		MaxAge: int(60000 /* 10 seconds */), // FIXME: 600
		Path:   "/",
	}
	sess.Values[defaultSessionIDKey] = userSession.ID
	c.Logger().Infof("userSession.ID = %s", userSession.ID)
	sess.Values[defaultUserIDKey] = userSession.UserID
	c.Logger().Infof("userSession.UserID = %d", userSession.UserID)
	sess.Values[defaultSessionExpiresKey] = int(sessionEndAt.Unix())
	c.Logger().Infof("sessionEndAt = %s", sessionEndAt.String())

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

// ユーザ詳細API
// GET /user/:userid
func userHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	userID := c.Param("user_id")
	user := User{}
	if err := dbConn.GetContext(ctx, &user, "SELECT name, display_name, description, created_at, updated_at FROM users WHERE id = ?", userID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, user)
}

// 配信者のテーマ取得API
// GET /user/:userid/theme
func getUserThemeHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		c.Logger().Printf("verifyUserSession: %+v\n", err)
		return err
	}

	userID := c.Param("user_id")
	theme := Theme{}
	if err := dbConn.GetContext(ctx, &theme, "SELECT dark_mode FROM themes WHERE user_id = ?", userID); err != nil {
		c.Logger().Printf("dbConn.Get: %+v\n", err)
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, theme)
}

// ユーザ
// XXX セッション情報返すみたいな？
// GET /user
func userSessionHandler(c echo.Context) error {
	return nil
}

func verifyUserSession(c echo.Context) error {
	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sessionExpires, ok := sess.Values[defaultSessionExpiresKey]
	if !ok {
		// FIXME: エラーメッセージを検討する
		return echo.NewHTTPError(http.StatusForbidden, "")
	}

	now := time.Now()
	if now.Unix() > int64(sessionExpires.(int)) {
		return echo.NewHTTPError(http.StatusUnauthorized, "session has expired")
	}

	return nil
}
