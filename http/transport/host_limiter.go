package transport

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// HostLimiter applica un rate limit distinto per ogni host contattato.
// Ogni host riceve il proprio limiter, creato lazy al primo utilizzo, cosi'
// domini diversi non si influenzano tra loro.
func HostLimiter(rl rate.Limit, burst int, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = Default()
	}

	return &hostLimiterTransport{
		next:     next,
		limiters: make(map[string]*rate.Limiter),
		rate:     rl,
		burst:    burst,
	}
}

type hostLimiterTransport struct {
	next http.RoundTripper

	mu       sync.Mutex
	limiters map[string]*rate.Limiter

	rate  rate.Limit
	burst int
}

func (t *hostLimiterTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	limiter := t.limiterFor(req.URL.Host)
	// Wait rispetta il context della request: timeout e cancellazioni vengono
	// propagati al chiamante.
	if err := limiter.Wait(req.Context()); err != nil {
		return nil, err
	}
	return t.next.RoundTrip(req)
}

func (t *hostLimiterTransport) limiterFor(host string) *rate.Limiter {
	t.mu.Lock()
	defer t.mu.Unlock()

	if l, ok := t.limiters[host]; ok {
		// Reimpieghiamo sempre lo stesso limiter per garantire fairness per host.
		return l
	}

	l := rate.NewLimiter(t.rate, t.burst)
	t.limiters[host] = l
	return l
}
