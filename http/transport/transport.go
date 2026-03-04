package transport

import (
	"net"
	"net/http"
	"time"
)

// Default costruisce un [http.Transport] con parametri sensati per client CLI
// e servizi generici: keep-alive abilitato, supporto HTTP/2 e timeout
// conservativi sulle fasi piu' costose della connessione.
func Default() *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
