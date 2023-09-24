package isupipe

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
)

type Client struct {
	agent *agent.Agent
	// FIXME: dns resolver
}

func NewClient(customOpts ...agent.AgentOption) (*Client, error) {
	opts := []agent.AgentOption{
		agent.WithBaseURL("http://127.0.0.1:12345"),
		agent.WithCloneTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			// Custom DNS Resolver
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				nameserverAddress := "1.1.1.1"
				dialTimeout := 10000 * time.Millisecond
				dialer := net.Dialer{
					Timeout: dialTimeout,
					Resolver: &net.Resolver{
						PreferGo: true,
						Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
							dialer := net.Dialer{Timeout: dialTimeout}
							nameserver := net.JoinHostPort(nameserverAddress, "53")
							return dialer.DialContext(ctx, "udp", nameserver)
						},
					},
				}
				return dialer.DialContext(ctx, network, address)
			},
		}),
		agent.WithNoCache(),
		agent.WithTimeout(500 * time.Millisecond),
	}
	for _, customOpt := range customOpts {
		opts = append(opts, customOpt)
	}
	customAgent, err := agent.NewAgent(opts...)
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

	if _, err = c.sendRequest(ctx, req); err != nil {
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

	if _, err = c.sendRequest(ctx, req); err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	return nil
}

func (c *Client) GetUser(ctx context.Context, userID string) error {
	urlPath := fmt.Sprintf("/user/%s", userID)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return err
	}
	if _, err := c.sendRequest(ctx, req); err != nil {
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

	if _, err := c.sendRequest(ctx, req); err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	return nil
}

func (c *Client) PostReaction(ctx context.Context, livestreamId int, r *PostReactionRequest) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return bencherror.WrapError(bencherror.SystemError, err)
	}

	urlPath := fmt.Sprintf("/livestream/%d/reaction", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
	if err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	if _, err := c.sendRequest(ctx, req); err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	return nil
}

func (c *Client) PostSuperchat(ctx context.Context, livestreamId int, r *PostSuperchatRequest) (*PostSuperchatResponse, error) {
	payload, err := json.Marshal(r)
	if err != nil {
		log.Println("failed to marshal json")
		return nil, bencherror.WrapError(bencherror.SystemError, err)
	}

	urlPath := fmt.Sprintf("/livestream/%d/superchat", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
	if err != nil {
		log.Println("faild to initiate request")
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		log.Printf("error = %s\n", err.Error())
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}
	defer resp.Body.Close()

	var superchatResponse *PostSuperchatResponse
	if err := json.NewDecoder(resp.Body).Decode(&superchatResponse); err != nil {
		log.Println("failed to unmarshal json")
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	benchscore.AddScore(benchscore.SuccessPostSuperchat)

	return superchatResponse, nil
}

func (c *Client) ReportSuperchat(ctx context.Context, superchatId int) error {
	urlPath := fmt.Sprintf("/superchat/%d/report", superchatId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, nil)
	if err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	if _, err := c.sendRequest(ctx, req); err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	return nil
}

func (c *Client) GetLivestreamsByTag(
	ctx context.Context,
	tag string,
) error {
	urlPath := fmt.Sprintf("/search_livestream?tag=%s", tag)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	if _, err := c.sendRequest(ctx, req); err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	return nil
}

func (c *Client) GetTags(ctx context.Context) error {
	req, err := c.agent.NewRequest(http.MethodGet, "/tags", nil)
	if err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	if _, err := c.sendRequest(ctx, req); err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	return nil
}

func (c *Client) sendRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := c.agent.Do(ctx, req)
	if errors.Is(context.DeadlineExceeded, err) {
		return resp, bencherror.WrapError(bencherror.BenchmarkTimeoutError, err)
	} else if err != nil {
		return resp, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	return resp, nil
}
