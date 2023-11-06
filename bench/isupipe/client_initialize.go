package isupipe

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type InitializeResponse struct {
	AdvertiseLevel int64  `json:"advertise_level"`
	Language       string `json:"language"`
}

func (c *Client) Initialize(ctx context.Context) (*InitializeResponse, error) {
	req, err := c.agent.NewRequest(http.MethodPost, "/api/initialize", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

	log.Println("request initialize")
	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()
	log.Println("response initialize")

	var initializeResp *InitializeResponse
	if json.NewDecoder(resp.Body).Decode(&initializeResp); err != nil {
		return nil, err
	}
	log.Println("decode initialize")

	return initializeResp, nil
}
