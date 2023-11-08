package isupipe

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
)

var ErrCancelRequest = errors.New("contextのタイムアウトによりリクエストがキャンセルされます")

type Client struct {
	agent     *agent.Agent
	username  string
	isPopular bool

	// ユーザカスタムテーマ適用ページアクセス用agent
	// ライブ配信画面など
	themeAgent *agent.Agent

	// 画像ダウンロード用agent
	// キャッシュ可能
	assetAgent *agent.Agent
}

// FIXME: テスト用に、ネームサーバのアドレスや接続先アドレスなどをオプションで渡せるように
func NewClient(dnsResolver *resolver.DNSResolver, customOpts ...agent.AgentOption) (*Client, error) {
	opts := []agent.AgentOption{
		agent.WithBaseURL(config.TargetBaseURL),
		agent.WithCloneTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.InsecureSkipVerify,
			},
			// Custom DNS Resolver
			DialContext: dnsResolver.DialContext,
		}),
		agent.WithTimeout(config.DefaultAgentTimeout),
		agent.WithNoCache(),
	}
	for _, customOpt := range customOpts {
		opts = append(opts, customOpt)
	}
	agent, err := agent.NewAgent(opts...)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	return &Client{
		agent: agent,
	}, nil
}

func (c *Client) LoginUserName() (string, error) {
	if len(c.username) == 0 {
		return "", bencherror.NewInternalError(fmt.Errorf("未ログインクライアントです"))
	}

	return c.username, nil
}

func (c *Client) IsPopular() bool {
	return c.isPopular
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
				// 接続ができないなど、ベンチ継続する上で致命的なエラー
				return resp, bencherror.NewViolationError(err, "webappの %s に対するリクエストで、接続に失敗しました", endpoint)
			}
		} else {
			return resp, bencherror.NewApplicationError(err, "%s に対するリクエストが失敗しました", endpoint)
		}
	}

	return resp, nil
}
