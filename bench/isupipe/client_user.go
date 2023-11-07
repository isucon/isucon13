package isupipe

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/internal/scheduler"
)

type User struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Theme       Theme  `json:"theme"`
}

type (
	RegisterRequest struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
		// Password is non-hashed password.
		Password string `json:"password"`
		Theme    Theme  `json:"theme"`
	}
	LoginRequest struct {
		UserName string `json:"username"`
		// Password is non-hashed password.
		Password string `json:"password"`
	}
)

type Theme struct {
	DarkMode bool `json:"dark_mode"`
}

type PostIconRequest struct {
	Image []byte `json:"image"`
}

type PostIconResponse struct {
	ID int64 `json:"id"`
}

func (c *Client) GetStreamerTheme(ctx context.Context, streamer *User, opts ...ClientOption) (*Theme, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	endpoint := fmt.Sprintf("/api/user/%s/theme", streamer.Name)
	req, err := c.agent.NewRequest(http.MethodGet, endpoint, nil)
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

	var theme *Theme
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&theme); err != nil {
			return nil, err
		}
	}

	return theme, nil
}

func (c *Client) GetIcon(ctx context.Context, username string, opts ...ClientOption) ([]byte, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	// FIXME: 配信者のユーザ名を含めてリクエスト
	endpoint := fmt.Sprintf("/api/user/%s/icon", username)
	req, err := c.assetAgent.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	resp, err := sendRequest(ctx, c.assetAgent, req)
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

	var imageBytes []byte
	if resp.StatusCode == defaultStatusCode {
		imageBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}
	}

	return imageBytes, nil
}

func (c *Client) PostIcon(ctx context.Context, r *PostIconRequest, opts ...ClientOption) (*PostIconResponse, error) {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	// FIXME: 配信者のユーザ名を含めてリクエスト
	endpoint := "/api/icon"
	req, err := c.agent.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
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

	var iconResp *PostIconResponse
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&iconResp); err != nil {
			return nil, err
		}
	}

	return iconResp, nil
}

func (c *Client) GetUser(ctx context.Context, username string, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/api/user/%s", username)
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

	return nil
}

func (c *Client) GetMe(ctx context.Context, opts ...ClientOption) (*User, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/api/user/me", nil)
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

	var user *User
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}
	}

	return user, nil
}

func (c *Client) Register(ctx context.Context, r *RegisterRequest, opts ...ClientOption) (*User, error) {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		// sendRequestはWrapErrorを行っているのでそのままreturn
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var user *User
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}
	}

	return user, nil
}

// ログインを行う.
// NOTE: ログイン後はログインユーザとして振る舞うので、各種agentやユーザ名、人気ユーザであるかの判定フラグなどの情報もここで確定する
func (c *Client) Login(ctx context.Context, r *LoginRequest, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(payload))
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

	c.username = r.UserName
	c.isPopular = scheduler.UserScheduler.IsPopularStreamer(c.username)

	// cookieを流用して各種ページアクセス用agentを初期化
	dnsResolver := resolver.NewNativeDNSResolver()
	domain := fmt.Sprintf("%s.u.isucon.dev", r.UserName)
	c.themeAgent, err = agent.NewAgent(
		agent.WithBaseURL(fmt.Sprintf("http://%s:%d", domain, config.TargetPort)),
		withClient(c.agent.HttpClient),
		agent.WithCloneTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			// Custom DNS Resolver
			DialContext: dnsResolver.DialContext,
		}),
		agent.WithNoCache(),
	)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	c.assetAgent, err = agent.NewAgent(
		agent.WithBaseURL(config.TargetBaseURL),
		withClient(c.agent.HttpClient),
		agent.WithCloneTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			// Custom DNS Resolver
			DialContext: dnsResolver.DialContext,
		}),
		// NOTE: 画像はキャッシュできるようにする
	)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	return nil
}
