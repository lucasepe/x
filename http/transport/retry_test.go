package transport_test

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lucasepe/x/http/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryRoundTripperRetriesOnConfiguredStatus(t *testing.T) {
	var calls int32

	upstream := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		attempt := atomic.AddInt32(&calls, 1)
		if attempt == 1 {
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Status:     "429 Too Many Requests",
				Header:     make(http.Header),
				Request:    req,
				Body:       http.NoBody,
			}, nil
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Request:    req,
			Body:       io.NopCloser(strings.NewReader("ok")),
		}, nil
	})

	rt := transport.RetryRoundTripper(upstream, transport.RetryOptions{
		MaxAttempts: 2,
		StatusCodes: []int{http.StatusTooManyRequests},
		BaseDelay:   time.Millisecond,
		MaxDelay:    5 * time.Millisecond,
	})

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", string(body))
}

func TestRetryRoundTripperDoesNotRetryNonReplayableBody(t *testing.T) {
	var calls int32

	upstream := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Status:     "429 Too Many Requests",
			Header:     make(http.Header),
			Request:    req,
			Body:       http.NoBody,
		}, nil
	})

	rt := transport.RetryRoundTripper(upstream, transport.RetryOptions{
		MaxAttempts: 3,
		StatusCodes: []int{http.StatusTooManyRequests},
		Methods:     []string{http.MethodPost},
		BaseDelay:   time.Millisecond,
		MaxDelay:    5 * time.Millisecond,
	})

	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)
	req.Body = io.NopCloser(strings.NewReader(`{"a":1}`))
	req.GetBody = nil

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
