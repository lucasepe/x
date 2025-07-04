package transport

import (
	"fmt"
	"net/http"
)

func BearerAuthRoundTripper(token string, next http.RoundTripper) http.RoundTripper {
	return &bearerAuthRoundTripper{
		bearer: token,
		next:   next,
	}
}

type bearerAuthRoundTripper struct {
	bearer string
	next   http.RoundTripper
}

func (rt *bearerAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) != 0 {
		return rt.next.RoundTrip(req)
	}

	req = cloneRequest(req)
	token := rt.bearer

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return rt.next.RoundTrip(req)
}
