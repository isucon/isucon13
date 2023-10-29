package isupipe

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type InitializeResponse struct {
	AdvertiseLevel int    `json:"advertise_level"`
	Language       string `json:"language"`
}

func (c *Client) Initialize(ctx context.Context) (*InitializeResponse, error) {
	req, err := c.agent.NewRequest(http.MethodPost, "/initialize", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	var initializeResp *InitializeResponse
	if json.NewDecoder(resp.Body).Decode(&initializeResp); err != nil {
		return nil, err
	}

	return initializeResp, nil
}
