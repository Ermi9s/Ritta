package proxy

import (
	"fmt"
	"strings"

	"ritta/internal/config"
	tlsproviders "ritta/internal/proxy/tls.providers"
	rittaSSH "ritta/internal/ssh"
)

func NewTLSProvider(client *rittaSSH.Client, cfg *config.Config) (TLSProvider, error) {
	if cfg.TLS == nil || cfg.TLS.Provider == "" {
		return nil, fmt.Errorf("TLS is not configured")
	}

	switch strings.ToLower(cfg.TLS.Provider) {
	case "letsencrypt", "certbot":
		return tlsproviders.NewLetsEncrypt(), nil
	default:
		return nil, fmt.Errorf("unsupported TLS provider: %s", cfg.TLS.Provider)
	}
}