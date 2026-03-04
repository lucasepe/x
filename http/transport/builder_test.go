package transport_test

import (
	"net/http"
	"testing"

	"github.com/lucasepe/x/http/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransportBuilderPreservesDeclarationOrder(t *testing.T) {
	steps := make([]string, 0, 2)

	builder := transport.NewTransportBuilder().
		Use(func(next http.RoundTripper) http.RoundTripper {
			return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				steps = append(steps, "outer")
				return next.RoundTrip(req)
			})
		}).
		Use(func(next http.RoundTripper) http.RoundTripper {
			return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				steps = append(steps, "inner")
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Status:     "204 No Content",
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

	assert.Equal(t, []string{"outer", "inner"}, steps)
}
