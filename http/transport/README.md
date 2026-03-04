# `transport`

Composable HTTP client transports built on top of `http.RoundTripper`.

This package provides small middleware-like wrappers that can be composed with
`TransportBuilder` to build reusable client stacks for:

- authentication
- retries
- rate limiting
- filesystem response caching
- sticky browser headers for scraping
- request/response debugging
- request correlation IDs

## Builder

`TransportBuilder` collects transport layers and applies them in declaration
order.

The first `Use(...)` call becomes the outermost layer and sees the request
first. The last `Use(...)` call stays closest to the base transport.

```go
b := transport.NewTransportBuilder().
    Use(layerA).
    Use(layerB)

rt := b.Build()
```

Execution order is:

1. `layerA`
2. `layerB`
3. `transport.Default()`

This is important when combining presets and custom layers.

## Implemented Transports

- `BasicAuthRoundTripper`: adds Basic auth only if `Authorization` is missing.
- `BearerAuthRoundTripper`: adds Bearer auth only if `Authorization` is missing.
- `FileCacheTransport`: caches configured request methods on filesystem.
- `FileCacheTransportWithOptions`: configurable file cache with explicit methods and cache keys.
- `HostLimiter`: per-host rate limiting.
- `RetryRoundTripper`: retries on configured status codes and optionally on transport errors.
- `RequestIDRoundTripper`: injects a request ID header if missing.
- `StickyBrowserRoundTripper`: keeps a stable browser profile per host.
- `VerboseRoundTripper`: logs request/response headers and a bounded body preview.
- `VerboseRoundTripperWithOptions`: configurable debug logging.

## Presets

- `ForDebug`: adds `RequestIDRoundTripper` and `VerboseRoundTripper`.
- `ForDebugWithOptions`: configurable debug preset.
- `ForScraping`: adds `HostLimiter`, `RetryRoundTripper`, and `StickyBrowserRoundTripper`.
- `ForScrapingWithOptions`: configurable scraping preset.

Presets are regular builder helpers. They do not replace the builder; they just
append layers to it, so they can be chained freely.

## Composition Examples

### JSON POST API with cache and debug

Useful for "read-only" POST APIs such as route calculation endpoints.

```go
b := transport.ForDebugWithOptions(
    transport.NewTransportBuilder().
        Use(func(next http.RoundTripper) http.RoundTripper {
            return transport.FileCacheTransportWithOptions(cache, next, transport.FileCacheOptions{
                Methods: []string{http.MethodGet, http.MethodPost},
                KeyFunc: transport.DigestCacheKeyMethodURLBody,
            })
        }).
        Use(func(next http.RoundTripper) http.RoundTripper {
            return transport.RetryRoundTripper(next, transport.RetryOptions{
                MaxAttempts: 3,
                StatusCodes: []int{
                    http.StatusTooManyRequests,
                    http.StatusBadGateway,
                    http.StatusServiceUnavailable,
                    http.StatusGatewayTimeout,
                },
                Methods:   []string{http.MethodGet, http.MethodPost},
                BaseDelay: 500 * time.Millisecond,
                MaxDelay:  5 * time.Second,
            })
        }),
    transport.DebugPresetOptions{
        Verbose: transport.VerboseOptions{
            LogBodies:    true,
            MaxBodyBytes: 8 * 1024,
        },
    },
)

client := &http.Client{Transport: b.Build()}
```

Order seen by the request:

1. debug preset (`RequestID`, then `Verbose`)
2. file cache
3. retry
4. base transport

### Add debug on top of scraping

```go
b := transport.ForDebug(
    transport.ForScraping(nil),
)

client := &http.Client{Transport: b.Build()}
```

This keeps scraping protections active while making requests easy to inspect
during development.

## Cache Key Strategies

When using `FileCacheTransportWithOptions`, the built-in helpers cover the most
common cases:

- `DigestCacheKeyMethodURL`
- `DigestCacheKeyMethodURLBody`
- `DigestCacheKeyMethodURLHeadersBody(...)`

For POST-based query APIs, `DigestCacheKeyMethodURLBody` is usually the minimum
safe choice.

## Retry Notes

`RetryRoundTripper` retries only when the request can be replayed safely:

- no body, or
- `req.GetBody` is available

This avoids retrying one-shot streamed bodies.
