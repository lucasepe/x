package transport

import "net/http"

func BasicAuthRoundTripper(user, pass string, next http.RoundTripper) http.RoundTripper {
	return &basicAuthRoundTripper{
		username: user,
		password: pass,
		next:     next,
	}
}

type basicAuthRoundTripper struct {
	username string
	password string `datapolicy:"password"`
	next     http.RoundTripper
}

func (rt *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) != 0 {
		return rt.next.RoundTrip(req)
	}
	req = cloneRequest(req)
	req.SetBasicAuth(rt.username, rt.password)
	return rt.next.RoundTrip(req)
}
