package main

import "sync"

// FIXME: 試験的にwebappに課金サーバを兼任させる
// とりあえずfinalcheck等を実装する上で必要なので用意
type Payment struct {
	ReservationId int
	Tip           int
}

type PaymentResult struct {
	Total    int
	Payments []*Payment
}

var (
	total     int
	payments  []*Payment
	paymentMu sync.RWMutex
)

func AddPayment(reservationId, tip int) {
	paymentMu.Lock()
	defer paymentMu.Unlock()

	payments = append(payments, &Payment{
		ReservationId: reservationId,
		Tip:           tip,
	})
	total += tip
}

func GetPaymentResult() *PaymentResult {
	paymentMu.RLock()
	defer paymentMu.RUnlock()

	return &PaymentResult{
		Total:    total,
		Payments: payments,
	}
}
