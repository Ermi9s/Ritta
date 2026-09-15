package proxyproviders

import (
	"fmt"
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
