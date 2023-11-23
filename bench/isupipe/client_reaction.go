package isupipe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/isucon/isucon13/bench/internal/bencherror"
)

type PostReactionRequest struct {
	EmojiName string `json:"emoji_name"`
}

type Reaction struct {
	ID         int64      `json:"id" validate:"required"`
	EmojiName  string     `json:"emoji_name" validate:"required"`
	User       User       `json:"user" validate:"required"`
	Livestream Livestream `json:"livestream" validate:"required"`
	CreatedAt  int64      `json:"created_at" validate:"required"`
}

func (c *Client) GetReactions(ctx context.Context, livestreamID int64, streamerName string, opts ...ClientOption) ([]Reaction, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	if err := c.setStreamerURL(streamerName); err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	urlPath := fmt.Sprintf("/api/livestream/%d/reaction", livestreamID)
	req, err := c.themeAgent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	if o.limitParam != nil {
		query := req.URL.Query()
		query.Add("limit", strconv.Itoa(o.limitParam.Limit))
		req.URL.RawQuery = query.Encode()
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

	reactions := []Reaction{}
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&reactions); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}

		if err := ValidateSlice(req, reactions); err != nil {
			return nil, err
		}
	}

	return reactions, nil
}

func (c *Client) PostReaction(ctx context.Context, livestreamID int64, streamerName string, r *PostReactionRequest, opts ...ClientOption) (*Reaction, error) {
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
	urlPath := fmt.Sprintf("/api/livestream/%d/reaction", livestreamID)
	req, err := c.themeAgent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
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

	reaction := &Reaction{}
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&reaction); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}

		if err := ValidateResponse(req, reaction); err != nil {
			return nil, err
		}
	}

	return reaction, nil
}
