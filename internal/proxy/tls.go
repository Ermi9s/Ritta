package proxy

import (
	"fmt"
	"ritta/internal/config"
	tlsproviders "ritta/internal/proxy/tls.providers"
	rittaSSH "ritta/internal/ssh"
)


func NewTLSProvider(client *rittaSSH.Client, cfg *config.Config) (TLSProvider, error) {
	if cfg.TLS == nil {
		return nil, fmt.Errorf("TLS is not configured");
	}

	switch cfg.TLS.Provider {
	case "letsencrypt":
		return tlsproviders.NewLetsEncrypt(), nil;
	default:
		return nil, fmt.Errorf("unsupported TLS provider: %s", cfg.TLS.Provider);
	}
}