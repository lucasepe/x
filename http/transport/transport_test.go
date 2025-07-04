package transport_test

import (
	"net/http"
	"net/http/httptest"
)

type mockRoundTripper struct {
	mux *http.ServeMux
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rr := httptest.NewRecorder()
	m.mux.ServeHTTP(rr, req)

	return rr.Result(), nil
}
