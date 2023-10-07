package main

// ISUCON的な参考: https://github.com/isucon/isucon12-qualify/blob/main/webapp/go/isuports.go#L336
// sqlx的な参考: https://jmoiron.github.io/sqlx/

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
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
	listenPort                     = 12345
	powerDNSHostEnvKey             = "ISUCON13_POWERDNS_HOST"
	powerDNSAPIKeyEnvKey           = "ISUCON13_POWERDNS_APIKEY"
	powerDNSSubdomainAddressEnvKey = "ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS"
	// FIXME: ISUCON当日までに削除する
	powerDNSDisableEnvKey = "ISUCON13_POWERDNS_DISABLED"
)

var (
	disablePowerDNS          bool   = false
	powerDNSHost             string = "localhost"
	powerDNSAPIKey           string
	powerDNSSubdomainAddress string
	dbConn                   *sqlx.DB
	secret                   = []byte("defaultsecret")
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	if secretKey, ok := os.LookupEnv("ISUCON13_SESSION_SECRETKEY"); ok {
		secret = []byte(secretKey)
	}
}

// FIXME: ポータルと足並み揃えて修正
type InitializeResponse struct {
	AdvertiseLevel int    `json:"advertise_level"`
	Language       string `json:"language"`
}

func loadDBDialConfigFromOSEnv() (*mysql.Config, error) {
	conf := mysql.NewConfig()
	return conf, nil
}

func connectDB() (*sqlx.DB, error) {
	const (
		networkTypeEnvKey = "ISUCON13_MYSQL_DIALCONFIG_NET"
		addrEnvKey        = "ISUCON13_MYSQL_DIALCONFIG_ADDRESS"
		portEnvKey        = "ISUCON13_MYSQL_DIALCONFIG_PORT"
		userEnvKey        = "ISUCON13_MYSQL_DIALCONFIG_USER"
		passwordEnvKey    = "ISUCON13_MYSQL_DIALCONFIG_PASSWORD"
		dbNameEnvKey      = "ISUCON13_MYSQL_DIALCONFIG_DATABASE"
		parseTimeEnvKey   = "ISUCON13_MYSQL_DIALCONFIG_PARSETIME"
	)

	conf := mysql.NewConfig()

	// 環境変数がセットされていなかった場合でも一旦動かせるように、デフォルト値を入れておく
	// この挙動を変更して、エラーを出すようにしてもいいかもしれない
	conf.Net = "tcp"
	conf.Addr = net.JoinHostPort("127.0.0.1", "3306")
	conf.User = "isucon"
	conf.Passwd = "isucon"
	conf.DBName = "isupipe"
	conf.ParseTime = true

	if v, ok := os.LookupEnv(networkTypeEnvKey); ok {
		conf.Net = v
	}
	if addr, ok := os.LookupEnv(addrEnvKey); ok {
		if port, ok2 := os.LookupEnv(portEnvKey); ok2 {
			conf.Addr = net.JoinHostPort(addr, port)
		} else {
			conf.Addr = net.JoinHostPort(addr, "3306")
		}
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
	if out, err := exec.Command("./init.sh").CombinedOutput(); err != nil {
		c.Logger().Warnf("init.sh failed with err=%s", string(out))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.Request().Header.Add("Content-Type", "application/json;chatset=utf-8")
	return c.JSON(http.StatusOK, InitializeResponse{
		AdvertiseLevel: 5,
		Language:       "golang",
	})
}

func main() {
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(echolog.DEBUG)
	e.Use(middleware.Logger())
	cookieStore := sessions.NewCookieStore(secret)
	// cookieStore.Options.Domain = "*.u.isucon.dev"
	e.Use(session.Middleware(cookieStore))
	// e.Use(middleware.Recover())

	// 初期化
	e.POST("/initialize", initializeHandler)

	// top
	e.GET("/tag", getTagHandler)
	e.GET("/theme", getStreamerThemeHandler)

	// livestream
	// reserve livestream
	e.POST("/livestream/reservation", reserveLivestreamHandler)
	// list livestream
	e.GET("/livestream", getLivestreamsHandler)
	// get livestream
	e.GET("/livestream/:livestream_id", getLivestreamHandler)
	// get polling livecomment timeline
	e.GET("/livestream/:livestream_id/livecomment", getLivecommentsHandler)
	// ライブコメント投稿
	e.POST("/livestream/:livestream_id/livecomment", postLivecommentHandler)
	e.POST("/livestream/:livestream_id/reaction", postReactionHandler)
	e.GET("/livestream/:livestream_id/reaction", getReactionsHandler)

	// (配信者向け)ライブコメントの報告一覧取得API
	e.GET("/livestream/:livestream_id/report", getLivecommentReportsHandler)
	// ライブコメント報告
	e.POST("/livestream/:livestream_id/livecomment/:livecomment_id/report", reportLivecommentHandler)
	// 配信者によるモデレーション (NGワード登録)
	e.POST("/livestream/:livestream_id/moderate", moderateNGWordHandler)

	// livestream_viewersにINSERTするため必要
	// ユーザ視聴開始 (viewer)
	e.POST("/livestream/:livestream_id/enter", enterLivestreamHandler)
	// ユーザ視聴終了 (viewer)
	e.DELETE("/livestream/:livestream_id/enter", leaveLivestreamHandler)

	// user
	e.POST("/user", postUserHandler)
	e.POST("/login", loginHandler)
	e.GET("/user", getUsersHandler)
	e.GET("/user/me", getUserSessionHandler)
	// FIXME: ユーザ一覧を返すAPI
	// フロントエンドで、配信予約のコラボレーターを指定する際に必要
	e.GET("/user/:user_id", getUserHandler)
	e.GET("/user/:user_id/statistics", getUserStatisticsHandler)

	// stats
	// ライブコメント統計情報
	e.GET("/livestream/:livestream_id/statistics", getLivestreamStatisticsHandler)

	// 課金情報
	e.GET("/payment", GetPaymentResult)

	// DB接続
	conn, err := connectDB()
	if err != nil {
		e.Logger.Fatalf("failed to connect db: %v", err)
		return
	}
	conn.SetMaxOpenConns(10)
	defer conn.Close()
	dbConn = conn

	host, ok := os.LookupEnv(powerDNSHostEnvKey)
	if ok {
		powerDNSHost = host
	}
	key, ok := os.LookupEnv(powerDNSAPIKeyEnvKey)
	if !ok {
		e.Logger.Fatalf("environ %s must be provided", powerDNSAPIKeyEnvKey)
	}
	powerDNSAPIKey = key
	subdomainAddr, ok := os.LookupEnv(powerDNSSubdomainAddressEnvKey)
	if !ok {
		e.Logger.Fatalf("environ %s must be provided", powerDNSSubdomainAddressEnvKey)
	}
	powerDNSSubdomainAddress = subdomainAddr

	disabledEnv, _ := os.LookupEnv(powerDNSDisableEnvKey)
	disabled, err := strconv.ParseBool(disabledEnv)
	if err != nil {
		disablePowerDNS = false
	}
	disablePowerDNS = disabled

	// HTTPサーバ起動
	listenAddr := net.JoinHostPort("", strconv.Itoa(listenPort))
	if err := e.Start(listenAddr); err != nil {
		e.Logger.Fatal(err)
	}
}
