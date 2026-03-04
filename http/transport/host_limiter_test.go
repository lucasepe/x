package transport_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lucasepe/x/http/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestHostLimiterAppliesPerHostRateLimit(t *testing.T) {
	var upstreamCalls int32
	upstream := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&upstreamCalls, 1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Request:    req,
			Body:       http.NoBody,
		}, nil
	})

	rt := transport.HostLimiter(rate.Every(time.Hour), 1, upstream)

	req1, err := http.NewRequest(http.MethodGet, "https://api.example.com/one", nil)
	require.NoError(t, err)

	_, err = rt.RoundTrip(req1)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/two", nil)
	require.NoError(t, err)

	_, err = rt.RoundTrip(req2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "would exceed context deadline")
	assert.Equal(t, int32(1), atomic.LoadInt32(&upstreamCalls))
}
