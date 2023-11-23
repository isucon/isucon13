package isupipe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type InitializeResponse struct {
	Language string `json:"language" validate:"required"`
}

func (c *Client) Initialize(ctx context.Context) (*InitializeResponse, error) {
	lgr := zap.S()

	req, err := c.agent.NewRequest(http.MethodPost, "/api/initialize", nil)
	if err != nil {
		lgr.Warnf("initializeのリクエスト初期化失敗: %s\n", err.Error())
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		c.contestantLogger.Warn("POST /api/initialize のリクエストが失敗しました", zap.Error(err))
		return nil, fmt.Errorf("initializeのリクエストに失敗しました %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("initialize へのリクエストに対して、期待されたHTTPステータスコードが確認できませんでした (expected:%d, actual:%d)", http.StatusOK, resp.StatusCode)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	var initializeResp *InitializeResponse
	if json.NewDecoder(resp.Body).Decode(&initializeResp); err != nil {
		return nil, fmt.Errorf("initializeのJSONのdecodeに失敗しました %v", err)
	}
	if err := ValidateResponse(req, initializeResp); err != nil {
		c.contestantLogger.Warn(err.Error())
		return nil, err
	}

	return initializeResp, nil
}
