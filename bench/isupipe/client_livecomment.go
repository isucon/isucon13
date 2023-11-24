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
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/scheduler"
)

type Livecomment struct {
	ID         int64      `json:"id" validate:"required"`
	User       User       `json:"user" validate:"required"`
	Livestream Livestream `json:"livestream" validate:"required"`
	Comment    string     `json:"comment" validate:"required"`
	// NOTE: Tipがない場合が許容される(tip=0)
	Tip       int `json:"tip"`
	CreatedAt int `json:"created_at" validate:"required"`
}

type LivecommentReport struct {
	ID          int64       `json:"id" validate:"required"`
	Reporter    User        `json:"reporter" validate:"required"`
	Livecomment Livecomment `json:"livecomment" validate:"required"`
	CreatedAt   int64       `json:"created_at" validate:"required"`
}

type (
	PostLivecommentRequest struct {
		Comment string `json:"comment"`
		Tip     int64  `json:"tip"`
	}
	PostLivecommentResponse struct {
		ID         int64      `json:"id" validate:"required"`
		User       User       `json:"user" validate:"required"`
		Livestream Livestream `json:"livestream" validate:"required"`
		Comment    string     `json:"comment" validate:"required"`
		Tip        int64      `json:"tip"`
		CreatedAt  int64      `json:"created_at" validate:"required"`
	}
)

type (
	ModerateRequest struct {
		NGWord string `json:"ng_word"`
	}

	ModerateResponse struct {
		WordID int64 `json:"word_id" validate:"required"`
	}
)

type NGWord struct {
	ID           int64  `json:"id" validate:"required"`
	UserID       int64  `json:"user_id" validate:"required"`
	LivestreamID int64  `json:"livestream_id" validate:"required"`
	Word         string `json:"word" validate:"required"`
	CreatedAt    int64  `json:"created_at" validate:"required"`
}

func (c *Client) GetLivecomments(ctx context.Context, livestreamID int64, streamerName string, opts ...ClientOption) ([]*Livecomment, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	if err := c.setStreamerURL(streamerName); err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	urlPath := fmt.Sprintf("/api/livestream/%d/livecomment", livestreamID)
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

	livecomments := []*Livecomment{}
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&livecomments); err != nil {
			return livecomments, bencherror.NewHttpResponseError(err, req)
		}

		if err := ValidateSlice(req, livecomments); err != nil {
			return nil, err
		}
	}

	return livecomments, nil
}

func (c *Client) GetLivecommentReports(ctx context.Context, livestreamID int64, streamerName string, opts ...ClientOption) ([]LivecommentReport, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	if err := c.setStreamerURL(streamerName); err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	urlPath := fmt.Sprintf("/api/livestream/%d/report", livestreamID)
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

	reports := []LivecommentReport{}
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
			return reports, bencherror.NewHttpResponseError(err, req)
		}

		if err := ValidateSlice(req, reports); err != nil {
			return nil, err
		}
	}

	return reports, nil
}

func (c *Client) GetNgwords(ctx context.Context, livestreamID int64, streamerName string, opts ...ClientOption) ([]*NGWord, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	if err := c.setStreamerURL(streamerName); err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	urlPath := fmt.Sprintf("/api/livestream/%d/ngwords", livestreamID)
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

	var ngwords []*NGWord
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&ngwords); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}

		if err := ValidateSlice(req, ngwords); err != nil {
			return nil, err
		}
	}

	return ngwords, nil
}

func (c *Client) PostLivecomment(ctx context.Context, livestreamID int64, streamerName string, comment string, tip *scheduler.Tip, opts ...ClientOption) (*PostLivecommentResponse, int, error) {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
		r                 = &PostLivecommentRequest{
			Comment: comment,
			Tip:     int64(tip.Tip),
		}
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, 0, bencherror.NewInternalError(err)
	}

	if err := c.setStreamerURL(streamerName); err != nil {
		return nil, 0, bencherror.NewInternalError(err)
	}
	urlPath := fmt.Sprintf("/api/livestream/%d/livecomment", livestreamID)
	req, err := c.themeAgent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

	resp, err := sendRequest(ctx, c.themeAgent, req)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != o.wantStatusCode {
		return nil, 0, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var livecommentResponse *PostLivecommentResponse
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&livecommentResponse); err != nil {
			return nil, 0, bencherror.NewHttpResponseError(err, req)
		}

		if err := ValidateResponse(req, livecommentResponse); err != nil {
			return nil, 0, err
		}

		benchscore.AddTip(uint64(tip.Tip))
	}

	return livecommentResponse, tip.Tip, nil
}

func (c *Client) ReportLivecomment(ctx context.Context, livestreamID int64, streamerName string, livecommentID int64, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	if err := c.setStreamerURL(streamerName); err != nil {
		return bencherror.NewInternalError(err)
	}
	urlPath := fmt.Sprintf("/api/livestream/%d/livecomment/%d/report", livestreamID, livecommentID)
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

	var livecommentReport *LivecommentReport
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&livecommentReport); err != nil {
			return bencherror.NewHttpResponseError(err, req)
		}

		if o.validateReportLivecomment {
			if err := ValidateResponse(req, livecommentReport); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) Moderate(ctx context.Context, livestreamID int64, streamerName string, ngWord string, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	if err := c.setStreamerURL(streamerName); err != nil {
		return bencherror.NewInternalError(err)
	}

	urlPath := fmt.Sprintf("/api/livestream/%d/moderate", livestreamID)
	payload, err := json.Marshal(&ModerateRequest{
		NGWord: ngWord,
	})
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	req, err := c.themeAgent.NewRequest(http.MethodPost, urlPath, bytes.NewBuffer(payload))
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

	var moderateResp *ModerateResponse
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&moderateResp); err != nil {
			return bencherror.NewHttpResponseError(err, req)
		}

		if err := ValidateResponse(req, moderateResp); err != nil {
			return err
		}
	}

	return nil
}
