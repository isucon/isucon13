package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type Tag struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt int `json:"created_at"`
}

type TagModel struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt time.Time `db:"created_at"`
}

type TagsResponse struct {
	Tags []*Tag `json:"tags"`
}

func getTagHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var tagModels []*TagModel
	if err := dbConn.SelectContext(ctx, &tagModels, "SELECT * FROM tags"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tags := make([]*Tag, len(tagModels))
	for i := range tagModels {
		tags[i] = &Tag{
			Id:        tagModels[i].Id,
			Name:      tagModels[i].Name,
			CreatedAt: int(tagModels[i].CreatedAt.Unix()),
		}
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

	userModel := UserModel{}
	err := dbConn.GetContext(ctx, &userModel, "SELECT id FROM users WHERE name = ?", username)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	theme := Theme{}
	if err := dbConn.GetContext(ctx, &theme, "SELECT dark_mode FROM themes WHERE user_id = ?", userModel.Id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, theme)
}
