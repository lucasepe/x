package transport

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestStickyBrowserRoundTripperKeepsProfilePerHost(t *testing.T) {
	var firstUA string

	upstream := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		ua := req.Header.Get("User-Agent")
		require.NotEmpty(t, ua)
		require.NotEmpty(t, req.Header.Get("Accept"))
		require.NotEmpty(t, req.Header.Get("Accept-Language"))

		if firstUA == "" {
			firstUA = ua
		} else {
			assert.Equal(t, firstUA, ua)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Request:    req,
			Body:       http.NoBody,
		}, nil
	})

	rt := StickyBrowserRoundTripper(upstream)
	sticky, ok := rt.(*stickyBrowserTransport)
	require.True(t, ok)

	// Seed deterministico per avere test ripetibile.
	sticky.rnd.Seed(1)

	req1, err := http.NewRequest(http.MethodGet, "https://example.com/one", nil)
	require.NoError(t, err)
	_, err = rt.RoundTrip(req1)
	require.NoError(t, err)

	req2, err := http.NewRequest(http.MethodGet, "https://example.com/two", nil)
	require.NoError(t, err)
	_, err = rt.RoundTrip(req2)
	require.NoError(t, err)
}

func TestStickyBrowserRoundTripperClonesRequestAndClearsChromeHintsForNonChromeProfiles(t *testing.T) {
	upstream := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:125.0) Gecko/20100101 Firefox/125.0", req.Header.Get("User-Agent"))
		assert.Empty(t, req.Header.Get("Sec-CH-UA"))
		assert.Empty(t, req.Header.Get("Sec-CH-UA-Mobile"))
		assert.Empty(t, req.Header.Get("Sec-CH-UA-Platform"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Request:    req,
			Body:       http.NoBody,
		}, nil
	})

	rt := StickyBrowserRoundTripper(upstream)
	sticky, ok := rt.(*stickyBrowserTransport)
	require.True(t, ok)

	sticky.perHost["example.com"] = browserProfile{
		UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:125.0) Gecko/20100101 Firefox/125.0",
		Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		AcceptLanguage: "it-IT,it;q=0.9,en-US;q=0.8,en;q=0.7",
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)
	req.Header.Set("Sec-CH-UA", "stale")
	req.Header.Set("Sec-CH-UA-Mobile", "?1")
	req.Header.Set("Sec-CH-UA-Platform", "\"Android\"")

	_, err = rt.RoundTrip(req)
	require.NoError(t, err)

	// La request originale non deve essere mutata dal transport.
	assert.Equal(t, "stale", req.Header.Get("Sec-CH-UA"))
	assert.Equal(t, "?1", req.Header.Get("Sec-CH-UA-Mobile"))
	assert.Equal(t, "\"Android\"", req.Header.Get("Sec-CH-UA-Platform"))
	assert.Empty(t, req.Header.Get("User-Agent"))
}
