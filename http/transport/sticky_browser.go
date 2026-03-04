package transport

import (
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// StickyBrowserRoundTripper applica a ogni host un profilo browser coerente e
// persistente per tutta la vita del transport.
// Il primo accesso a un host sceglie casualmente un profilo desktop; le
// richieste successive verso lo stesso host riutilizzano sempre lo stesso set
// di header per simulare un client "stabile".
func StickyBrowserRoundTripper(upstream http.RoundTripper) http.RoundTripper {
	if upstream == nil {
		upstream = Default()
	}

	return &stickyBrowserTransport{
		upstream: upstream,
		perHost:  make(map[string]browserProfile),
		rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

type browserProfile struct {
	UserAgent       string
	Accept          string
	AcceptLanguage  string
	SecCHUA         string
	SecCHUAPlatform string
}

var (
	desktopProfiles = []browserProfile{
		// ---------- Chrome macOS ----------
		{
			UserAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
			Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			AcceptLanguage:  "it-IT,it;q=0.9,en-US;q=0.8,en;q=0.7",
			SecCHUA:         `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`,
			SecCHUAPlatform: `"macOS"`,
		},

		// ---------- Chrome Windows ----------
		{
			UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
			Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			AcceptLanguage:  "it-IT,it;q=0.9,en-US;q=0.8,en;q=0.7",
			SecCHUA:         `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`,
			SecCHUAPlatform: `"Windows"`,
		},

		// ---------- Firefox macOS ----------
		{
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:125.0) Gecko/20100101 Firefox/125.0",
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "it-IT,it;q=0.9,en-US;q=0.8,en;q=0.7",
		},

		// ---------- Firefox Windows ----------
		{
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0",
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "it-IT,it;q=0.9,en-US;q=0.8,en;q=0.7",
		},

		// ---------- Safari macOS ----------
		{
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15",
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "it-IT,it;q=0.9,en-US;q=0.8,en;q=0.7",
		},
	}
)

type stickyBrowserTransport struct {
	upstream http.RoundTripper

	mu      sync.Mutex
	perHost map[string]browserProfile
	rnd     *rand.Rand
}

func (h *stickyBrowserTransport) profileForHost(host string) browserProfile {
	h.mu.Lock()
	defer h.mu.Unlock()

	if profile, ok := h.perHost[host]; ok {
		return profile
	}

	profile := desktopProfiles[h.rnd.Intn(len(desktopProfiles))]
	h.perHost[host] = profile
	return profile
}

func (h *stickyBrowserTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	profile := h.profileForHost(req.URL.Host)

	// Cloniamo la request prima di mutare gli header, cosi' il chiamante puo'
	// riusare la request senza side effect.
	req = cloneRequest(req)

	req.Header.Set("User-Agent", profile.UserAgent)
	req.Header.Set("Accept", profile.Accept)
	req.Header.Set("Accept-Language", profile.AcceptLanguage)

	// Gli header client hints vengono inviati solo dai profili Chromium. Negli
	// altri casi li rimuoviamo esplicitamente per evitare fingerprint misti.
	if profile.SecCHUA != "" {
		req.Header.Set("Sec-CH-UA", profile.SecCHUA)
		req.Header.Set("Sec-CH-UA-Mobile", "?0")
		req.Header.Set("Sec-CH-UA-Platform", profile.SecCHUAPlatform)
	} else {
		req.Header.Del("Sec-CH-UA")
		req.Header.Del("Sec-CH-UA-Mobile")
		req.Header.Del("Sec-CH-UA-Platform")
	}

	return h.upstream.RoundTrip(req)
}
