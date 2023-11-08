package isupipe

type ClientOption func(o *ClientOptions)

type ClientOptions struct {
	wantStatusCode int
	spamCheck      bool
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
