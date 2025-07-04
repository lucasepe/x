package client

import (
	"os"
	"strconv"
)

func ConfigFromEnv() (res Config) {
	if v, ok := os.LookupEnv(serverUrlEnv); ok {
		res.ServerURL = string(v)
	}

	if v, ok := os.LookupEnv(proxyUrlEnv); ok {
		res.ProxyURL = string(v)
	}

	if v, ok := os.LookupEnv(tokenEnv); ok {
		res.Token = string(v)
	}

	if v, ok := os.LookupEnv(usernameEnv); ok {
		res.Username = string(v)
	}

	if v, ok := os.LookupEnv(passwordEnv); ok {
		res.Password = string(v)
	}

	if v, ok := os.LookupEnv(caEnv); ok {
		res.CertificateAuthorityData = string(v)
	}

	if v, ok := os.LookupEnv(clientKeyEnv); ok {
		res.ClientKeyData = string(v)
	}

	if v, ok := os.LookupEnv(clientCertEnv); ok {
		res.ClientCertificateData = string(v)
	}

	if v, ok := os.LookupEnv(verboseEnv); ok {
		res.Verbose, _ = strconv.ParseBool(string(v))
	}

	if v, ok := os.LookupEnv(insecureEnv); ok {
		res.Insecure, _ = strconv.ParseBool(string(v))
	}

	return res
}

type Config struct {
	ServerURL                string
	ProxyURL                 string
	CertificateAuthorityData string
	ClientCertificateData    string
	ClientKeyData            string
	Token                    string
	Username                 string
	Password                 string
	Verbose                  bool
	Insecure                 bool
}

// HasCA returns whether the configuration has a certificate authority or not.
func (c *Config) HasCA() bool {
	return len(c.CertificateAuthorityData) > 0
}

// HasBasicAuth returns whether the configuration has basic authentication or not.
func (c *Config) HasBasicAuth() bool {
	return len(c.Password) != 0
}

// HasTokenAuth returns whether the configuration has token authentication or not.
func (c *Config) HasTokenAuth() bool {
	return len(c.Token) != 0
}

// HasCertAuth returns whether the configuration has certificate authentication or not.
func (c *Config) HasCertAuth() bool {
	return len(c.ClientCertificateData) != 0 && len(c.ClientKeyData) != 0
}

const (
	serverUrlEnv  = "SERVER_URL"
	proxyUrlEnv   = "PROXY_URL"
	passwordEnv   = "PASSWORD"
	usernameEnv   = "USERNAME"
	tokenEnv      = "TOKEN"
	clientCertEnv = "CERT"
	clientKeyEnv  = "CERT_KEY"
	caEnv         = "CA_CERT"
	insecureEnv   = "INSECURE"
	verboseEnv    = "VERBOSE"
)
