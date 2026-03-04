package transport

import (
	"crypto/rand"
	"encoding/binary"
	"net/http"
	"strings"
	"time"
)

// RequestIDOptions controlla come viene valorizzato l'header di correlazione.
type RequestIDOptions struct {
	HeaderName string
	Generator  func() string
}

// RequestIDRoundTripper aggiunge un request id se l'header non e' gia' presente.
// Il comportamento di default usa `X-Request-Id` e un identificatore casuale in
// formato esadecimale.
func RequestIDRoundTripper(next http.RoundTripper) http.RoundTripper {
	return RequestIDRoundTripperWithOptions(next, RequestIDOptions{})
}

// RequestIDRoundTripperWithOptions consente di personalizzare nome header e
// strategia di generazione dell'identificatore.
func RequestIDRoundTripperWithOptions(next http.RoundTripper, opts RequestIDOptions) http.RoundTripper {
	if next == nil {
		next = Default()
	}
	if strings.TrimSpace(opts.HeaderName) == "" {
		opts.HeaderName = "X-Request-Id"
	}
	if opts.Generator == nil {
		opts.Generator = defaultRequestID
	}

	return &requestIDTransport{
		next:       next,
		headerName: http.CanonicalHeaderKey(opts.HeaderName),
		generator:  opts.Generator,
	}
}

type requestIDTransport struct {
	next       http.RoundTripper
	headerName string
	generator  func() string
}

func (t *requestIDTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return t.next.RoundTrip(req)
	}
	if req.Header.Get(t.headerName) != "" {
		return t.next.RoundTrip(req)
	}

	req = cloneRequest(req)
	req.Header.Set(t.headerName, t.generator())
	return t.next.RoundTrip(req)
}

func defaultRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fallback difensivo: in caso di errore usiamo comunque un valore non vuoto.
		binary.BigEndian.PutUint64(b[:8], uint64(time.Now().UnixNano()))
		binary.BigEndian.PutUint64(b[8:], uint64(time.Now().UnixNano()))
	}
	return hex16(b)
}

func hex16(b [16]byte) string {
	const hexdigits = "0123456789abcdef"

	out := make([]byte, 32)
	for i, v := range b {
		out[i*2] = hexdigits[v>>4]
		out[i*2+1] = hexdigits[v&0x0f]
	}
	return string(out)
}
