package isupipe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/bencherror"
)

type LivestreamStatistics struct {
	Rank           int64 `json:"rank" validate:"required"`
	ViewersCount   int64 `json:"viewers_count"`
	TotalReactions int64 `json:"total_reactions"`
	TotalReports   int64 `json:"total_reports"`
	MaxTip         int64 `json:"max_tip"`
}

type UserStatistics struct {
	Rank              int64 `json:"rank" validate:"required"`
	ViewersCount      int64 `json:"viewers_count"`
	TotalReactions    int64 `json:"total_reactions"`
	TotalLivecomments int64 `json:"total_livecomments"`
	TotalTip          int64 `json:"total_tip"`
	// NOTE: リアクション投稿がない場合、空文字になるのでvalidate対象外
	FavoriteEmoji string `json:"favorite_emoji"`
}

func (c *Client) GetUserStatistics(ctx context.Context, username string, opts ...ClientOption) (*UserStatistics, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/api/user/%s/statistics", username)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var stats *UserStatistics
	if resp.StatusCode == defaultStatusCode {
		if json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return nil, err
		}

		if err := ValidateResponse(req, stats); err != nil {
			return nil, err
		}
	}

	return stats, nil
}

func (c *Client) GetLivestreamStatistics(ctx context.Context, livestreamID int64, streamerName string, opts ...ClientOption) (*LivestreamStatistics, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	if err := c.setStreamerURL(streamerName); err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	urlPath := fmt.Sprintf("/api/livestream/%d/statistics", livestreamID)
	req, err := c.themeAgent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var stats *LivestreamStatistics
	if resp.StatusCode == defaultStatusCode {
		if json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return nil, err
		}

		if err := ValidateResponse(req, stats); err != nil {
			return nil, err
		}
	}

	return stats, nil
}
