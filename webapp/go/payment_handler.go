package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// webappに課金サーバを兼任させる
// とりあえずfinalcheck等を実装する上で必要なので用意
type Payment struct {
	ReservationID int64 `json:"reservation_id"`
	Tip           int64 `json:"tip"`
}

type PaymentResult struct {
	Total    int64      `json:"total"`
	Payments []*Payment `json:"payments"`
}

// FIXME: 予約一覧を返す(payments)
func GetPaymentResult(c echo.Context) error {
	ctx := c.Request().Context()

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	defer tx.Rollback()

	var total int64
	if err := tx.GetContext(ctx, &total, "SELECT SUM(tip) FROM livecomments"); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusOK, &PaymentResult{})
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, &PaymentResult{
		Total: total,
	})
}
