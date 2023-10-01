package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type Tag struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt time.Time `db:"created_at"`
}

type TagsResponse struct {
	Tags []*Tag `json:"tags"`
}

func getTagHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var tags []*Tag
	if err := dbConn.SelectContext(ctx, &tags, "SELECT * FROM tags"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, &TagsResponse{
		Tags: tags,
	})
}

// 配信者のテーマ取得API
// GET /theme
func getStreamerThemeHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		c.Logger().Printf("verifyUserSession: %+v\n", err)
		return err
	}

	hostHeaders := c.Request().Header["Host"]
	if len(hostHeaders) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Host header must be specified")
	}

	host := hostHeaders[0]

	username := strings.Split(host, ".")[0]

	user := User{}
	if err := dbConn.GetContext(ctx, &user, "SELECT id FROM users WHERE name = ?", username); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	theme := Theme{}
	if err := dbConn.GetContext(ctx, &theme, "SELECT dark_mode FROM themes WHERE user_id = ?", user.ID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, theme)
}
