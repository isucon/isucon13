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
