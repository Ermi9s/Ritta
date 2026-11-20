package tlsproviders

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

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

	hostsMap := make(map[string]bool)
	for _, domain := range cfg.Domains {
		if domain.TLS && domain.Host != "" {
			hostsMap[domain.Host] = true
		}
	}

	// Discover domains requiring SSL certificates from deploy/nginx/sites-available
	deployDir := cfg.DeployDir
	if deployDir == "" {
		deployDir = "./deploy"
	}
	sitesAvail := filepath.Join(deployDir, "nginx", "sites-available")
	if entries, err := os.ReadDir(sitesAvail); err == nil {
		certRegex := regexp.MustCompile(`/etc/letsencrypt/live/([^/\s"';]+)/`)
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".conf") {
				if content, err := os.ReadFile(filepath.Join(sitesAvail, e.Name())); err == nil {
					matches := certRegex.FindAllStringSubmatch(string(content), -1)
					for _, m := range matches {
						if len(m) > 1 {
							hostsMap[m[1]] = true
						}
					}
				}
			}
		}
	}

	var hosts []string
	for host := range hostsMap {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)

	for _, host := range hosts {
		if err := l.ensureDomainCertificate(client, host, cfg.TLS.Email); err != nil {
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

	// Ensure Nginx is enabled and running
	_ = client.RunSudo("command -v nginx >/dev/null 2>&1 || [ -x /usr/sbin/nginx ]")
	_ = client.RunSudo("systemctl enable --now nginx || true")

	// 1. Install temporary HTTP-only bootstrap config so Nginx can respond to the ACME challenge
	bootstrapSite := fmt.Sprintf("/etc/nginx/sites-available/ritta-bootstrap-%s.conf", host)
	bootstrapEnabled := fmt.Sprintf("/etc/nginx/sites-enabled/ritta-bootstrap-%s.conf", host)
	bootstrapContent := fmt.Sprintf(`server {
    listen 80;
    listen [::]:80;
    server_name %s;
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    location / {
        return 200 "bootstrapping certificate\n";
    }
}
`, host)

	writeCmd := fmt.Sprintf("tee %s > /dev/null", bootstrapSite)
	if stdinRunner, ok := client.(interface{ RunSudoWithStdin(string, string) error }); ok {
		_ = stdinRunner.RunSudoWithStdin(writeCmd, bootstrapContent)
	} else {
		_ = client.RunSudo(fmt.Sprintf("echo %q > %s", bootstrapContent, bootstrapSite))
	}
	_ = client.RunSudo(fmt.Sprintf("ln -sf %s %s", bootstrapSite, bootstrapEnabled))
	_ = client.RunSudo("systemctl reload nginx || systemctl restart nginx || true")

	defer func() {
		_ = client.RunSudo(fmt.Sprintf("rm -f %s %s", bootstrapEnabled, bootstrapSite))
		_ = client.RunSudo("systemctl reload nginx || true")
	}()

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
