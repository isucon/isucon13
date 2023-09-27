package isupipe

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

const DefaultClientBaseURL = "http://127.0.0.1:12345"

func NewClient(customOpts ...agent.AgentOption) (*Client, error) {
	opts := []agent.AgentOption{
		agent.WithBaseURL(DefaultClientBaseURL),
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

func (c *Client) PostUser(ctx context.Context, r *PostUserRequest) (*User, error) {
	payload, err := json.Marshal(r)
	if err != nil {
		bencherror.WrapError(bencherror.SystemError, err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/user", bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("not created: %s", string(body))
	}

	var user *User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return user, nil
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

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("not OK: %s", string(body))
	}

	return nil
}

func (c *Client) GetUser(ctx context.Context, userID int) error {
	urlPath := fmt.Sprintf("/user/%d", userID)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return err
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("not OK: %s", string(body))
	}

	return nil
}

func (c *Client) GetUserTheme(ctx context.Context, userID int) error {
	urlPath := fmt.Sprintf("/user/%d/theme", userID)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return err
	}
	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("not OK: %s", string(body))
	}

	return nil
}

func (c *Client) ReserveLivestream(ctx context.Context, r *ReserveLivestreamRequest) (*Livestream, error) {
	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.WrapError(bencherror.SystemError, err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/livestream/reservation", bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("not created: %s", string(body))
	}
	var livestream *Livestream
	if err := json.NewDecoder(resp.Body).Decode(&livestream); err != nil {
		return nil, err
	}

	return livestream, err
}

func (c *Client) PostReaction(ctx context.Context, livestreamId int, r *PostReactionRequest) (*Reaction, error) {
	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.WrapError(bencherror.SystemError, err)
	}

	urlPath := fmt.Sprintf("/livestream/%d/reaction", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		// FIXME: それっぽいエラーに細分化
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, fmt.Errorf("リアクション投稿が正常に行えませんでした"))
	}
	reaction := &Reaction{}
	if err := json.NewDecoder(resp.Body).Decode(&reaction); err != nil {
		return reaction, err
	}

	benchscore.AddScore(benchscore.SuccessPostReaction)
	return reaction, nil
}

func (c *Client) PostSuperchat(ctx context.Context, livestreamId int, r *PostSuperchatRequest) (*PostSuperchatResponse, error) {
	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.WrapError(bencherror.SystemError, err)
	}

	urlPath := fmt.Sprintf("/livestream/%d/superchat", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("not created: %s", string(body))
	}

	var superchatResponse *PostSuperchatResponse
	if err := json.NewDecoder(resp.Body).Decode(&superchatResponse); err != nil {
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	// webappで正しくtipsの下限上限がバリデーションできているかチェック
	if superchatResponse.Tip < 0 || 20000 < superchatResponse.Tip {
		return superchatResponse, bencherror.WrapError(bencherror.BenchmarkCriticalError, err)
	}

	benchscore.AddScore(benchscore.SuccessPostSuperchat)
	benchscore.AddTipProfit(superchatResponse.Tip)

	return superchatResponse, nil
}

func (c *Client) ReportSuperchat(ctx context.Context, superchatId int) error {
	urlPath := fmt.Sprintf("/superchat/%d/report", superchatId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, nil)
	if err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("not created: %s", string(body))
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

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("not OK: %s", string(body))
	}

	return nil
}

func (c *Client) GetTags(ctx context.Context) error {
	req, err := c.agent.NewRequest(http.MethodGet, "/tag", nil)
	if err != nil {
		return bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("not OK: %s", string(body))
	}

	return nil
}

func (c *Client) GetReactions(ctx context.Context, livestreamID int) ([]Reaction, error) {
	req, err := c.agent.NewRequest(http.MethodGet, fmt.Sprintf("/livestream/%d/reaction", livestreamID), nil)
	if err != nil {
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, fmt.Errorf("リアクション一覧の取得に失敗しました"))
	}

	reactions := []Reaction{}
	if err := json.NewDecoder(resp.Body).Decode(&reactions); err != nil {
		return reactions, err
	}

	return reactions, nil
}

func (c *Client) GetSuperchats(ctx context.Context, livestreamID int) ([]Superchat, error) {
	req, err := c.agent.NewRequest(http.MethodGet, fmt.Sprintf("/livestream/%d/superchat", livestreamID), nil)
	if err != nil {
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, bencherror.WrapError(bencherror.BenchmarkApplicationError, fmt.Errorf("スーパーチャット一覧の取得に失敗しました"))
	}

	superchats := []Superchat{}
	if err := json.NewDecoder(resp.Body).Decode(&superchats); err != nil {
		return superchats, err
	}

	return superchats, nil
}

// sendRequestはagent.Doをラップしたリクエスト送信関数
// bencherror.WrapErrorはここで実行しているので、呼び出し側ではwrapしない
func (c *Client) sendRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		var (
			netErr net.Error
		)
		if errors.Is(context.DeadlineExceeded, err) {
			// 締切がすぎるのはベンチの都合なので、減点しない
			return resp, err
		} else if errors.As(err, &netErr) {
			if netErr.Timeout() {
				return resp, bencherror.WrapError(bencherror.BenchmarkTimeoutError, err)
			} else {
				// 接続ができないなど、ベンチ継続する上で致命的なエラー
				return resp, bencherror.WrapError(bencherror.BenchmarkCriticalError, err)
			}
		} else {
			// app errors
			return resp, bencherror.WrapError(bencherror.BenchmarkApplicationError, err)
		}
	}

	return resp, nil
}
