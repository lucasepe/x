package transport

import (
	"fmt"
	"net/http"
)

// BearerAuthRoundTripper aggiunge l'header `Authorization: Bearer <token>` se
// la richiesta non contiene gia' un header di autorizzazione.
// Anche in questo caso la request viene clonata per evitare side effect.
func BearerAuthRoundTripper(token string, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = Default()
	}

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
		// Non sovrascriviamo credenziali impostate a monte.
		return rt.next.RoundTrip(req)
	}

	req = cloneRequest(req)
	token := rt.bearer

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return rt.next.RoundTrip(req)
}
