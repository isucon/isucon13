package isupipe

type ClientOption func(o *ClientOptions)

type LimitParam struct {
	Limit int
}

type SearchTagParam struct {
	Tag string
}

type ClientOptions struct {
	wantStatusCode int
	limitParam     *LimitParam
	searchTag      *SearchTagParam
	eTag           string
	// NOTE: スパム報告は、ベンチ走行中は粛清されたライブコメントを期待する場合が有り、エラーになることがある
	// Pretestでのみスパム報告のバリデーションを行うための対応
	validateReportLivecomment bool
}

func newClientOptions(defaultStatusCode int, opts ...ClientOption) *ClientOptions {
	o := &ClientOptions{
		wantStatusCode: defaultStatusCode,
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(o)
	}
	return o
}

func WithStatusCode(statusCode int) ClientOption {
	return func(o *ClientOptions) {
		o.wantStatusCode = statusCode
	}
}

func WithLimitQueryParam(limit int) ClientOption {
	return func(o *ClientOptions) {
		o.limitParam = &LimitParam{
			Limit: limit,
		}
	}
}

func WithSearchTagQueryParam(tag string) ClientOption {
	return func(o *ClientOptions) {
		o.searchTag = &SearchTagParam{
			Tag: tag,
		}
	}
}

func WithETag(eTag string) ClientOption {
	return func(o *ClientOptions) {
		o.eTag = eTag
	}
}

func WithValidateReportLivecomment() ClientOption {
	return func(o *ClientOptions) {
		o.validateReportLivecomment = true
	}
}
