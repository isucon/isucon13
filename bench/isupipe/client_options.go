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
	spamCheck      bool
	limitParam     *LimitParam
	searchTag      *SearchTagParam
	eTag           string
}

func newClientOptions(defaultStatusCode int, opts ...ClientOption) *ClientOptions {
	o := &ClientOptions{
		wantStatusCode: defaultStatusCode,
		spamCheck:      true,
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

func WithNoSpamCheck() ClientOption {
	return func(o *ClientOptions) {
		o.spamCheck = false
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
