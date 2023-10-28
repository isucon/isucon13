package isupipe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
)

type Livestream struct {
	Id           int    `json:"id"`
	Owner        User   `json:"owner"`
	Tags         []Tag  `json:"tags"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	PlaylistUrl  string `json:"playlist_url"`
	ThumbnailUrl string `json:"thumbnail_url"`
	ViewersCount int    `json:"viewers_count"`
	StartAt      int    `json:"start_at"`
	EndAt        int    `json:"end_at"`
	CreatedAt    int    `json:"created_at"`
	UpdatedAt    int    `json:"updated_at"`
}

type (
	ReserveLivestreamRequest struct {
		Tags        []int  `json:"tags"`
		Title       string `json:"title"`
		Description string `json:"description"`
		StartAt     int64  `json:"start_at"`
		EndAt       int64  `json:"end_at"`
	}
)

type Theme struct {
	DarkMode bool `json:"dark_mode"`
}

func (c *Client) GetTheme(ctx context.Context, streamer *User, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	// FIXME: 配信者のユーザ名を含めてリクエスト
	req, err := c.agent.NewRequest(http.MethodGet, "/theme", nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessGetUserTheme)
	return nil
}

func (c *Client) GetLivestream(
	ctx context.Context,
	livestreamId int,
	opts ...ClientOption,
) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d", livestreamId)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessGetLivestream)
	return nil
}

func (c *Client) GetLivestreams(
	ctx context.Context,
	opts ...ClientOption,
) ([]*Livestream, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/livestream", nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var livestreams []*Livestream
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&livestreams); err != nil {
			return nil, err
		}
	}

	return livestreams, nil
}

func (c *Client) GetLivestreamsByTag(
	ctx context.Context,
	tag string,
	opts ...ClientOption,
) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/livestream", nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	query := req.URL.Query()
	query.Add("tag", tag)
	req.URL.RawQuery = query.Encode()

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessGetLivestreamByTag)
	return nil
}

func (c *Client) ReserveLivestream(ctx context.Context, r *ReserveLivestreamRequest, opts ...ClientOption) (*Livestream, error) {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	req, err := c.themeAgent.NewRequest(http.MethodPost, "/livestream/reservation", bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := sendRequest(ctx, c.themeAgent, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var livestream *Livestream
	if err := json.NewDecoder(resp.Body).Decode(&livestream); err != nil {
		return nil, bencherror.NewHttpResponseError(err, req)
	}

	benchscore.AddScore(benchscore.SuccessReserveLivestream)
	return livestream, nil
}

func (c *Client) EnterLivestream(ctx context.Context, livestreamId int, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d/enter", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessEnterLivestream)
	return nil
}

func (c *Client) LeaveLivestream(ctx context.Context, livestreamId int, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d/enter", livestreamId)
	req, err := c.agent.NewRequest(http.MethodDelete, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessLeaveLivestream)
	return nil
}
