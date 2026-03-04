package transport

import (
	"net/http"
)

// cloneRequest crea una copia shallow della request e una copia deep degli
// header, cosi' i middleware possono modificarli senza effetti collaterali.
func cloneRequest(req *http.Request) *http.Request {
	r := new(http.Request)

	// Copia strutturale della request.
	*r = *req

	// Copia profonda della mappa header.
	r.Header = cloneHeader(req.Header)

	return r
}

// cloneHeader restituisce una copia profonda di un [http.Header].
func cloneHeader(in http.Header) http.Header {
	out := make(http.Header, len(in))
	for key, values := range in {
		newValues := make([]string, len(values))
		copy(newValues, values)
		out[key] = newValues
	}
	return out
}
