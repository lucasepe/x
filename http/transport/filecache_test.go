package transport_test

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	xfilecache "github.com/lucasepe/x/filecache"
	"github.com/lucasepe/x/http/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestFileCacheTransportCachesGETResponses(t *testing.T) {
	t.Setenv("SKIP_CACHE", "false")

	dir := t.TempDir()
	cache, err := xfilecache.New(dir)
	require.NoError(t, err)

	var upstreamCalls int32
	upstream := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&upstreamCalls, 1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("cached body")),
			Request:    req,
		}, nil
	})

	rt := transport.FileCacheTransport(cache, upstream)

	req1, err := http.NewRequest(http.MethodGet, "https://example.com/resource", nil)
	require.NoError(t, err)

	resp1, err := rt.RoundTrip(req1)
	require.NoError(t, err)
	body1, err := io.ReadAll(resp1.Body)
	require.NoError(t, err)
	resp1.Body.Close()

	assert.Equal(t, "cached body", string(body1))
	assert.Equal(t, int32(1), atomic.LoadInt32(&upstreamCalls))

	req2, err := http.NewRequest(http.MethodGet, "https://example.com/resource", nil)
	require.NoError(t, err)

	resp2, err := rt.RoundTrip(req2)
	require.NoError(t, err)
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	resp2.Body.Close()

	assert.Equal(t, "cached body", string(body2))
	assert.Equal(t, "200 OK (from cache)", resp2.Status)
	assert.Equal(t, int32(1), atomic.LoadInt32(&upstreamCalls))
}

func TestFileCacheTransportPromotes304UsingCachedBody(t *testing.T) {
	t.Setenv("SKIP_CACHE", "true")

	dir := t.TempDir()
	cache, err := xfilecache.New(dir)
	require.NoError(t, err)
	cacheKey, err := transport.DigestCacheKeyMethodURL(mustRequest(t, http.MethodGet, "https://example.com/etag", nil), nil)
	require.NoError(t, err)
	require.NoError(t, cache.Put(cacheKey, []byte("etag body")))

	upstream := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotModified,
			Status:     "304 Not Modified",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	})

	rt := transport.FileCacheTransport(cache, upstream)

	req, err := http.NewRequest(http.MethodGet, "https://example.com/etag", nil)
	require.NoError(t, err)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "200 OK (from cache after 304)", resp.Status)
	assert.Equal(t, "etag body", string(body))
}

func mustRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	return req
}

func TestFileCacheTransportCachesPOSTOnlyWhenExplicitlyEnabled(t *testing.T) {
	t.Setenv("SKIP_CACHE", "false")

	dir := t.TempDir()
	cache, err := xfilecache.New(dir)
	require.NoError(t, err)

	var upstreamCalls int32
	upstream := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		payload, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		req.Body.Close()

		atomic.AddInt32(&upstreamCalls, 1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("route:" + string(payload))),
			Request:    req,
		}, nil
	})

	rt := transport.FileCacheTransportWithOptions(cache, upstream, transport.FileCacheOptions{
		Methods: []string{http.MethodGet, http.MethodPost},
		KeyFunc: transport.DigestCacheKeyMethodURLBody,
	})

	req1, err := http.NewRequest(http.MethodPost, "https://api.openrouteservice.org/v2/directions", strings.NewReader(`{"a":1}`))
	require.NoError(t, err)

	resp1, err := rt.RoundTrip(req1)
	require.NoError(t, err)
	body1, err := io.ReadAll(resp1.Body)
	require.NoError(t, err)
	resp1.Body.Close()

	assert.Equal(t, "route:{\"a\":1}", string(body1))
	assert.Equal(t, int32(1), atomic.LoadInt32(&upstreamCalls))

	req2, err := http.NewRequest(http.MethodPost, "https://api.openrouteservice.org/v2/directions", strings.NewReader(`{"a":1}`))
	require.NoError(t, err)

	resp2, err := rt.RoundTrip(req2)
	require.NoError(t, err)
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	resp2.Body.Close()

	assert.Equal(t, "route:{\"a\":1}", string(body2))
	assert.Equal(t, "200 OK (from cache)", resp2.Status)
	assert.Equal(t, int32(1), atomic.LoadInt32(&upstreamCalls))

	req3, err := http.NewRequest(http.MethodPost, "https://api.openrouteservice.org/v2/directions", strings.NewReader(`{"a":2}`))
	require.NoError(t, err)

	resp3, err := rt.RoundTrip(req3)
	require.NoError(t, err)
	body3, err := io.ReadAll(resp3.Body)
	require.NoError(t, err)
	resp3.Body.Close()

	assert.Equal(t, "route:{\"a\":2}", string(body3))
	assert.Equal(t, int32(2), atomic.LoadInt32(&upstreamCalls))
}
