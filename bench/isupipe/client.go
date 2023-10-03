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
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
)

var ErrCancelRequest = errors.New("contextのタイムアウトによりリクエストがキャンセルされます")

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

func (c *Client) PostUser(ctx context.Context, r *PostUserRequest, options ...AssertOption) (*User, error) {
	pat := ClientAssertPattern{
		StatusCode: http.StatusCreated,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/user", bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		// sendRequestはWrapErrorを行っているのでそのままreturn
		return nil, err
	}

	if err := pat.assertStatuscode(req, resp); err != nil {
		return nil, err
	}

	var user *User
	if pat.DecodeBody {
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return nil, bencherror.NewHttpResponseError(err, req)
		}
	}

	benchscore.AddScore(benchscore.SuccessRegister)
	return user, nil
}

func (c *Client) Login(ctx context.Context, r *LoginRequest, options ...AssertOption) error {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
	}

	for _, option := range options {
		option(&pat)
	}
	payload, err := json.Marshal(r)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/login", bytes.NewReader(payload))
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if err := pat.assertStatuscode(req, resp); err != nil {
		return err
	}

	benchscore.AddScore(benchscore.SuccessLogin)
	return nil
}

func (c *Client) GetUser(ctx context.Context, userId int, options ...AssertOption) error {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

	urlPath := fmt.Sprintf("/user/%d", userId)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := pat.assertStatuscode(req, resp); err != nil {
		return err
	}

	benchscore.AddScore(benchscore.SuccessGetUser)
	return nil
}

// FIXME: Hostヘッダにusernameを含めたドメインを入れてリクエスト
func (c *Client) GetStreamerTheme(ctx context.Context, options ...AssertOption) error {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

	req, err := c.agent.NewRequest(http.MethodGet, "/theme", nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if err := pat.assertStatuscode(req, resp); err != nil {
		return err
	}

	benchscore.AddScore(benchscore.SuccessGetUserTheme)
	return nil
}

func (c *Client) ReserveLivestream(ctx context.Context, r *ReserveLivestreamRequest, options ...AssertOption) (*Livestream, error) {
	pat := ClientAssertPattern{
		StatusCode: http.StatusCreated,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/livestream/reservation", bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := pat.assertStatuscode(req, resp); err != nil {
		return nil, err
	}

	var livestream *Livestream
	if err := json.NewDecoder(resp.Body).Decode(&livestream); err != nil {
		return nil, bencherror.NewHttpResponseError(err, req)
	}

	benchscore.AddScore(benchscore.SuccessReserveLivestream)
	return livestream, nil
}

func (c *Client) PostReaction(ctx context.Context, livestreamId int, r *PostReactionRequest, options ...AssertOption) (*Reaction, error) {
	pat := ClientAssertPattern{
		StatusCode: http.StatusCreated,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	urlPath := fmt.Sprintf("/livestream/%d/reaction", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := pat.assertStatuscode(req, resp); err != nil {
		return nil, err
	}

	reaction := &Reaction{}
	if err := json.NewDecoder(resp.Body).Decode(&reaction); err != nil {
		return nil, bencherror.NewHttpResponseError(err, req)
	}

	benchscore.AddScore(benchscore.SuccessPostReaction)
	return reaction, nil
}

func (c *Client) PostLivecomment(ctx context.Context, livestreamId int, r *PostLivecommentRequest, options ...AssertOption) (*PostLivecommentResponse, error) {
	pat := ClientAssertPattern{
		StatusCode: http.StatusCreated,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	urlPath := fmt.Sprintf("/livestream/%d/livecomment", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, bytes.NewReader(payload))
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := pat.assertStatuscode(req, resp); err != nil {
		return nil, err
	}

	var livecommentResponse *PostLivecommentResponse
	if err := json.NewDecoder(resp.Body).Decode(&livecommentResponse); err != nil {
		return nil, bencherror.NewHttpResponseError(err, req)
	}

	benchscore.AddScore(benchscore.SuccessPostLivecomment)
	benchscore.AddTipProfit(livecommentResponse.Tip)

	return livecommentResponse, nil
}

func (c *Client) ReportLivecomment(ctx context.Context, livestreamId, livecommentId int, options ...AssertOption) error {
	pat := ClientAssertPattern{
		StatusCode: http.StatusCreated,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

	urlPath := fmt.Sprintf("/livestream/%d/livecomment/%d/report", livestreamId, livecommentId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if err := pat.assertStatuscode(req, resp); err != nil {
		return err
	}

	benchscore.AddScore(benchscore.SuccessReportLivecomment)
	return nil
}

func (c *Client) Moderate(ctx context.Context, livestreamId int, ngWord string, options ...AssertOption) error {
	pat := ClientAssertPattern{
		StatusCode: http.StatusCreated,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

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

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := pat.assertStatuscode(req, resp); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetLivestream(
	ctx context.Context,
	livestreamId int,
	options ...AssertOption,
) error {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

	urlPath := fmt.Sprintf("/livestream/%d", livestreamId)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if err := pat.assertStatuscode(req, resp); err != nil {
		return err
	}

	benchscore.AddScore(benchscore.SuccessGetLivestream)
	return nil
}

func (c *Client) GetLivestreamsByTag(
	ctx context.Context,
	tag string,
	options ...AssertOption,
) error {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

	urlPath := fmt.Sprintf("/livestream?tag=%s", tag)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if err := pat.assertStatuscode(req, resp); err != nil {
		return err
	}

	benchscore.AddScore(benchscore.SuccessGetLivestreamByTag)
	return nil
}

func (c *Client) GetTags(ctx context.Context, options ...AssertOption) (*TagsResponse, error) {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

	req, err := c.agent.NewRequest(http.MethodGet, "/tag", nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := pat.assertStatuscode(req, resp); err != nil {
		return nil, err
	}

	var tags *TagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	benchscore.AddScore(benchscore.SuccessGetTags)
	return tags, nil
}

func (c *Client) GetReactions(ctx context.Context, livestreamId int, options ...AssertOption) ([]Reaction, error) {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

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

	if err := pat.assertStatuscode(req, resp); err != nil {
		return nil, err
	}

	reactions := []Reaction{}
	if err := json.NewDecoder(resp.Body).Decode(&reactions); err != nil {
		return nil, bencherror.NewHttpResponseError(err, req)
	}

	benchscore.AddScore(benchscore.SuccessGetReactions)
	return reactions, nil
}

func (c *Client) GetLivecomments(ctx context.Context, livestreamId int, options ...AssertOption) ([]Livecomment, error) {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}
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

	if err := pat.assertStatuscode(req, resp); err != nil {
		return nil, err
	}

	livecomments := []Livecomment{}
	if err := json.NewDecoder(resp.Body).Decode(&livecomments); err != nil {
		return livecomments, bencherror.NewHttpResponseError(err, req)
	}

	benchscore.AddScore(benchscore.SuccessGetLivecomments)
	return livecomments, nil
}

func (c *Client) GetUsers(ctx context.Context, options ...AssertOption) ([]*User, error) {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}
	req, err := c.agent.NewRequest(http.MethodGet, "/user", nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := pat.assertStatuscode(req, resp); err != nil {
		return nil, err
	}

	var users []*User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return users, bencherror.NewHttpResponseError(err, req)
	}

	benchscore.AddScore(benchscore.SuccessGetUsers)
	return users, nil
}

func (c *Client) GetLivecommentReports(ctx context.Context, livestreamId int, options ...AssertOption) ([]LivecommentReport, error) {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

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

	if err := pat.assertStatuscode(req, resp); err != nil {
		return nil, err
	}

	reports := []LivecommentReport{}
	if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
		return reports, bencherror.NewHttpResponseError(err, req)
	}

	benchscore.AddScore(benchscore.SuccessGetLivecommentReports)
	return reports, nil
}

func (c *Client) EnterLivestream(ctx context.Context, livestreamId int, options ...AssertOption) error {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

	urlPath := fmt.Sprintf("/livestream/%d/enter", livestreamId)
	req, err := c.agent.NewRequest(http.MethodPost, urlPath, nil)
	if err != nil {
		return bencherror.NewInternalError(err)
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := pat.assertStatuscode(req, resp); err != nil {
		return err
	}

	benchscore.AddScore(benchscore.SuccessEnterLivestream)
	return nil
}

func (c *Client) LeaveLivestream(ctx context.Context, livestreamId int, options ...AssertOption) error {
	pat := ClientAssertPattern{
		StatusCode: http.StatusOK,
		DecodeBody: true,
	}

	for _, option := range options {
		option(&pat)
	}

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

	if err := pat.assertStatuscode(req, resp); err != nil {
		return err
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

type ClientAssertPattern struct {
	StatusCode int
	SetCookie  bool
	DecodeBody bool
}

func (pat *ClientAssertPattern) assertStatuscode(
	req *http.Request,
	resp *http.Response,
) error {
	if resp.StatusCode != pat.StatusCode {
		return bencherror.NewHttpStatusError(req, pat.StatusCode, resp.StatusCode)
	}

	return nil
}

type AssertOption = func(cap *ClientAssertPattern)

func WithStatusCode(code int) AssertOption {
	return func(cap *ClientAssertPattern) {
		cap.StatusCode = code
	}
}

func DecodeBody(decode bool) AssertOption {
	return func(cap *ClientAssertPattern) {
		cap.DecodeBody = decode
	}
}
