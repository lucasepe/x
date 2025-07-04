package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/lucasepe/x/http/transport"
)

func HTTPClientForConfig(cfg Config) (*http.Client, error) {
	rt, err := tlsConfigFor(&cfg)
	if err != nil {
		return &http.Client{
			Transport: transport.Default(),
		}, err
	}

	if cfg.Verbose {
		log.Println("using verbose roundtripper")
		rt = transport.VerboseRoundTripper(rt)
	}

	// Set authentication wrappers
	switch {
	case cfg.HasBasicAuth() && cfg.HasTokenAuth():
		return nil, fmt.Errorf("username/password or bearer token may be set, but not both")

	case cfg.HasTokenAuth():
		if cfg.Verbose {
			log.Println("using bearer auth roundtripper")
		}
		rt = transport.BearerAuthRoundTripper(cfg.Token, rt)

	case cfg.HasBasicAuth():
		if cfg.Verbose {
			log.Println("using basic auth roundtripper")
		}
		rt = transport.BasicAuthRoundTripper(cfg.Username, cfg.Password, rt)
	}

	return &http.Client{Transport: rt}, nil
}

func tlsConfigFor(c *Config) (http.RoundTripper, error) {
	res := transport.Default()

	if c.ProxyURL != "" {
		u, err := parseProxyURL(c.ProxyURL)
		if err != nil {
			return nil, err
		}

		res.Proxy = http.ProxyURL(u)
	}

	var caCertPool *x509.CertPool
	if len(c.CertificateAuthorityData) > 0 {
		caData, err := base64.StdEncoding.DecodeString(c.CertificateAuthorityData)
		if err != nil {
			return nil, fmt.Errorf("unable to decode certificate authority data")
		}

		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caData)

		res.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: c.Insecure,
			RootCAs:            caCertPool,
		}
	}

	if !c.HasCertAuth() {
		return res, nil
	}

	certData, err := base64.StdEncoding.DecodeString(c.ClientCertificateData)
	if err != nil {
		return nil, fmt.Errorf("unable to decode client certificate data")
	}

	keyData, err := base64.StdEncoding.DecodeString(c.ClientKeyData)
	if err != nil {
		return nil, fmt.Errorf("unable to decode client key data")
	}

	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return res, err
	}

	if res.TLSClientConfig == nil {
		res.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: c.Insecure,
		}
	}
	res.TLSClientConfig.Certificates = []tls.Certificate{cert}

	return res, nil
}
