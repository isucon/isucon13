package main

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

// webappに課金サーバを兼任させる
// とりあえずfinalcheck等を実装する上で必要なので用意
type Payment struct {
	ReservationId int64 `json:"reservation_id"`
	Tip           int64 `json:"tip"`
}

type PaymentResult struct {
	Total    int64      `json:"total"`
	Payments []*Payment `json:"payments"`
}

var (
	total     int64
	payments  []*Payment
	paymentMu sync.RWMutex
)

func AddPayment(reservationId, tip int64) {
	paymentMu.Lock()
	defer paymentMu.Unlock()

	payments = append(payments, &Payment{
		ReservationId: reservationId,
		Tip:           tip,
	})
	total += tip
}

func GetPaymentResult(c echo.Context) error {
	paymentMu.RLock()
	defer paymentMu.RUnlock()

	return c.JSON(http.StatusOK, &PaymentResult{
		Total:    total,
		Payments: payments,
	})
}
