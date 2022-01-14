package xtls

import (
	"crypto/tls"
)

type PKCS12Config struct {
	Path     string
	Password string
}

type ClientOptions struct {
	PKCS12        PKCS12Config `json:"pkcs12" yaml:"pkcs12"`
	TLSSkipVerify bool         `json:"-" yaml:"-"`
	TLSMinVersion uint16       `json:"-" yaml:"-"`
	TLSMaxVersion uint16       `json:"-" yaml:"-"`
}

func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		TLSSkipVerify: true,
		TLSMinVersion: tls.VersionSSL30, // nolint[:staticcheck]
		TLSMaxVersion: tls.VersionTLS13,
	}
}
