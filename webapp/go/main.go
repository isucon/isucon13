package main

// ISUCON的な参考: https://github.com/isucon/isucon12-qualify/blob/main/webapp/go/isuports.go#L336
// sqlx的な参考: https://jmoiron.github.io/sqlx/

import (
	"log"
	"net"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echolog "github.com/labstack/gommon/log"
)

const (
	listenPort = 12345
)

var (
	dbConn *sqlx.DB
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func connectDB() (*sqlx.DB, error) {
	// FIXME: envから読み出す
	conf := mysql.NewConfig()
	conf.Net = "tcp"
	conf.Addr = "127.0.0.1:3306"
	conf.User = "isucon"
	conf.Passwd = "isucon"
	conf.DBName = "isupipe"
	conf.ParseTime = true
	return sqlx.Open("mysql", conf.FormatDSN())
}

func main() {
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(echolog.DEBUG)
	e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	// top
	e.GET("/tag", nil)

	// livestream
	e.GET("/livestream", getLivestreamsHandler)
	e.GET("/livestream/:livestream_id", getLivestreamHandler)
	e.GET("/livestream/:livestream_id/comment", nil)
	e.GET("/livestream/user/:userid", nil)
	e.GET("/reaction", nil)
	// movie
	e.GET("/movie", nil)

	// user
	e.POST("/user", nil)
	e.POST("/login", nil)
	e.GET("/user/:user_id", nil)
	e.GET("/user/:user_id/channel", nil)
	e.POST("/user/:user_id/channel/:channel_id/subscribe", nil)
	e.POST("/user/:user_id/channel/:channel_id/unsubscribe", nil)
	e.GET("/channel", nil)
	e.GET("/channel/:channel_id/subscribers", nil)
	e.GET("/channel/:channel_id/movie", nil)
	e.POST("/channel", nil)
	e.PUT("/channel/:channel_id", nil)
	e.DELETE("/channel/:channel_id", nil)

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
