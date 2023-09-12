package isupipe

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
)

type Client struct {
	agent *agent.Agent
	// FIXME: dns resolver
}

func NewClient() (*Client, error) {
	baseURL := "http://127.0.0.1:12345"
	timeout := 500 * time.Millisecond
	customAgent, err := agent.NewAgent(
		agent.WithBaseURL(baseURL),
		agent.WithCloneTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}),
		agent.WithNoCache(),
		agent.WithTimeout(timeout),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		agent: customAgent,
	}, nil
}

func (c *Client) PostUser(ctx context.Context, r *PostUserRequest) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return bencherror.WrapError(bencherror.SystemError, err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/user", bytes.NewReader(payload))
	if err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	if _, err = c.agent.Do(ctx, req); err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	return nil
}

func (c *Client) Login(ctx context.Context, r *LoginRequest) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return bencherror.WrapError(bencherror.SystemError, err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/login", bytes.NewReader(payload))
	if err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	if _, err = c.agent.Do(ctx, req); err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	return nil
}

func (c *Client) ReserveLivestream(ctx context.Context, r *ReserveLivestreamRequest) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return bencherror.WrapError(bencherror.SystemError, err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/livestream/reservation", bytes.NewReader(payload))
	if err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	if _, err := c.agent.Do(ctx, req); err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	return nil
}

func (c *Client) PostSuperchat(ctx context.Context, livestreamId int, r *PostSuperchatRequest) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return bencherror.WrapError(bencherror.SystemError, err)
	}

	urlPath := fmt.Sprintf("/livestream/%d/superchat", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
	if err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	if _, err := c.agent.Do(ctx, req); err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	return nil
}
