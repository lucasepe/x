package transport

import "net/http"

// BasicAuthRoundTripper aggiunge l'header `Authorization: Basic ...` solo
// quando la richiesta non ne possiede gia' uno.
// Clona la request per non mutare accidentalmente l'istanza originale del chiamante.
func BasicAuthRoundTripper(user, pass string, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = Default()
	}

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
		// Se il chiamante ha gia' impostato un header esplicito, lo rispettiamo.
		return rt.next.RoundTrip(req)
	}
	req = cloneRequest(req)
	req.SetBasicAuth(rt.username, rt.password)
	return rt.next.RoundTrip(req)
}
