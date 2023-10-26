package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type Tag struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt int64 `json:"created_at"`
}

type TagModel struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt int64 `db:"created_at"`
}

type TagsResponse struct {
	Tags []*Tag `json:"tags"`
}

func getTagHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	var tagModels []*TagModel
	if err := tx.SelectContext(ctx, &tagModels, "SELECT * FROM tags"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit")
	}

	tags := make([]*Tag, len(tagModels))
	for i := range tagModels {
		tags[i] = &Tag{
			Id:        tagModels[i].Id,
			Name:      tagModels[i].Name,
			CreatedAt: tagModels[i].CreatedAt,
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

	hostHeader := c.Request().Host

	username := strings.Split(hostHeader, ".")[0]

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	userModel := UserModel{}
	err = tx.GetContext(ctx, &userModel, "SELECT id FROM users WHERE name = ?", username)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	themeModel := ThemeModel{}
	if err := tx.GetContext(ctx, &themeModel, "SELECT * FROM themes WHERE user_id = ?", userModel.Id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit")
	}

	theme := Theme{
		Id:        themeModel.Id,
		DarkMode:  themeModel.DarkMode,
		CreatedAt: themeModel.CreatedAt,
	}

	return c.JSON(http.StatusOK, theme)
}
