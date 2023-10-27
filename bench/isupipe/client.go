package isupipe

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
)

var ErrCancelRequest = errors.New("contextのタイムアウトによりリクエストがキャンセルされます")

type Client struct {
	agent *agent.Agent

	// ユーザカスタムテーマ適用ページアクセス用agent
	// ライブ配信画面など
	themeAgent *agent.Agent

	// 画像ダウンロード用agent
	assetAgent *agent.Agent
}

func NewClient(customOpts ...agent.AgentOption) (*Client, error) {
	opts := []agent.AgentOption{
		agent.WithBaseURL(config.TargetBaseURL),
		agent.WithCloneTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			// Custom DNS Resolver
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				dialTimeout := 10000 * time.Millisecond
				dialer := net.Dialer{
					Timeout: dialTimeout,
					Resolver: &net.Resolver{
						PreferGo: true,
						Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
							dialer := net.Dialer{Timeout: dialTimeout}
							nameserver := net.JoinHostPort(config.TargetNameserver, strconv.Itoa(config.DNSPort))
							return dialer.DialContext(ctx, "udp", nameserver)
						},
					},
				}
				return dialer.DialContext(ctx, network, address)
			},
		}),
		agent.WithNoCache(),
		agent.WithTimeout(1 * time.Second),
	}
	for _, customOpt := range customOpts {
		opts = append(opts, customOpt)
	}
	customAgent, err := agent.NewAgent(opts...)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	return &Client{
		agent: customAgent,
	}, nil
}

func (c *Client) Initialize(ctx context.Context) (*InitializeResponse, error) {
	req, err := c.agent.NewRequest(http.MethodPost, "/initialize", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var initializeResp *InitializeResponse
	if json.NewDecoder(resp.Body).Decode(&initializeResp); err != nil {
		return nil, err
	}

	return initializeResp, nil
}

func (c *Client) PostUser(ctx context.Context, r *PostUserRequest, opts ...ClientOption) (*User, error) {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/user", bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		// sendRequestはWrapErrorを行っているのでそのままreturn
		return nil, err
	}

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var user *User
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}
	}

	benchscore.AddScore(benchscore.SuccessRegister)
	return user, nil
}

func (c *Client) Login(ctx context.Context, r *LoginRequest, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/login", bytes.NewReader(payload))
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	// cookieを流用して各種ページアクセス用agentを初期化
	domain := fmt.Sprintf("%s.u.isucon.dev", r.UserName)
	c.themeAgent, err = agent.NewAgent(
		agent.WithBaseURL(fmt.Sprintf("http://%s:12345", domain)),
		WithClient(c.agent.HttpClient),
		agent.WithNoCache(),
	)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	c.assetAgent, err = agent.NewAgent(
		agent.WithBaseURL(config.TargetBaseURL),
		WithClient(c.agent.HttpClient),
		// NOTE: 画像はキャッシュできるようにする
	)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	benchscore.AddScore(benchscore.SuccessLogin)
	return nil
}

// FIXME: meに変える
func (c *Client) GetUserSession(ctx context.Context, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/user/me", nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	return nil
}

func (c *Client) GetUser(ctx context.Context, username string, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/user/%s", username)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessGetUser)
	return nil
}

// FIXME: Hostヘッダにusernameを含めたドメインを入れてリクエスト
func (c *Client) GetStreamerTheme(ctx context.Context, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/theme", nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessGetUserTheme)
	return nil
}

func (c *Client) ReserveLivestream(ctx context.Context, r *ReserveLivestreamRequest, opts ...ClientOption) (*Livestream, error) {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/livestream/reservation", bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var livestream *Livestream
	if err := json.NewDecoder(resp.Body).Decode(&livestream); err != nil {
		return nil, bencherror.NewHttpResponseError(err, req)
	}

	benchscore.AddScore(benchscore.SuccessReserveLivestream)
	return livestream, nil
}

func (c *Client) PostReaction(ctx context.Context, livestreamId int, r *PostReactionRequest, opts ...ClientOption) (*Reaction, error) {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	urlPath := fmt.Sprintf("/livestream/%d/reaction", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	reaction := &Reaction{}
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&reaction); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}
	}

	benchscore.AddScore(benchscore.SuccessPostReaction)
	return reaction, nil
}

func (c *Client) PostLivecomment(ctx context.Context, livestreamId int, r *PostLivecommentRequest, opts ...ClientOption) (*PostLivecommentResponse, error) {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	urlPath := fmt.Sprintf("/livestream/%d/livecomment", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

func (c *Client) ReportLivecomment(ctx context.Context, livestreamId, livecommentId int, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d/livecomment/%d/report", livestreamId, livecommentId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessReportLivecomment)
	return nil
}

func (c *Client) Moderate(ctx context.Context, livestreamId int, ngWord string, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusCreated
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d/moderate", livestreamId)
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
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	return nil
}

func (c *Client) GetLivestream(
	ctx context.Context,
	livestreamId int,
	opts ...ClientOption,
) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d", livestreamId)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessGetLivestream)
	return nil
}

func (c *Client) GetLivestreams(
	ctx context.Context,
	opts ...ClientOption,
) ([]*Livestream, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/livestream", nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var livestreams []*Livestream
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&livestreams); err != nil {
			return nil, err
		}
	}

	return livestreams, nil
}

func (c *Client) GetLivestreamsByTag(
	ctx context.Context,
	tag string,
	opts ...ClientOption,
) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/livestream", nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	query := req.URL.Query()
	query.Add("tag", tag)
	req.URL.RawQuery = query.Encode()

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessGetLivestreamByTag)
	return nil
}

func (c *Client) GetTags(ctx context.Context, opts ...ClientOption) (*TagsResponse, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/tag", nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var tags *TagsResponse
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
			return nil, err
		}
	}

	benchscore.AddScore(benchscore.SuccessGetTags)
	return tags, nil
}

func (c *Client) GetReactions(ctx context.Context, livestreamId int, opts ...ClientOption) ([]Reaction, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d/reaction", livestreamId)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	reactions := []Reaction{}
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&reactions); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}
	}

	benchscore.AddScore(benchscore.SuccessGetReactions)
	return reactions, nil
}

func (c *Client) GetLivecomments(ctx context.Context, livestreamId int, opts ...ClientOption) ([]Livecomment, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d/livecomment", livestreamId)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

func (c *Client) GetUsers(ctx context.Context, opts ...ClientOption) ([]*User, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/user", nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var users []*User
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
			return users, bencherror.NewHttpResponseError(err, req)
		}
	}

	benchscore.AddScore(benchscore.SuccessGetUsers)
	return users, nil
}

func (c *Client) GetLivecommentReports(ctx context.Context, livestreamId int, opts ...ClientOption) ([]LivecommentReport, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d/report", livestreamId)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

func (c *Client) EnterLivestream(ctx context.Context, livestreamId int, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d/enter", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	req.Header.Add("Content-Type", "application/json;chatset=utf-8")

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessEnterLivestream)
	return nil
}

func (c *Client) LeaveLivestream(ctx context.Context, livestreamId int, opts ...ClientOption) error {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d/enter", livestreamId)
	req, err := c.agent.NewRequest(http.MethodDelete, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	benchscore.AddScore(benchscore.SuccessLeaveLivestream)
	return nil
}

// FIXME: 統計情報取得

func (c *Client) GetPaymentResult(ctx context.Context) (*PaymentResult, error) {
	req, err := c.agent.NewRequest(http.MethodGet, "/payment", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var paymentResp *PaymentResult
	if json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, err
	}

	return paymentResp, nil
}

func (c *Client) GetUserStatistics(ctx context.Context, username string, opts ...ClientOption) (*UserStatistics, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/user/%s/statistics", username)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var stats *UserStatistics
	if resp.StatusCode == defaultStatusCode {
		if json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return nil, err
		}
	}

	return stats, nil
}

func (c *Client) GetLivestreamStatistics(ctx context.Context, livestreamId int, opts ...ClientOption) (*LivestreamStatistics, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d/statistics", livestreamId)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var stats *LivestreamStatistics
	if resp.StatusCode == defaultStatusCode {
		if json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return nil, err
		}
	}

	return stats, nil
}

// sendRequestはagent.Doをラップしたリクエスト送信関数
// bencherror.WrapErrorはここで実行しているので、呼び出し側ではwrapしない
func (c *Client) sendRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s %s", req.Method, req.URL.EscapedPath())
	resp, err := c.agent.Do(ctx, req)
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
