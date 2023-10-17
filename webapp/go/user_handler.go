package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultSessionIdKey      = "SESSIONID"
	defaultSessionExpiresKey = "EXPIRES"
	defaultUserIdKey         = "USERID"
	bcryptDefaultCost        = 10
)

type UserModel struct {
	Id          int    `db:"id"`
	Name        string `db:"name"`
	DisplayName string `db:"display_name"`
	Description string `db:"description"`
	// HashedPassword is hashed password.
	HashedPassword string `db:"password"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type User struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt int `json:"created_at"`
	UpdatedAt int `json:"updated_at"`

	IsPopular bool  `json:"is_popular"`
	Theme     Theme `json:"theme"`
}

type Theme struct {
	Id        int  `json:"id"`
	UserId    int  `json:"user_id"`
	DarkMode  bool `json:"dark_mode"`
	CreatedAt int  `json:"created_at"`
}

type ThemeModel struct {
	Id        int       `db:"id"`
	UserId    int       `db:"user_id"`
	DarkMode  bool      `db:"dark_mode"`
	CreatedAt time.Time `db:"created_at"`
}

type PostUserRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	// Password is non-hashed password.
	Password string               `json:"password"`
	Theme    PostUserRequestTheme `json:"theme"`
}

type PostUserRequestTheme struct {
	DarkMode bool `json:"dark_mode"`
}

type LoginRequest struct {
	UserName string `json:"username"`
	// Password is non-hashed password.
	Password string `json:"password"`
}

func getUserSessionHandler(c echo.Context) error {
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

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	userModel := UserModel{}
	err = tx.GetContext(ctx, &userModel, "SELECT * FROM users WHERE id = ?", userId)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	user, err := fillUserResponse(ctx, tx, userModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

// ユーザ登録API
// POST /user
func postUserHandler(c echo.Context) error {
	ctx := c.Request().Context()

	req := PostUserRequest{}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	if req.Name == "pipe" {
		return echo.NewHTTPError(http.StatusBadRequest, "the username 'pipe' is reserved")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptDefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "begin tx failed")
	}
	defer tx.Rollback()

	now := time.Now()
	userModel := UserModel{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		HashedPassword: string(hashedPassword),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result, err := tx.NamedExecContext(ctx, "INSERT INTO users (name, display_name, description, password, created_at, updated_at) VALUES(:name, :display_name, :description, :password, :created_at, :updated_at)", userModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "user insertion failed")
	}

	userId, err := result.LastInsertId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "last insert id failed")
	}

	userModel.Id = int(userId)

	themeModel := ThemeModel{
		UserId:    int(userId),
		DarkMode:  req.Theme.DarkMode,
		CreatedAt: now,
	}
	if _, err := tx.NamedExecContext(ctx, "INSERT INTO themes (user_id, dark_mode, created_at) VALUES(:user_id, :dark_mode, :created_at)", themeModel); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "theme insertion failed")
	}

	user, err := fillUserResponse(ctx, tx, userModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fill user response")
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "commit failed")
	}

	if disablePowerDNS {
		c.Logger().Info("disbale dns")
		return c.JSON(http.StatusCreated, user)
	}

	if err := exec.Command("pdnsutil", "add-record", "u.isucon.dev", req.Name, "a", "30", powerDNSSubdomainAddress).Run(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add A record to PowerDNS")
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

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	userModel := UserModel{}
	// usernameはUNIQUEなので、whereで一意に特定できる
	err = tx.GetContext(ctx, &userModel, "SELECT * FROM users WHERE name = ?", req.UserName)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid username or password")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(userModel.HashedPassword), []byte(req.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid username or password")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sessionEndAt := time.Now().Add(10 * time.Minute)

	sessionId := uuid.NewString()

	sess, err := session.Get(defaultSessionIdKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sess.Options = &sessions.Options{
		MaxAge: int(60000 /* 10 seconds */), // FIXME: 600
		Path:   "/",
	}
	sess.Values[defaultSessionIdKey] = sessionId
	sess.Values[defaultUserIdKey] = userModel.Id
	sess.Values[defaultSessionExpiresKey] = int(sessionEndAt.Unix())

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

// ユーザ詳細API
// GET /user/:userid
func getUserHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	username := c.Param("username")

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	userModel := UserModel{}
	if err := tx.GetContext(ctx, &userModel, "SELECT * FROM users WHERE name = ?", username); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	user, err := fillUserResponse(ctx, tx, userModel)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

func getUsersHandler(c echo.Context) (err error) {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	var userModels []*UserModel
	if err := tx.SelectContext(ctx, &userModels, "SELECT id, name, display_name, description, created_at, updated_at FROM users"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	users := make([]User, len(userModels))
	for i := range userModels {
		user, err := fillUserResponse(ctx, tx, *userModels[i])
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		users[i] = user
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, users)
}

func verifyUserSession(c echo.Context) error {
	sess, err := session.Get(defaultSessionIdKey, c)
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

func userIsPopular(ctx context.Context, tx *sqlx.Tx, userId int) (bool, error) {
	var livestreamModels []*LivestreamModel
	if err := tx.SelectContext(ctx, &livestreamModels, "SELECT * FROM livestreams WHERE user_id = ?", userId); err != nil {
		return false, err
	}

	totalSpamReports := 0
	totalTips := 0
	totalLivecomments := 0
	for _, ls := range livestreamModels {
		spamReports := 0
		if err := tx.GetContext(ctx, &spamReports, "SELECT COUNT(*) FROM livecomment_reports WHERE livestream_id = ? ", ls.Id); err != nil {
			return false, err
		}

		var livecommentModels []*LivecommentModel
		if err := tx.SelectContext(ctx, &livecommentModels, "SELECT * FROM livecomments WHERE livestream_id = ?", ls.Id); err != nil {
			return false, err
		}

		for _, lc := range livecommentModels {
			totalTips += lc.Tip
		}

		totalSpamReports += spamReports
		totalLivecomments += len(livecommentModels)
	}

	if totalSpamReports >= 10 {
		return false, nil
	}

	if totalTips < 1000 {
		return false, nil
	}

	if totalLivecomments < 50 {
		return false, nil
	}

	return true, nil
}

func fillUserResponse(ctx context.Context, tx *sqlx.Tx, userModel UserModel) (User, error) {
	themeModel := ThemeModel{}
	if err := tx.GetContext(ctx, &themeModel, "SELECT * FROM themes WHERE user_id = ?", userModel.Id); err != nil {
		return User{}, err
	}

	popular, err := userIsPopular(ctx, tx, userModel.Id)
	if err != nil {
		return User{}, err
	}

	user := User{
		Id:          userModel.Id,
		Name:        userModel.Name,
		DisplayName: userModel.DisplayName,
		Description: userModel.Description,
		CreatedAt:   int(userModel.CreatedAt.Unix()),
		UpdatedAt:   int(userModel.UpdatedAt.Unix()),
		IsPopular:   popular,
		Theme: Theme{
			Id:        themeModel.Id,
			UserId:    themeModel.UserId,
			DarkMode:  themeModel.DarkMode,
			CreatedAt: int(themeModel.CreatedAt.Unix()),
		},
	}

	return user, nil
}
