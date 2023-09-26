package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Tag struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt time.Time `db:"created_at"`
}

func getTagHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tags := []Tag{}
	if err := dbConn.SelectContext(ctx, &tags, "SELECT id, name FROM tags"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, tags)
}

func searchLivestreamsByTagHandler(c echo.Context) error {
	ctx := c.Request().Context()

	keyTagName := c.QueryParam("tag")

	keyTag := Tag{}
	if err := dbConn.GetContext(ctx, &keyTag, "SELECT id FROM tags WHERE name = ?", keyTagName); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	livestreams := []Livestream{}
	if err := dbConn.SelectContext(ctx, &livestreams, "SELECT id FROM livestream_tags WHERE tag_id = ?", keyTag.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, livestreams)
}
