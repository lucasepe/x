package transport_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/lucasepe/x/http/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestForDebugUsesComposableOrder(t *testing.T) {
	steps := make([]string, 0, 3)

	builder := transport.ForDebug(transport.NewTransportBuilder()).
		Use(func(next http.RoundTripper) http.RoundTripper {
			return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				steps = append(steps, "custom")
				assert.NotEmpty(t, req.Header.Get("X-Request-Id"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Header:     make(http.Header),
					Request:    req,
					Body:       http.NoBody,
				}, nil
			})
		})

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	resp, err := builder.Build().RoundTrip(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Il custom layer aggiunto dopo il preset resta il piu' esterno.
	assert.Equal(t, []string{"custom"}, steps)
}

func TestForDebugCreatesBuilderWhenNil(t *testing.T) {
	builder := transport.ForDebug(nil).
		Use(func(next http.RoundTripper) http.RoundTripper {
			return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				assert.NotEmpty(t, req.Header.Get("X-Request-Id"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Header:     make(http.Header),
					Request:    req,
					Body:       http.NoBody,
				}, nil
			})
		})
	require.NotNil(t, builder)

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	resp, err := builder.Build().RoundTrip(req)
	require.NoError(t, err)
	resp.Body.Close()
}

func TestForScrapingCreatesBuilderWhenNil(t *testing.T) {
	builder := transport.ForScraping(nil).
		Use(func(next http.RoundTripper) http.RoundTripper {
			return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				assert.NotEmpty(t, req.Header.Get("User-Agent"))
				assert.NotEmpty(t, req.Header.Get("Accept"))
				assert.NotEmpty(t, req.Header.Get("Accept-Language"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Header:     make(http.Header),
					Request:    req,
					Body:       http.NoBody,
				}, nil
			})
		})
	require.NotNil(t, builder)

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	resp, err := builder.Build().RoundTrip(req)
	require.NoError(t, err)
	resp.Body.Close()
}

func TestForDebugWithOptionsUsesCustomRequestIDHeader(t *testing.T) {
	builder := transport.ForDebugWithOptions(nil, transport.DebugPresetOptions{
		RequestID: transport.RequestIDOptions{
			HeaderName: "X-Correlation-Id",
			Generator:  func() string { return "corr-1" },
		},
		Verbose: transport.VerboseOptions{
			LogBodies: false,
		},
	}).Use(func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "corr-1", req.Header.Get("X-Correlation-Id"))
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     make(http.Header),
				Request:    req,
				Body:       http.NoBody,
			}, nil
		})
	})

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	resp, err := builder.Build().RoundTrip(req)
	require.NoError(t, err)
	resp.Body.Close()
}

func TestForScrapingWithOptionsSupportsPOSTRetry(t *testing.T) {
	var calls int

	builder := transport.ForScrapingWithOptions(nil, transport.ScrapingPresetOptions{
		HostRate:  rate.Inf,
		HostBurst: 10,
		Retry: transport.RetryOptions{
			MaxAttempts: 2,
			StatusCodes: []int{http.StatusTooManyRequests},
			Methods:     []string{http.MethodPost},
			BaseDelay:   time.Millisecond,
			MaxDelay:    5 * time.Millisecond,
		},
	}).Use(func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			assert.NotEmpty(t, req.Header.Get("User-Agent"))
			if calls == 1 {
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
				Body:       http.NoBody,
			}, nil
		})
	})

	req, err := http.NewRequest(http.MethodPost, "https://api.openrouteservice.org/v2/directions", nil)
	require.NoError(t, err)

	resp, err := builder.Build().RoundTrip(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, 2, calls)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
