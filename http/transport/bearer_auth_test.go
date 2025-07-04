package transport_test

import (
	"net/http"
	"testing"

	"github.com/lucasepe/x/http/transport"
	"github.com/stretchr/testify/assert"
)

func TestBearerAuthRoundTripper(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer mytoken", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	})

	mockRT := &mockRoundTripper{mux: mux}

	client := transport.BearerAuthRoundTripper("mytoken", mockRT)

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	_, err := client.RoundTrip(req)
	assert.NoError(t, err)
}
