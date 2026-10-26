package tlsproviders

import (
	"fmt"

	"ritta/internal/config"
	"ritta/internal/logger"
	rittaSSH "ritta/internal/ssh"
)

type sshRunner interface {
	RunSudo(command string) error
}

type LetsEncrypt struct {
	log *logger.Logger
}

func NewLetsEncrypt(log *logger.Logger) *LetsEncrypt {
	return &LetsEncrypt{log: log}
}

func (l *LetsEncrypt) Configure(client *rittaSSH.Client, cfg *config.Config) error {
	return l.EnsureCertificates(client, cfg)
}

func (l *LetsEncrypt) EnsureCertificates(client sshRunner, cfg *config.Config) error {
	if cfg.TLS == nil {
		return nil
	}

	for _, domain := range cfg.Domains {
		if !domain.TLS {
			continue
		}

		if err := l.ensureDomainCertificate(client, domain.Host, cfg.TLS.Email); err != nil {
			return err
		}
	}

	return nil
}

func (l *LetsEncrypt) ensureDomainCertificate(client sshRunner, host string, email string) error {
	checkCmd := fmt.Sprintf("test -f /etc/letsencrypt/live/%s/fullchain.pem", host)
	if err := client.RunSudo(checkCmd); err == nil {
		l.log.Infof("Certificate for %s already exists, skipping issue.", host)
		return nil
	}

	l.log.Infof("Bootstrapping TLS certificate for %s via webroot...", host)

	if err := client.RunSudo("mkdir -p /var/www/certbot && chmod -R 755 /var/www/certbot"); err != nil {
		return fmt.Errorf("preparing webroot directory: %w", err)
	}

	var emailFlag string
	if email != "" {
		emailFlag = fmt.Sprintf("--email %q", email)
	} else {
		emailFlag = "--register-unsafely-without-email"
	}

	command := fmt.Sprintf(
		"certbot certonly --webroot -w /var/www/certbot "+
			"--non-interactive "+
			"--agree-tos "+
			"%s "+
			"-d %q",
		emailFlag,
		host,
	)

	if err := client.RunSudo(command); err != nil {
		// Fallback to certbot --nginx if webroot standalone challenges fail
		fallbackCmd := fmt.Sprintf(
			"certbot --nginx --non-interactive --agree-tos %s -d %q",
			emailFlag,
			host,
		)
		if fallbackErr := client.RunSudo(fallbackCmd); fallbackErr != nil {
			return fmt.Errorf("obtaining TLS certificate for %s: %w (fallback error: %v)", host, err, fallbackErr)
		}
	}

	l.log.Successf(":) TLS certificate ready for %s", host)
	return nil
}
