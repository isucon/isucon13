package isupipe

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/assets"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"go.uber.org/zap"
)

var ErrCancelRequest = errors.New("ベンチマーク走行が継続できないエラーが発生しました")

// Client は、ISUPipeに対するHTTPクライアントです
// NOTE: スレッドセーフではありません
// NOTE: ログインは一度しかできません (何回もログインする場合はClientを個別に作り直す必要がある)
type Client struct {
	agent        *agent.Agent
	agentOptions []agent.AgentOption

	username string

	// ユーザカスタムテーマ適用ページアクセス用agent
	// ライブ配信画面など
	themeAgent   *agent.Agent
	themeOptions []agent.AgentOption

	// 画像ダウンロード用agent
	// キャッシュ可能
	assetAgent   *agent.Agent
	assetOptions []agent.AgentOption

	contestantLogger *zap.Logger
}

func NewClient(contestantLogger *zap.Logger, customOpts ...agent.AgentOption) (*Client, error) {
	return NewCustomResolverClient(contestantLogger, resolver.NewDNSResolver(), customOpts...)
}

// NewClient は、HTTPクライアント群を初期化します
// NOTE: キャッシュ無効化オプションなどを指定すると、意図しない挙動をする可能性があります
// タイムアウトやURLなどの振る舞いでないパラメータを指定するのにcustomOptsを用いてください
func NewCustomResolverClient(contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver, customOpts ...agent.AgentOption) (*Client, error) {
	opts := []agent.AgentOption{
		agent.WithBaseURL(config.TargetBaseURL),
		agent.WithCloneTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.InsecureSkipVerify,
			},
			DialContext:       dnsResolver.DialContext,
			IdleConnTimeout:   config.ClientIdleConnTimeout,
			ForceAttemptHTTP2: true,
		}),
		agent.WithTimeout(config.DefaultAgentTimeout),
		agent.WithNoCache(),
	}
	for _, customOpt := range customOpts {
		opts = append(opts, customOpt)
	}

	baseAgent, err := agent.NewAgent(opts...)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	themeOpts := []agent.AgentOption{
		withClient(baseAgent.HttpClient),
		agent.WithCloneTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.InsecureSkipVerify,
			},
			// Custom DNS Resolver
			DialContext:       dnsResolver.DialContext,
			IdleConnTimeout:   config.ClientIdleConnTimeout,
			ForceAttemptHTTP2: true,
		}),
		agent.WithTimeout(config.DefaultAgentTimeout),
		agent.WithNoCache(),
	}
	for _, customOpt := range customOpts {
		themeOpts = append(themeOpts, customOpt)
	}

	assetOpts := []agent.AgentOption{
		agent.WithBaseURL(config.TargetBaseURL),
		withClient(baseAgent.HttpClient),
		agent.WithCloneTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.InsecureSkipVerify,
			},
			DialContext:     dnsResolver.DialContext,
			IdleConnTimeout: config.ClientIdleConnTimeout,
		}),
		agent.WithTimeout(config.DefaultAgentTimeout),
		agent.WithNoCache(),
	}
	for _, customOpt := range customOpts {
		assetOpts = append(assetOpts, customOpt)
	}

	client := &Client{
		agent:            baseAgent,
		themeOptions:     themeOpts,
		assetOptions:     assetOpts,
		contestantLogger: contestantLogger,
	}
	if contestantLogger != nil {
		client.contestantLogger = contestantLogger
	} else {
		client.contestantLogger = zap.NewNop()
	}

	return client, nil
}

func (c *Client) Username() (string, error) {
	if len(c.username) == 0 {
		return "", bencherror.NewInternalError(fmt.Errorf("未ログインクライアントです"))
	}

	return c.username, nil
}

func (c *Client) ValidateAssets(ctx context.Context) error {
	allAssets, err := assets.Load()
	if err != nil {
		return err
	}
	for _, asset := range allAssets {
		fmt.Println(asset.Path)
		req, err := c.agent.NewRequest(http.MethodGet, asset.Path, nil)
		if err != nil {
			return bencherror.NewInternalError(err)
		}
		resp, err := sendRequest(ctx, c.agent, req)
		if err != nil {
			fmt.Printf("%#v\n", err)
			return err
		}
		defer func() {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		actualAssetHash := sha256.Sum256(b)
		if asset.Hash != actualAssetHash {
			return fmt.Errorf("ファイル %s のハッシュが一致しません actual: %x expected: %x", asset.Path, actualAssetHash, asset.Hash)
		}
	}
	return nil
}

func (c *Client) setStreamerURL(streamerName string) error {
	domain := fmt.Sprintf("%s.%s", streamerName, config.BaseDomain)
	baseURL, err := url.Parse(fmt.Sprintf("%s://%s:%d", config.HTTPScheme, domain, config.TargetPort))
	if err != nil {
		return err
	}

	c.themeAgent.BaseURL = baseURL

	return nil
}

// sendRequestはagent.Doをラップしたリクエスト送信関数
// bencherror.WrapErrorはここで実行しているので、呼び出し側ではwrapしない
func sendRequest(ctx context.Context, agent *agent.Agent, req *http.Request) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s %s", req.Method, req.URL.EscapedPath())
	resp, err := agent.Do(ctx, req)
	if err != nil {
		var (
			netErr net.Error
		)
		if errors.Is(err, context.DeadlineExceeded) {
			// 締切がすぎるのはベンチの都合なので、減点しない
			// リクエストをキャンセルする
			return resp, ErrCancelRequest
		} else if errors.As(err, &netErr) {
			if netErr.Timeout() {
				return resp, bencherror.NewTimeoutError(err, "%s", endpoint)
			} else {
				return resp, ErrCancelRequest
			}
		} else {
			return resp, bencherror.NewApplicationError(err, "%s に対するリクエストが失敗しました", endpoint)
		}
	}

	return resp, nil
}
