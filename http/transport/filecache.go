package transport

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/lucasepe/x/env"
	"github.com/lucasepe/x/filecache"
	"github.com/lucasepe/x/log"
)

// CacheKeyFunc costruisce la chiave usata dal file cache a partire dalla
// request e dall'eventuale body gia' letto per la generazione della chiave.
type CacheKeyFunc func(req *http.Request, body []byte) (string, error)

// FileCacheOptions controlla quali richieste possono essere cacheate e come
// viene generata la chiave di cache.
type FileCacheOptions struct {
	Methods []string
	KeyFunc CacheKeyFunc
}

// FileCacheTransport restituisce un transport che cachea su filesystem le
// risposte GET di successo usando come chiave un digest di metodo e URL.
// Il comportamento e' "cache-first" per le URL gia' presenti: in quel caso il
// backend non viene contattato. Per le 304, se il contenuto e' disponibile nel
// file cache, viene restituita una risposta sintetica con body ricostruito.
func FileCacheTransport(fs *filecache.FileCacheFS, us http.RoundTripper) http.RoundTripper {
	return FileCacheTransportWithOptions(fs, us, FileCacheOptions{
		Methods: []string{http.MethodGet},
		KeyFunc: DigestCacheKeyMethodURL,
	})
}

// FileCacheTransportWithOptions estende [FileCacheTransport] consentendo al
// chiamante di scegliere esplicitamente i metodi ammessi (ad esempio GET+POST)
// e la strategia di generazione della chiave di cache.
func FileCacheTransportWithOptions(fs *filecache.FileCacheFS, us http.RoundTripper, opts FileCacheOptions) http.RoundTripper {
	if len(opts.Methods) == 0 {
		opts.Methods = []string{http.MethodGet}
	}
	if opts.KeyFunc == nil {
		opts.KeyFunc = DigestCacheKeyMethodURL
	}

	return &cacheRoundTripper{
		cache:    fs,
		upstream: us,
		methods:  normalizeMethods(opts.Methods),
		keyFunc:  opts.KeyFunc,
	}
}

// DigestCacheKeyMethodURL produce un digest stabile basato su metodo e URL.
func DigestCacheKeyMethodURL(req *http.Request, _ []byte) (string, error) {
	return digestParts(req.Method, req.URL.String()), nil
}

// DigestCacheKeyMethodURLBody produce un digest basato su metodo, URL e body.
// E' il default piu' utile quando si abilita esplicitamente il caching delle POST.
func DigestCacheKeyMethodURLBody(req *http.Request, body []byte) (string, error) {
	return digestParts(req.Method, req.URL.String(), string(body)), nil
}

// DigestCacheKeyMethodURLHeadersBody costruisce una CacheKeyFunc che include
// anche il valore degli header indicati, in ordine lessicografico normalizzato.
// E' utile quando la risposta varia per lingua, tenant o autorizzazione.
func DigestCacheKeyMethodURLHeadersBody(headerNames ...string) CacheKeyFunc {
	names := append([]string(nil), headerNames...)
	for i := range names {
		names[i] = http.CanonicalHeaderKey(names[i])
	}
	sort.Strings(names)

	return func(req *http.Request, body []byte) (string, error) {
		parts := []string{req.Method, req.URL.String()}
		for _, name := range names {
			values := append([]string(nil), req.Header.Values(name)...)
			sort.Strings(values)
			parts = append(parts, name)
			parts = append(parts, values...)
		}
		parts = append(parts, string(body))
		return digestParts(parts...), nil
	}
}

type cacheRoundTripper struct {
	cache    *filecache.FileCacheFS
	upstream http.RoundTripper
	methods  []string
	keyFunc  CacheKeyFunc
}

func (t *cacheRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.upstream == nil {
		// Fallback difensivo: se il caller non passa un upstream, usiamo quello base.
		t.upstream = Default()
	}

	if !slices.Contains(t.methods, req.Method) {
		return t.upstream.RoundTrip(req)
	}

	cacheKey, err := t.cacheKey(req)
	if err != nil {
		return nil, err
	}

	if t.cache != nil && !env.True("SKIP_CACHE") {
		if data, err := t.cache.Get(cacheKey); err == nil {
			log.D("cache hit",
				log.String("method", req.Method),
				log.String("url", req.URL.String()),
			)
			// Cache hit: costruiamo una response sintetica che emula una 200.
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK (from cache)",
				Body:       io.NopCloser(bytes.NewReader(data)),
				Request:    req,
				Header:     make(http.Header),
			}, nil
		}
	}

	// Cache miss: lasciamo lavorare il transport reale.
	log.D("cache missed",
		log.String("method", req.Method),
		log.String("url", req.URL.String()),
	)

	resp, err := t.upstream.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		// Per i successi 2xx persistiamo il body e lo reiniettiamo nella response.
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		if t.cache != nil {
			if err := t.cache.Put(cacheKey, body); err != nil {
				log.E("unable to put response body in cache",
					log.String("method", req.Method),
					log.String("url", req.URL.String()),
					log.Err("err", err),
				)
			}
		}

		resp.Body = io.NopCloser(bytes.NewReader(body))
		return resp, nil

	case resp.StatusCode == http.StatusNotModified:
		// Una 304 senza cache disponibile non e' riutilizzabile.
		if t.cache == nil {
			return resp, fmt.Errorf("file cache is nil")
		}

		data, err := t.cache.Get(cacheKey)
		if err != nil {
			// Se il body non e' presente, restituiamo la 304 originale.
			return resp, nil
		}

		resp.Body.Close()

		// Se invece il body esiste, promuoviamo la risposta a 200 usando il dato
		// persistito.
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK (from cache after 304)",
			Body:       io.NopCloser(bytes.NewReader(data)),
			Request:    req,
			Header:     make(http.Header),
		}, nil

	default:
		// Tutto il resto viene passato attraverso senza side effect.
		return resp, nil
	}
}

func (t *cacheRoundTripper) cacheKey(req *http.Request) (string, error) {
	if t.keyFunc == nil {
		return DigestCacheKeyMethodURL(req, nil)
	}

	body, err := requestBodyBytes(req)
	if err != nil {
		return "", err
	}
	return t.keyFunc(req, body)
}

func normalizeMethods(methods []string) []string {
	if len(methods) == 0 {
		return []string{http.MethodGet}
	}

	out := make([]string, 0, len(methods))
	seen := make(map[string]struct{}, len(methods))

	for _, method := range methods {
		normalized := strings.ToUpper(strings.TrimSpace(method))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	if len(out) == 0 {
		return []string{http.MethodGet}
	}

	return out
}

func requestBodyBytes(req *http.Request) ([]byte, error) {
	if req == nil || req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(body))
	if req.GetBody != nil {
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(body)), nil
		}
	}

	return body, nil
}

func digestParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
