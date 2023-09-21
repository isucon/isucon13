package main

// ISUCON的な参考: https://github.com/isucon/isucon12-qualify/blob/main/webapp/go/isuports.go#L336
// sqlx的な参考: https://jmoiron.github.io/sqlx/

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	echolog "github.com/labstack/gommon/log"
)

const (
	listenPort = 12345
)

var (
	dbConn *sqlx.DB
	secret = []byte("defaultsecret")
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	if secretKey, ok := os.LookupEnv("ISUCON13_SESSION_SECRETKEY"); ok {
		secret = []byte(secretKey)
	}
}

// FIXME: ポータルと足並み揃えて修正
type InitializeResponse struct {
	AvailableDays int    `json:"available_days"`
	Language      string `json:"language"`
}

func loadDBDialConfigFromOSEnv() (*mysql.Config, error) {
	conf := mysql.NewConfig()
	return conf, nil
}

func connectDB() (*sqlx.DB, error) {
	const (
		networkTypeEnvKey = "ISUCON13_MYSQL_DIALCONFIG_NET"
		addrEnvKey        = "ISUCON13_MYSQL_DIALCONFIG_ADDRESS"
		userEnvKey        = "ISUCON13_MYSQL_DIALCONFIG_USER"
		passwordEnvKey    = "ISUCON13_MYSQL_DIALCONFIG_PASSWORD"
		dbNameEnvKey      = "ISUCON13_MYSQL_DIALCONFIG_DATABASE"
		parseTimeEnvKey   = "ISUCON13_MYSQL_DIALCONFIG_PARSETIME"
	)

	conf := mysql.NewConfig()

	// 環境変数がセットされていなかった場合でも一旦動かせるように、デフォルト値を入れておく
	// この挙動を変更して、エラーを出すようにしてもいいかもしれない
	conf.Net = "tcp"
	conf.Addr = "127.0.0.1:3306"
	conf.User = "isucon"
	conf.Passwd = "isucon"
	conf.DBName = "isupipe"
	conf.ParseTime = true

	if v, ok := os.LookupEnv(networkTypeEnvKey); ok {
		conf.Net = v
	}
	if v, ok := os.LookupEnv(addrEnvKey); ok {
		conf.Addr = v
	}
	if v, ok := os.LookupEnv(userEnvKey); ok {
		conf.User = v
	}
	if v, ok := os.LookupEnv(passwordEnvKey); ok {
		conf.Passwd = v
	}
	if v, ok := os.LookupEnv(dbNameEnvKey); ok {
		conf.DBName = v
	}
	if v, ok := os.LookupEnv(parseTimeEnvKey); ok {
		parseTime, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse environment variable '%s' as bool: %+v", parseTimeEnvKey, err)
		}
		conf.ParseTime = parseTime
	}

	return sqlx.Open("mysql", conf.FormatDSN())
}

func initializeHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := dbConn.ExecContext(ctx, "DELETE FROM superchats"); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if _, err := dbConn.ExecContext(ctx, "DELETE FROM livestreams"); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if _, err := dbConn.ExecContext(ctx, "DELETE FROM users"); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tx.Commit()

	c.Request().Header.Add("Content-Type", "application/json;chatset=utf-8")
	return c.JSON(http.StatusOK, InitializeResponse{
		AvailableDays: 0,
		Language:      "golang",
	})
}

func main() {
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(echolog.DEBUG)
	e.Use(middleware.Logger())
	e.Use(session.Middleware(sessions.NewCookieStore(secret)))
	// e.Use(middleware.Recover())

	// 初期化
	e.POST("/initialize", initializeHandler)

	// top
	e.GET("/tag", nil)

	// livestream
	// reserve livestream
	e.POST("/livestream/reservation", reserveLivestreamHandler)
	// list livestream
	e.GET("/livestream", getLivestreamsHandler)
	// get livestream
	e.GET("/livestream/:livestream_id", getLivestreamHandler)
	e.POST("/livestream/:livestream_id/superchat", postSuperchatHandler)
	// get polling superchat timeline
	e.GET("/livestream/:livestream_id/superchat", nil)
	// スパチャ投稿
	// スパチャ報告

	// get reaction 候補 (FIXME: フロントエンドで持つなら要らなそう)
	e.GET("/reaction", nil)

	// ユーザ視聴開始 (viewer)
	// ユーザ視聴終了 (viewer)

	// user
	e.POST("/user", userRegisterHandler)
	e.POST("/login", loginHandler)
	e.GET("/user", userSessionHandler)
	e.GET("/user/:user_id", userHandler)

	// DB接続
	conn, err := connectDB()
	if err != nil {
		e.Logger.Fatalf("failed to connect db: %v", err)
		return
	}
	conn.SetMaxOpenConns(10)
	defer conn.Close()
	dbConn = conn

	// HTTPサーバ起動
	listenAddr := net.JoinHostPort("", strconv.Itoa(listenPort))
	if err := e.Start(listenAddr); err != nil {
		e.Logger.Fatal(err)
	}
}
