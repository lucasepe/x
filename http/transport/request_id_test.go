package transport_test

import (
	"net/http"
	"testing"

	"github.com/lucasepe/x/http/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDRoundTripperAddsGeneratedHeaderWithoutMutatingOriginalRequest(t *testing.T) {
	upstream := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "req-123", req.Header.Get("X-Request-Id"))
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Request:    req,
			Body:       http.NoBody,
		}, nil
	})

	rt := transport.RequestIDRoundTripperWithOptions(upstream, transport.RequestIDOptions{
		Generator: func() string { return "req-123" },
	})

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Empty(t, req.Header.Get("X-Request-Id"))
}

func TestRequestIDRoundTripperPreservesExistingHeader(t *testing.T) {
	upstream := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "already-set", req.Header.Get("X-Correlation-Id"))
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Request:    req,
			Body:       http.NoBody,
		}, nil
	})

	rt := transport.RequestIDRoundTripperWithOptions(upstream, transport.RequestIDOptions{
		HeaderName: "X-Correlation-Id",
		Generator:  func() string { return "generated" },
	})

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)
	req.Header.Set("X-Correlation-Id", "already-set")

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "already-set", req.Header.Get("X-Correlation-Id"))
}
