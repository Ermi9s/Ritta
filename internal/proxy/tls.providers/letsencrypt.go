package tlsproviders

import (
	"fmt"

	"ritta/internal/config"
	rittaSSH "ritta/internal/ssh"
)

type LetsEncrypt struct{}

func NewLetsEncrypt() *LetsEncrypt {
	return &LetsEncrypt{}
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
	fmt.Printf("Requesting TLS certificate for %s...\n", host)

	command := fmt.Sprintf(
		"sudo certbot --nginx "+
			"--non-interactive "+
			"--agree-tos "+
			"--email %q "+
			"-d %q "+
			"--redirect",
		email,
		host,
	)

	if err := client.Run(command); err != nil {
		return fmt.Errorf("obtaining TLS certificate for %s: %w", host, err)
	}

	fmt.Printf(":) TLS enabled for %s\n", host)

	return nil
}