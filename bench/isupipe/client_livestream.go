package isupipe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/isucon/isucon13/bench/internal/bencherror"
)

type Livestream struct {
	ID           int64  `json:"id" validate:"required"`
	Owner        User   `json:"owner" validate:"required"`
	Tags         []Tag  `json:"tags" validate:"required,dive,required"`
	Title        string `json:"title" validate:"required"`
	Description  string `json:"description" validate:"required"`
	PlaylistUrl  string `json:"playlist_url" validate:"required"`
	ThumbnailUrl string `json:"thumbnail_url" validate:"required"`
	StartAt      int64  `json:"start_at" validate:"required"`
	EndAt        int64  `json:"end_at" validate:"required"`
}

func (l *Livestream) Hours() int {
	diffSec := time.Unix(l.EndAt, 0).Sub(time.Unix(l.StartAt, 0))
	return int(diffSec / time.Hour)
}

type (
	ReserveLivestreamRequest struct {
		Tags         []int64 `json:"tags"`
		Title        string  `json:"title"`
		Description  string  `json:"description"`
		PlaylistUrl  string  `json:"playlist_url"`
		ThumbnailUrl string  `json:"thumbnail_url"`
		StartAt      int64   `json:"start_at"`
		EndAt        int64   `json:"end_at"`
	}
)

func (c *Client) GetLivestream(
	ctx context.Context,
	livestreamID int64,
	streamerName string,
	opts ...ClientOption,
) (*Livestream, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	if err := c.setStreamerURL(streamerName); err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	urlPath := fmt.Sprintf("/api/livestream/%d", livestreamID)
	req, err := c.themeAgent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

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
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&livestream); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}

		if err := ValidateResponse(req, livestream); err != nil {
			return nil, err
		}
	}

	return livestream, nil
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
	if o.searchTag != nil {
		query := req.URL.Query()
		query.Add("tag", o.searchTag.Tag)
		req.URL.RawQuery = query.Encode()
	}

	if o.limitParam != nil {
		query := req.URL.Query()
		query.Add("limit", strconv.Itoa(o.limitParam.Limit))
		req.URL.RawQuery = query.Encode()
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

		if err := ValidateSlice(req, livestreams); err != nil {
			return nil, err
		}
	}

	return livestreams, nil
}

// 自分のライブ配信一覧取得
func (c *Client) GetMyLivestreams(ctx context.Context, opts ...ClientOption) ([]*Livestream, error) {
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

		if err := ValidateSlice(req, livestreams); err != nil {
			return nil, err
		}
	}

	return livestreams, nil
}

// 特定ユーザのライブ配信取得
func (c *Client) GetUserLivestreams(ctx context.Context, username string, opts ...ClientOption) ([]*Livestream, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, fmt.Sprintf("/api/user/%s/livestream", username), nil)
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

		if err := ValidateSlice(req, livestreams); err != nil {
			return nil, err
		}
	}

	return livestreams, nil
}

func (c *Client) ReserveLivestream(ctx context.Context, streamerName string, r *ReserveLivestreamRequest, opts ...ClientOption) (*Livestream, error) {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	if err := c.setStreamerURL(streamerName); err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	req, err := c.themeAgent.NewRequest(http.MethodPost, "/api/livestream/reservation", bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

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
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&livestream); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}

		if err := ValidateResponse(req, livestream); err != nil {
			return nil, err
		}
	}

	return livestream, nil
}

func (c *Client) EnterLivestream(ctx context.Context, livestreamID int64, streamerName string, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	if err := c.setStreamerURL(streamerName); err != nil {
		return bencherror.NewInternalError(err)
	}
	urlPath := fmt.Sprintf("/api/livestream/%d/enter", livestreamID)
	req, err := c.themeAgent.NewRequest(http.MethodPost, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

	resp, err := sendRequest(ctx, c.themeAgent, req)
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

	return nil
}

func (c *Client) ExitLivestream(ctx context.Context, livestreamID int64, streamerName string, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	if err := c.setStreamerURL(streamerName); err != nil {
		return bencherror.NewInternalError(err)
	}
	urlPath := fmt.Sprintf("/api/livestream/%d/exit", livestreamID)
	req, err := c.themeAgent.NewRequest(http.MethodDelete, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.themeAgent, req)
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

	return nil
}
