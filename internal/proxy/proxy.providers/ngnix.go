package proxyproviders

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ritta/internal/config"
	"ritta/internal/logger"
)

// fake for testing
type sshRunner interface {
	RunSudo(command string) error
	RunSudoWithStdin(command, stdin string) error
}

type Nginx struct {
	SSH sshRunner
	log *logger.Logger
}

func NewNginx(client sshRunner, log *logger.Logger) *Nginx {
	return &Nginx{
		SSH: client,
		log: log,
	}
}

func (n *Nginx) EnsureInstalled() error {
	if err := n.SSH.RunSudo("command -v nginx >/dev/null 2>&1 || [ -x /usr/sbin/nginx ]"); err == nil {
		return nil
	}
	return fmt.Errorf("nginx is not installed; install it in the setup script")
}

func (n *Nginx) ConfigureDomain(domain config.Domain) error {
	configContent := GenerateConfig(domain)
	path := fmt.Sprintf("/etc/nginx/conf.d/ritta-%s.conf", domain.Host)

	command := fmt.Sprintf("tee %s > /dev/null", path)
	if err := n.SSH.RunSudoWithStdin(command, configContent); err != nil {
		return fmt.Errorf("writing nginx configuration: %w", err)
	}

	n.log.Successf(":) %s to localhost:%d", domain.Host, domain.Port)
	return nil
}

func GenerateConfig(domain config.Domain) string {
	var builder strings.Builder

	builder.WriteString("server {\n")
	builder.WriteString("    listen 80;\n")
	builder.WriteString("    listen [::]:80;\n")
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("    server_name %s;\n", domain.Host))
	builder.WriteString("\n")
	builder.WriteString("    location / {\n")
	builder.WriteString(fmt.Sprintf("        proxy_pass http://127.0.0.1:%d;\n", domain.Port))
	builder.WriteString("        proxy_http_version 1.1;\n")
	builder.WriteString("        proxy_set_header Host $host;\n")
	builder.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
	builder.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
	builder.WriteString("        proxy_set_header X-Forwarded-Proto $scheme;\n")
	builder.WriteString("    }\n")
	builder.WriteString("}\n")

	return builder.String()
}

func (n *Nginx) Test() error {
	n.log.Info("Testing Nginx configuration...")

	if err := n.SSH.RunSudo("nginx -t"); err != nil {
		return fmt.Errorf("nginx configuration test failed: %w", err)
	}

	return nil
}

func (n *Nginx) Reload() error {
	if err := n.SSH.RunSudo("systemctl enable --now nginx"); err != nil {
		return err
	}
	return n.SSH.RunSudo("systemctl reload nginx")
}

func (n *Nginx) Configure(domains []config.Domain) error {
	if len(domains) == 0 {
		n.log.Info("No domains configured")
		return nil
	}

	n.log.Info("Configuring Nginx...")

	if err := n.EnsureInstalled(); err != nil {
		return err
	}

	for _, domain := range domains {
		if err := n.ConfigureDomain(domain); err != nil {
			return fmt.Errorf("configuring domain %s: %w", domain.Host, err)
		}
	}

	if err := n.Test(); err != nil {
		return err
	}

	if err := n.Reload(); err != nil {
		return fmt.Errorf("reloading nginx: %w", err)
	}

	n.log.Success(":) Nginx configured")

	return nil
}

func (n *Nginx) SyncNginx(deployDir string, domains []config.Domain, envFilter string) error {
	if err := n.EnsureInstalled(); err != nil {
		return err
	}

	n.log.Info("Preparing remote Nginx directories and archiving legacy configs...")
	preflightCmd := `mkdir -p /etc/nginx/conf.d /etc/nginx/sites-available /etc/nginx/sites-enabled /var/www/certbot /etc/nginx/conf.d/backup-ritta && ` +
		`chmod -R 755 /var/www/certbot && ` +
		`([ -f /etc/nginx/conf.d/00-limit-zones.conf ] && mv /etc/nginx/conf.d/00-limit-zones.conf /etc/nginx/conf.d/backup-ritta/ 2>/dev/null || true) && ` +
		`(for f in /etc/nginx/conf.d/ritta-*.conf; do [ -f "$f" ] && mv "$f" /etc/nginx/conf.d/backup-ritta/ 2>/dev/null || true; done)`

	if err := n.SSH.RunSudo(preflightCmd); err != nil {
		return fmt.Errorf("remote preflight: %w", err)
	}

	localConfD := filepath.Join(deployDir, "nginx", "conf.d")
	localSitesAvail := filepath.Join(deployDir, "nginx", "sites-available")

	// 1. Sync conf.d files
	if entries, err := os.ReadDir(localConfD); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".conf") {
				content, err := os.ReadFile(filepath.Join(localConfD, e.Name()))
				if err != nil {
					return fmt.Errorf("reading conf.d %s: %w", e.Name(), err)
				}
				dest := fmt.Sprintf("/etc/nginx/conf.d/%s", e.Name())
				if err := n.SSH.RunSudoWithStdin(fmt.Sprintf("tee %s > /dev/null", dest), string(content)); err != nil {
					return fmt.Errorf("uploading %s: %w", dest, err)
				}
				n.log.Infof("Uploaded %s -> %s", e.Name(), dest)
			}
		}
	}

	// 2. Sync sites-available files according to envFilter
	var enabledSites []string
	if entries, err := os.ReadDir(localSitesAvail); err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".conf") {
				continue
			}
			name := e.Name()
			isStaging := strings.Contains(name, "staging")

			if envFilter == "staging" && !isStaging {
				continue
			}
			if (envFilter == "prod" || envFilter == "production") && isStaging {
				continue
			}

			content, err := os.ReadFile(filepath.Join(localSitesAvail, name))
			if err != nil {
				return fmt.Errorf("reading site config %s: %w", name, err)
			}
			dest := fmt.Sprintf("/etc/nginx/sites-available/%s", name)
			if err := n.SSH.RunSudoWithStdin(fmt.Sprintf("tee %s > /dev/null", dest), string(content)); err != nil {
				return fmt.Errorf("uploading %s: %w", dest, err)
			}
			n.log.Infof("Uploaded %s -> %s", name, dest)
			enabledSites = append(enabledSites, name)
		}
	}

	// 3. Symlink to sites-enabled
	for _, site := range enabledSites {
		symlinkCmd := fmt.Sprintf("ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/%s", site, site)
		if err := n.SSH.RunSudo(symlinkCmd); err != nil {
			return fmt.Errorf("symlinking %s: %w", site, err)
		}
	}

	// 4. Test configuration
	if err := n.Test(); err != nil {
		// Rollback newly created symlinks on failure
		for _, site := range enabledSites {
			_ = n.SSH.RunSudo(fmt.Sprintf("rm -f /etc/nginx/sites-enabled/%s", site))
		}
		return fmt.Errorf("nginx validation failed, reverted symlinks: %w", err)
	}

	// 5. Reload Nginx
	if err := n.Reload(); err != nil {
		return fmt.Errorf("reloading nginx: %w", err)
	}

	n.log.Success(":) Nginx synchronized and reloaded successfully")
	return nil
}
