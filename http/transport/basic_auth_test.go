package transport_test

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/lucasepe/x/http/transport"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuthRoundTripper(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("user:pass")), r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	})

	mockRT := &mockRoundTripper{mux: mux}

	client := transport.BasicAuthRoundTripper("user", "pass", mockRT)

	req, _ := http.NewRequest("GET", "https://httpbin.org", nil)
	_, err := client.RoundTrip(req)
	assert.NoError(t, err)
}
