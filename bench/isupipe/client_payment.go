package isupipe

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type Payment struct {
	ReservationId int
	Tip           int
}

type PaymentResult struct {
	Total    int
	Payments []*Payment
}

func (c *Client) GetPaymentResult(ctx context.Context) (*PaymentResult, error) {
	req, err := c.agent.NewRequest(http.MethodGet, "/payment", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	var paymentResp *PaymentResult
	if json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, err
	}

	return paymentResp, nil
}
