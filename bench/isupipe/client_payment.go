package isupipe

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type PaymentResult struct {
	// NOTE: 売上0を許容
	TotalTip int64 `json:"total_tip"`
}

func (c *Client) GetPaymentResult(ctx context.Context) (*PaymentResult, error) {
	req, err := c.agent.NewRequest(http.MethodGet, "/api/payment", nil)
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

	if err := ValidateResponse(req, paymentResp); err != nil {
		return nil, err
	}

	return paymentResp, nil
}
