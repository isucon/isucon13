package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultSessionIdKey      = "SESSIONId"
	defaultSessionExpiresKey = "EXPIRES"
	defaultUserIdKey         = "USERId"
	bcryptDefaultCost        = 10
)

type User struct {
	Id          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	DisplayName string `json:"display_name" db:"display_name"`
	Description string `json:"description" db:"description"`
	// HashedPassword is hashed password.
	HashedPassword string `json:"hashed_password" db:"password"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	IsPopular bool `json:"is_popular"`
}

type Theme struct {
	UserId   int  `json:"user_id" db:"user_id"`
	DarkMode bool `json:"dark_mode" db:"dark_mode"`
}

type Session struct {
	// Id is an identifier that forms an UUIdv4.
	Id     string `json:"id" db:"id"`
	UserId int    `json:"user_id" db:"user_id"`
	// Expires is the UNIX timestamp that the sesison will be expired.
	Expires int `json:"expires" db:"expires"`
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

	user := User{}
	if err := dbConn.GetContext(ctx, &user, "SELECT name, display_name, description, created_at, updated_at FROM users WHERE id = ?", userId); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	popular, err := userIsPopular(ctx, userId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	user.IsPopular = popular

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

	user := User{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		HashedPassword: string(hashedPassword),
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "begin tx failed")
	}

	result, err := tx.NamedExecContext(ctx, "INSERT INTO users (name, display_name, description, password) VALUES(:name, :display_name, :description, :password)", user)
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "user insertion failed")
	}

	userId, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "last insert id failed")
	}

	user.Id = int(userId)

	theme := Theme{
		UserId:   int(userId),
		DarkMode: req.Theme.DarkMode,
	}
	if _, err := tx.NamedExecContext(ctx, "INSERT INTO themes (user_id, dark_mode) VALUES(:user_id, :dark_mode)", theme); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "theme insertion failed")
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "commit failed")
	}

	if disablePowerDNS {
		c.Logger().Info("disbale dns")
		return c.JSON(http.StatusCreated, user)
	}

	if err := exec.Command("pdnsutil", "add-record", "u.isucon.dev", req.Name, "30").Run(); err != nil {
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
		// return echo.NewHTTPError(http.StatusUnauthorized, "invalid username or password")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sessionEndAt := time.Now().Add(10 * time.Minute)

	sessionId := uuid.NewString()
	userSession := Session{
		Id:      sessionId,
		UserId:  user.Id,
		Expires: int(sessionEndAt.Unix()),
	}

	sess, err := session.Get(defaultSessionIdKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sess.Options = &sessions.Options{
		MaxAge: int(60000 /* 10 seconds */), // FIXME: 600
		Path:   "/",
	}
	sess.Values[defaultSessionIdKey] = userSession.Id
	c.Logger().Infof("userSession.Id = %s", userSession.Id)
	sess.Values[defaultUserIdKey] = userSession.UserId
	c.Logger().Infof("userSession.UserId = %d", userSession.UserId)
	sess.Values[defaultSessionExpiresKey] = int(sessionEndAt.Unix())
	c.Logger().Infof("sessionEndAt = %s", sessionEndAt.String())

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

	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	user := User{}
	if err := dbConn.GetContext(ctx, &user, "SELECT name, display_name, description, created_at, updated_at FROM users WHERE id = ?", userId); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	popular, err := userIsPopular(ctx, userId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	user.IsPopular = popular

	return c.JSON(http.StatusOK, user)
}

func getUsersHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	var users []*User
	if err := dbConn.SelectContext(ctx, &users, "SELECT id, name, display_name, description, created_at, updated_at FROM users"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// FIXME: IsFamousのアルゴリズムを作る
	for i := range users {
		userIsPopular(ctx, users[i].Id)
		users[i].IsPopular = true
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

func userIsPopular(ctx context.Context, userId int) (bool, error) {
	var livestreams []*Livestream
	if err := dbConn.SelectContext(ctx, &livestreams, "SELECT * FROM livestreams WHERE user_id = ?", userId); err != nil {
		return false, err
	}

	totalSpamReports := 0
	totalTips := 0
	totalLivecomments := 0
	for _, ls := range livestreams {
		spamReports := 0
		if err := dbConn.SelectContext(ctx, &spamReports, "SELECT COUNT(*) FROM livecomment_reports WHERE livestream_id = ? ", ls.Id); err != nil {
			return false, err
		}

		var livecomments []*Livecomment
		if err := dbConn.SelectContext(ctx, &livecomments, "SELECT * FROM livecomments WHERE livestream_id = ?", ls.Id); err != nil {
			return false, err
		}

		for _, lc := range livecomments {
			totalTips += lc.Tip
		}

		totalSpamReports += spamReports
		totalLivecomments += len(livecomments)
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
