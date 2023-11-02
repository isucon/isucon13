package isupipe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
)

type Livestream struct {
	ID           int    `json:"id"`
	Owner        User   `json:"owner"`
	Tags         []Tag  `json:"tags"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	PlaylistUrl  string `json:"playlist_url"`
	ThumbnailUrl string `json:"thumbnail_url"`
	StartAt      int    `json:"start_at"`
	EndAt        int    `json:"end_at"`
	CreatedAt    int    `json:"created_at"`
}

type (
	ReserveLivestreamRequest struct {
		Tags         []int  `json:"tags"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		PlaylistUrl  string `json:"playlist_url"`
		ThumbnailUrl string `json:"thumbnail_url"`
		StartAt      int64  `json:"start_at"`
		EndAt        int64  `json:"end_at"`
	}
)

func (c *Client) GetLivestream(
	ctx context.Context,
	livestreamID int,
	opts ...ClientOption,
) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/api/livestream/%d", livestreamID)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessGetLivestream)
	return nil
}

func (c *Client) SearchLivestreams(
	ctx context.Context,
	opts ...ClientOption,
) ([]*Livestream, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/api/livestream/search", nil)
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

	var livestreams []*Livestream
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&livestreams); err != nil {
			return nil, err
		}
	}

	return livestreams, nil
}

// 自分のライブ配信一覧取得
func (c *Client) GetLivestreams(ctx context.Context, opts ...ClientOption) ([]*Livestream, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/api/livestream", nil)
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

	var livestreams []*Livestream
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&livestreams); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}
	}

	return livestreams, nil
}

func (c *Client) SearchLivestreamsByTag(
	ctx context.Context,
	tag string,
	opts ...ClientOption,
) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/api/livestream/search", nil)
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
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

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

	req, err := c.themeAgent.NewRequest(http.MethodPost, "/api/livestream/reservation", bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := sendRequest(ctx, c.themeAgent, req)
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

	var livestream *Livestream
	if err := json.NewDecoder(resp.Body).Decode(&livestream); err != nil {
		return nil, bencherror.NewHttpResponseError(err, req)
	}

	benchscore.AddScore(benchscore.SuccessReserveLivestream)
	return livestream, nil
}

func (c *Client) EnterLivestream(ctx context.Context, livestreamID int, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/api/livestream/%d/enter", livestreamID)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessEnterLivestream)
	return nil
}

func (c *Client) LeaveLivestream(ctx context.Context, livestreamID int, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/api/livestream/%d/enter", livestreamID)
	req, err := c.agent.NewRequest(http.MethodDelete, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessLeaveLivestream)
	return nil
}
