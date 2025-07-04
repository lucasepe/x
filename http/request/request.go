package request

import (
	"context"
	"io"
	"net/http"
	"strings"
)

type Options struct {
	BaseURL string
	Method  string
	Path    string
	Params  []string
	Headers []string
}

func NewWithContext(ctx context.Context, opts Options, body io.Reader) (*http.Request, error) {
	uri, err := composeURL(opts.BaseURL, opts.Path, opts.Params...)
	if err != nil {
		return nil, err
	}

	method := strings.ToUpper(strings.TrimSpace(opts.Method))
	if method == "" {
		method = http.MethodGet
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, err
	}

	setHeaders(req, opts.Headers...)

	return req, nil
}

func New(opts Options, body io.Reader) (*http.Request, error) {
	return NewWithContext(context.Background(), opts, body)
}
