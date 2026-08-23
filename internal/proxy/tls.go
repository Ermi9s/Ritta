package proxy

import (
	"fmt"
	"ritta/internal/config"
	rittaSSH "ritta/internal/ssh"
	"ritta/internal/proxy/providers"
)

type TLSProvider interface {
	Configure(
		client *rittaSSH.Client,
		cfg *config.Config,
	) error
}

func ConfigureTLS(
	client *rittaSSH.Client,
	cfg *config.Config,
) error {
	if cfg.TLS == nil {
		return nil
	}

	switch cfg.TLS.Provider {
	case "letsencrypt":
		provider := providers.NewLetsEncrypt()

		return provider.Configure(client, cfg)

	default:
		return fmt.Errorf(
			"unsupported TLS provider: %s",
			cfg.TLS.Provider,
		)
	}
}