package transport

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

// RetryOptions controlla quando e come il transport deve ritentare una
// richiesta fallita.
type RetryOptions struct {
	// MaxAttempts include il tentativo iniziale. Valori <= 1 disabilitano i retry.
	MaxAttempts int
	// StatusCodes elenca i codici HTTP per cui e' ammesso un retry.
	StatusCodes []int
	// Methods limita i metodi ritentabili. Se vuoto usa GET, HEAD e OPTIONS.
	Methods []string
	// BaseDelay e' il ritardo usato quando la risposta non espone Retry-After.
	BaseDelay time.Duration
	// MaxDelay limita il ritardo massimo applicabile.
	MaxDelay time.Duration
	// RetryOnError abilita il retry anche sugli errori restituiti dal transport.
	RetryOnError bool
}

// RetryRoundTripper applica retry configurabili su status code specifici
// (tipicamente 429 o 5xx) e, opzionalmente, sugli errori del transport.
// I retry avvengono solo per metodi esplicitamente ammessi e solo se il body
// della request puo' essere ricostruito tramite GetBody o se non esiste.
func RetryRoundTripper(next http.RoundTripper, opts RetryOptions) http.RoundTripper {
	if next == nil {
		next = Default()
	}
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 1
	}
	if len(opts.Methods) == 0 {
		opts.Methods = []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	}
	if opts.BaseDelay <= 0 {
		opts.BaseDelay = 250 * time.Millisecond
	}
	if opts.MaxDelay <= 0 {
		opts.MaxDelay = 5 * time.Second
	}

	return &retryTransport{
		next: next,
		opts: RetryOptions{
			MaxAttempts:  opts.MaxAttempts,
			StatusCodes:  append([]int(nil), opts.StatusCodes...),
			Methods:      normalizeMethods(opts.Methods),
			BaseDelay:    opts.BaseDelay,
			MaxDelay:     opts.MaxDelay,
			RetryOnError: opts.RetryOnError,
		},
	}
}

type retryTransport struct {
	next http.RoundTripper
	opts RetryOptions
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return t.next.RoundTrip(req)
	}
	if t.opts.MaxAttempts <= 1 || !slices.Contains(t.opts.Methods, req.Method) {
		return t.next.RoundTrip(req)
	}
	if !canRetryRequest(req) {
		return t.next.RoundTrip(req)
	}

	var lastResp *http.Response
	var lastErr error

	for attempt := 1; attempt <= t.opts.MaxAttempts; attempt++ {
		currentReq, err := requestForAttempt(req, attempt)
		if err != nil {
			return nil, err
		}

		resp, err := t.next.RoundTrip(currentReq)
		if err != nil {
			lastErr = err
			if !t.opts.RetryOnError || attempt == t.opts.MaxAttempts {
				return nil, err
			}

			if err := waitForRetry(currentReq.Context(), t.opts.BaseDelay, t.opts.MaxDelay); err != nil {
				return nil, err
			}
			continue
		}

		lastResp = resp
		lastErr = nil

		if !slices.Contains(t.opts.StatusCodes, resp.StatusCode) || attempt == t.opts.MaxAttempts {
			return resp, nil
		}

		delay := retryDelay(resp, t.opts.BaseDelay, t.opts.MaxDelay)
		resp.Body.Close()
		if err := waitForRetry(currentReq.Context(), delay, t.opts.MaxDelay); err != nil {
			return nil, err
		}
	}

	if lastResp != nil {
		return lastResp, nil
	}
	return nil, lastErr
}

func canRetryRequest(req *http.Request) bool {
	return req.Body == nil || req.Body == http.NoBody || req.GetBody != nil
}

func requestForAttempt(req *http.Request, attempt int) (*http.Request, error) {
	if attempt == 1 {
		return req, nil
	}

	cloned := cloneRequest(req)
	if req.GetBody == nil || req.Body == nil || req.Body == http.NoBody {
		return cloned, nil
	}

	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	cloned.Body = body
	return cloned, nil
}

func retryDelay(resp *http.Response, baseDelay, maxDelay time.Duration) time.Duration {
	delay := parseRetryAfter(resp)
	if delay <= 0 {
		delay = baseDelay
	}
	if maxDelay > 0 && delay > maxDelay {
		return maxDelay
	}
	return delay
}

func parseRetryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}

	raw := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if raw == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	when, err := http.ParseTime(raw)
	if err != nil {
		return 0
	}

	delay := time.Until(when)
	if delay < 0 {
		return 0
	}
	return delay
}

func waitForRetry(ctx context.Context, delay, maxDelay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	if maxDelay > 0 && delay > maxDelay {
		delay = maxDelay
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
