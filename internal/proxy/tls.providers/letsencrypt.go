package tlsproviders

import (
	"fmt"

	"ritta/internal/config"
	"ritta/internal/logger"
	rittaSSH "ritta/internal/ssh"
)

type LetsEncrypt struct {
	log *logger.Logger
}

func NewLetsEncrypt(log *logger.Logger) *LetsEncrypt {
	return &LetsEncrypt{log: log}
}

func (l *LetsEncrypt) Configure(client *rittaSSH.Client, cfg *config.Config) error {
	if cfg.TLS == nil {
		return nil
	}

	for _, domain := range cfg.Domains {
		if !domain.TLS {
			continue
		}

		if err := l.configureDomain(
			client,
			domain.Host,
			cfg.TLS.Email,
		); err != nil {
			return err
		}
	}

	return nil
}

func (l *LetsEncrypt) configureDomain(client *rittaSSH.Client, host string, email string) error {
	l.log.Infof("Requesting TLS certificate for %s...", host)

	command := fmt.Sprintf(
		"certbot --nginx "+
			"--non-interactive "+
			"--agree-tos "+
			"--email %q "+
			"-d %q "+
			"--redirect",
		email,
		host,
	)

	if err := client.RunSudo(command); err != nil {
		return fmt.Errorf("obtaining TLS certificate for %s: %w", host, err)
	}

	l.log.Successf(":) TLS enabled for %s", host)

	return nil
}
