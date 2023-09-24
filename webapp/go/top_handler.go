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

	rows, err := dbConn.QueryxContext(ctx, "SELECT id, name FROM tags")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	for rows.Next() {
		tag := Tag{}
		if err := rows.Scan(&tag); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		tags = append(tags, tag)
	}

	return c.JSON(http.StatusOK, tags)
}
