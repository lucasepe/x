package transport_test

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/lucasepe/x/http/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerboseRoundTripperWithOptionsTruncatesPreviewAndPreservesBody(t *testing.T) {
	originalStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = originalStderr
	})

	done := make(chan string, 1)
	go func() {
		var out bytes.Buffer
		_, _ = io.Copy(&out, r)
		done <- out.String()
	}()

	upstream := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"abcdefgh":"ijklmnop"}`)),
			Request:    req,
		}, nil
	})

	rt := transport.VerboseRoundTripperWithOptions(upstream, transport.VerboseOptions{
		LogBodies:    true,
		MaxBodyBytes: 8,
	})

	req, err := http.NewRequest(http.MethodPost, "https://example.com", strings.NewReader(`{"abcdefgh":"ijklmnop"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.NoError(t, w.Close())
	logged := <-done
	require.NoError(t, r.Close())

	assert.Equal(t, `{"abcdefgh":"ijklmnop"}`, string(body))
	assert.Contains(t, logged, "[body truncated to 8 bytes]")
}
