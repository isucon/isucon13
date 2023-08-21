package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Livestream struct {
	Id        int       `db:"id"`
	OwnerId   int       `db:"owner_id"`
	ChannelId int       `db:"channel_id"`
	Title     string    `db:"title"`
	StartAt   time.Time `db:"start_at"`
	EndAt     time.Time `db:"end_at"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func getLivestreamsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// 追加
	now := time.Now()
	ls := &Livestream{
		OwnerId:   1,
		ChannelId: 1,
		Title:     "test",
		StartAt:   now,
		EndAt:     now,
	}
	if _, err := tx.NamedExecContext(ctx, "INSERT INTO livestreams (owner_id, channel_id, title, start_at, end_at) VALUES(:owner_id, :channel_id, :title, :start_at, :end_at)", ls); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// 更新
	ls.Title = "test2"
	if _, err := tx.NamedExecContext(ctx, "UPDATE livestreams SET title = 'test2' WHERE id = :id", ls); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// １件取得
	var livestream Livestream
	if err := tx.GetContext(ctx, &livestream, "SELECT * FROM livestreams WHERE id = ?", 22); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusUnauthorized, "livestream not found")
		}
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	c.Logger().Debugf("livestream = %+v\n", livestream)

	// 複数件取得
	var livestreams []*Livestream
	if err := tx.SelectContext(ctx, &livestreams, "SELECT * FROM livestreams"); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	c.Logger().Debugf("livestreams = %+v\n", livestreams)

	// 削除
	if _, err := tx.NamedExecContext(ctx, "DELETE FROM livestreams WHERE id = :id", livestreams[0]); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, livestreams)
}

func getLivestreamHandler(c echo.Context) error {
	return nil
}

func getLivestreamCommentsHandler(c echo.Context) error {
	return nil
}
