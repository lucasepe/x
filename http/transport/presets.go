package transport

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// DebugPresetOptions consente di personalizzare i layer applicati da
// [ForDebugWithOptions].
type DebugPresetOptions struct {
	RequestID RequestIDOptions
	Verbose   VerboseOptions
}

// ScrapingPresetOptions consente di personalizzare i layer applicati da
// [ForScrapingWithOptions].
type ScrapingPresetOptions struct {
	HostRate  rate.Limit
	HostBurst int
	Retry     RetryOptions
}

// ForDebug applica al builder i layer tipici per debugging e tracing locale:
// request id automatico e verbose logging.
// Se il builder e' nil, ne crea uno nuovo. Il verbose viene applicato per
// ultimo cosi' i log includono gia' l'header di correlazione.
//
// Esempio: aggiungere debug a un client JSON che fa POST verso una API
// "read-only" come OpenRouteService:
//
//	b := transport.ForDebug(
//		transport.NewTransportBuilder().
//			Use(func(next http.RoundTripper) http.RoundTripper {
//				return transport.FileCacheTransportWithOptions(cache, next, transport.FileCacheOptions{
//					Methods: []string{http.MethodGet, http.MethodPost},
//					KeyFunc: transport.DigestCacheKeyMethodURLBody,
//				})
//			}),
//	)
func ForDebug(b *TransportBuilder) *TransportBuilder {
	return ForDebugWithOptions(b, DebugPresetOptions{})
}

// ForDebugWithOptions e' la variante configurabile di [ForDebug].
// Il preset resta componibile: puoi applicarlo a un builder esistente oppure
// concatenarlo a [ForScraping] per abilitare logging e request id anche durante
// sviluppo di workflow di scraping.
func ForDebugWithOptions(b *TransportBuilder, opts DebugPresetOptions) *TransportBuilder {
	if b == nil {
		b = NewTransportBuilder()
	}

	return b.
		Use(func(next http.RoundTripper) http.RoundTripper {
			return RequestIDRoundTripperWithOptions(next, opts.RequestID)
		}).
		Use(func(next http.RoundTripper) http.RoundTripper {
			return VerboseRoundTripperWithOptions(next, opts.Verbose)
		})
}

// ForScraping applica al builder un preset prudente per scraping e crawling:
// limiting per host, retry su errori transitori e profilo browser stabile.
// Se il builder e' nil, ne crea uno nuovo.
//
// Esempio: aggiungere il debug sopra il preset di scraping:
//
//	b := transport.ForDebug(
//		transport.ForScraping(nil),
//	)
func ForScraping(b *TransportBuilder) *TransportBuilder {
	return ForScrapingWithOptions(b, ScrapingPresetOptions{})
}

// ForScrapingWithOptions e' la variante configurabile di [ForScraping].
// I default restano conservativi: 1 richiesta al secondo per host, retry su
// 429/502/503/504 e profilo browser sticky.
func ForScrapingWithOptions(b *TransportBuilder, opts ScrapingPresetOptions) *TransportBuilder {
	if b == nil {
		b = NewTransportBuilder()
	}
	if opts.HostRate <= 0 {
		opts.HostRate = rate.Every(time.Second)
	}
	if opts.HostBurst <= 0 {
		opts.HostBurst = 1
	}
	opts.Retry = withDefaultScrapingRetryOptions(opts.Retry)

	return b.
		Use(func(next http.RoundTripper) http.RoundTripper {
			return HostLimiter(opts.HostRate, opts.HostBurst, next)
		}).
		Use(func(next http.RoundTripper) http.RoundTripper {
			return RetryRoundTripper(next, opts.Retry)
		}).
		Use(func(next http.RoundTripper) http.RoundTripper {
			return StickyBrowserRoundTripper(next)
		})
}

func withDefaultScrapingRetryOptions(opts RetryOptions) RetryOptions {
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 3
	}
	if len(opts.StatusCodes) == 0 {
		opts.StatusCodes = []int{
			http.StatusTooManyRequests,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		}
	}
	if len(opts.Methods) == 0 {
		opts.Methods = []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	}
	if opts.BaseDelay <= 0 {
		opts.BaseDelay = 500 * time.Millisecond
	}
	if opts.MaxDelay <= 0 {
		opts.MaxDelay = 5 * time.Second
	}
	return opts
}
