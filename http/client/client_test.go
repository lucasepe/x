package client_test

import (
	"testing"

	httpclient "github.com/lucasepe/x/http/client"
)

func TestHTTPClientForConfig(t *testing.T) {
	tests := []struct {
		name      string
		cfg       httpclient.Config
		expectErr bool
	}{
		{
			name:      "valid config without auth",
			cfg:       httpclient.Config{},
			expectErr: false,
		},
		{
			name:      "valid config with bearer token",
			cfg:       httpclient.Config{Token: "test-token"},
			expectErr: false,
		},
		{
			name:      "valid config with basic auth",
			cfg:       httpclient.Config{Username: "user", Password: "pass"},
			expectErr: false,
		},
		{
			name:      "invalid config with both auth methods",
			cfg:       httpclient.Config{Username: "user", Password: "pass", Token: "token"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := httpclient.HTTPClientForConfig(tc.cfg)
			if (err != nil) != tc.expectErr {
				t.Errorf("unexpected error status: got %v, expectErr %v", err, tc.expectErr)
			}
			if !tc.expectErr && client == nil {
				t.Errorf("expected client, got nil")
			}
		})
	}
}
