package request_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/lucasepe/x/http/request"
)

func TestNewWithContext(t *testing.T) {
	tests := []struct {
		name       string
		opts       request.Options
		body       io.Reader
		wantMethod string
		wantURL    string
		wantHeader map[string]string
		wantErr    bool
	}{
		{
			name: "simple GET request with no headers or params",
			opts: request.Options{
				BaseURL: "https://example.com",
				Method:  "GET",
				Path:    "api/test",
			},
			body:       nil,
			wantMethod: http.MethodGet,
			wantURL:    "https://example.com/api/test",
			wantHeader: map[string]string{},
			wantErr:    false,
		},
		{
			name: "POST with headers and body",
			opts: request.Options{
				BaseURL: "https://example.com",
				Method:  "post",
				Path:    "/submit",
				Headers: []string{
					"Content-Type: application/json",
					"Authorization: Bearer token",
				},
			},
			body:       bytes.NewBufferString(`{"name":"test"}`),
			wantMethod: http.MethodPost,
			wantURL:    "https://example.com/submit",
			wantHeader: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
			},
			wantErr: false,
		},
		{
			name: "query parameters",
			opts: request.Options{
				BaseURL: "https://example.com/",
				Path:    "/search",
				Params: []string{
					"q:test",
					"debug:",
				},
			},
			wantMethod: http.MethodGet,
			wantURL:    "https://example.com/search?q=test&debug=",
			wantHeader: map[string]string{},
			wantErr:    false,
		},
		{
			name: "invalid param format",
			opts: request.Options{
				BaseURL: "https://example.com",
				Path:    "/invalid",
				Params:  []string{"invalidparam"}, // no colon
			},
			wantMethod: http.MethodGet,
			wantURL:    "https://example.com/invalid",
			wantHeader: map[string]string{},
			wantErr:    false,
		},
		{
			name: "invalid URL",
			opts: request.Options{
				BaseURL: "://bad-url",
				Path:    "test",
			},
			wantErr: true,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := request.NewWithContext(ctx, tt.opts, tt.body)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if req.Method != tt.wantMethod {
				t.Errorf("expected method %s, got %s", tt.wantMethod, req.Method)
			}

			gotURL, _ := url.Parse(req.URL.String())
			wantURL, _ := url.Parse(tt.wantURL)

			if gotURL.Scheme != wantURL.Scheme || gotURL.Host != wantURL.Host || gotURL.Path != wantURL.Path {
				t.Errorf("unexpected URL: got %s, want %s", gotURL, wantURL)
			}

			if gotURL.Query().Encode() != wantURL.Query().Encode() {
				t.Errorf("unexpected query params: got %s, want %s", gotURL.RawQuery, wantURL.RawQuery)
			}

			for k, v := range tt.wantHeader {
				if req.Header.Get(k) != v {
					t.Errorf("expected header %s=%s, got %s", k, v, req.Header.Get(k))
				}
			}
		})
	}
}
