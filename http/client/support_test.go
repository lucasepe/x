package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProxyURL(t *testing.T) {
	tests := []struct {
		name      string
		proxyURL  string
		expectErr bool
	}{
		{"Valid HTTP proxy", "http://proxy.example.com", false},
		{"Valid HTTPS proxy", "https://secure-proxy.example.com", false},
		{"Invalid scheme", "ftp://invalid-proxy.com", true},
		{"Malformed URL", ":://bad-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := parseProxyURL(tt.proxyURL)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, u)
			}
		})
	}
}
