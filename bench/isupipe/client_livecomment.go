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

type Livecomment struct {
	ID         int64      `json:"id"`
	User       User       `json:"user"`
	Livestream Livestream `json:"livestream"`
	Comment    string     `json:"comment"`
	Tip        int        `json:"tip"`
	CreatedAt  int        `json:"created_at"`
}

type LivecommentReport struct {
	ID          int64       `json:"id"`
	Reporter    User        `json:"reporter"`
	Livecomment Livecomment `json:"livecomment"`
	CreatedAt   int64       `json:"created_at"`
}

type (
	PostLivecommentRequest struct {
		Comment string `json:"comment"`
		Tip     int64  `json:"tip"`
	}
	PostLivecommentResponse struct {
		ID         int64      `json:"id"`
		User       User       `json:"user"`
		Livestream Livestream `json:"livestream"`
		Comment    string     `json:"comment"`
		Tip        int64      `json:"tip"`
		CreatedAt  int64      `json:"created_at"`
	}
)

type ModerateRequest struct {
	NGWord string `json:"ng_word"`
}

type NGWord struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	LivestreamID int64  `json:"livestream_id"`
	Word         string `json:"word"`
	CreatedAt    int64  `json:"created_at"`
}

func (c *Client) GetLivecomments(ctx context.Context, livestreamID int64, opts ...ClientOption) ([]Livecomment, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/api/livestream/%d/livecomment", livestreamID)
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

	livecomments := []Livecomment{}
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&livecomments); err != nil {
			return livecomments, bencherror.NewHttpResponseError(err, req)
		}
	}

	benchscore.AddScore(benchscore.SuccessGetLivecomments)
	return livecomments, nil
}

func (c *Client) GetLivecommentReports(ctx context.Context, livestreamID int64, opts ...ClientOption) ([]LivecommentReport, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/api/livestream/%d/report", livestreamID)
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

	reports := []LivecommentReport{}
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
			return reports, bencherror.NewHttpResponseError(err, req)
		}
	}

	benchscore.AddScore(benchscore.SuccessGetLivecommentReports)
	return reports, nil
}

// FIXME: 実装
func (c *Client) GetNgwords(ctx context.Context, livestreamID int64, opts ...ClientOption) ([]*NGWord, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/api/livestream/%d/ngwords", livestreamID)
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

	var ngwords []*NGWord
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&ngwords); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}
	}

	return ngwords, nil
}

func (c *Client) PostLivecomment(ctx context.Context, livestreamID int64, r *PostLivecommentRequest, opts ...ClientOption) (*PostLivecommentResponse, error) {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	urlPath := fmt.Sprintf("/api/livestream/%d/livecomment", livestreamID)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

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

	var livecommentResponse *PostLivecommentResponse
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&livecommentResponse); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}

		benchscore.AddTipProfit(livecommentResponse.Tip)
	}

	benchscore.AddScore(benchscore.SuccessPostLivecomment)

	return livecommentResponse, nil
}

func (c *Client) ReportLivecomment(ctx context.Context, livestreamID, livecommentID int64, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/api/livestream/%d/livecomment/%d/report", livestreamID, livecommentID)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

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

	benchscore.AddScore(benchscore.SuccessReportLivecomment)
	return nil
}

func (c *Client) Moderate(ctx context.Context, livestreamID int64, ngWord string, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/api/livestream/%d/moderate", livestreamID)
	payload, err := json.Marshal(&ModerateRequest{
		NGWord: ngWord,
	})
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, urlPath, bytes.NewBuffer(payload))
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

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

	return nil
}
