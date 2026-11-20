package tlsproviders

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ritta/internal/config"
	"ritta/internal/logger"
)

type mockSSHRunner struct {
	sudoCalls       []string
	certExists      bool
}

func (m *mockSSHRunner) RunSudo(command string) error {
	m.sudoCalls = append(m.sudoCalls, command)
	if strings.HasPrefix(command, "test -f /etc/letsencrypt/live/") {
		if !m.certExists {
			return errors.New("file not found")
		}
		return nil
	}
	return nil
}

func TestLetsEncrypt_BootstrapWhenCertMissing(t *testing.T) {
	mock := &mockSSHRunner{certExists: false}
	log := logger.New(100)
	le := NewLetsEncrypt(log)

	cfg := &config.Config{
		Domains: []config.Domain{{Host: "app.example.com", TLS: true}},
		TLS: &config.TLS{Provider: "letsencrypt", Email: "test@example.com"},
	}

	err := le.EnsureCertificates(mock, cfg)
	if err != nil {
		t.Fatalf("EnsureCertificates failed: %v", err)
	}

	foundWebroot := false
	foundBootstrapSite := false
	foundBootstrapCleanup := false
	for _, call := range mock.sudoCalls {
		if strings.Contains(call, "certbot certonly --webroot -w /var/www/certbot") && strings.Contains(call, "app.example.com") {
			foundWebroot = true
		}
		if strings.Contains(call, "ritta-bootstrap-app.example.com.conf") && strings.Contains(call, "sites-enabled") && !strings.Contains(call, "rm -f") {
			foundBootstrapSite = true
		}
		if strings.Contains(call, "rm -f") && strings.Contains(call, "ritta-bootstrap-app.example.com.conf") {
			foundBootstrapCleanup = true
		}
	}
	if !foundWebroot {
		t.Errorf("expected certbot webroot command for app.example.com, got calls: %v", mock.sudoCalls)
	}
	if !foundBootstrapSite {
		t.Errorf("expected temporary HTTP bootstrap site to be enabled before certbot, got calls: %v", mock.sudoCalls)
	}
	if !foundBootstrapCleanup {
		t.Errorf("expected temporary HTTP bootstrap site to be cleaned up after certbot, got calls: %v", mock.sudoCalls)
	}
}

func TestLetsEncrypt_SkipWhenCertExists(t *testing.T) {
	mock := &mockSSHRunner{certExists: true}
	log := logger.New(100)
	le := NewLetsEncrypt(log)

	cfg := &config.Config{
		Domains: []config.Domain{{Host: "app.example.com", TLS: true}},
		TLS: &config.TLS{Provider: "letsencrypt", Email: "test@example.com"},
	}

	err := le.EnsureCertificates(mock, cfg)
	if err != nil {
		t.Fatalf("EnsureCertificates failed: %v", err)
	}

	for _, call := range mock.sudoCalls {
		if strings.Contains(call, "certbot certonly") {
			t.Errorf("expected certbot certonly NOT to be called when cert exists, but was called: %s", call)
		}
	}
}

func TestLetsEncrypt_DiscoversDomainsFromDeployNginxSitesAvailable(t *testing.T) {
	mock := &mockSSHRunner{certExists: false}
	log := logger.New(100)
	le := NewLetsEncrypt(log)

	tmpDir := t.TempDir()
	deployDir := filepath.Join(tmpDir, "deploy")
	sitesAvail := filepath.Join(deployDir, "nginx", "sites-available")
	if err := os.MkdirAll(sitesAvail, 0755); err != nil {
		t.Fatalf("mkdir sitesAvail: %v", err)
	}

	siteConf := `server {
    listen 443 ssl;
    server_name discovered.example.com;
    ssl_certificate /etc/letsencrypt/live/discovered.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/discovered.example.com/privkey.pem;
}`
	if err := os.WriteFile(filepath.Join(sitesAvail, "discovered.example.com.conf"), []byte(siteConf), 0644); err != nil {
		t.Fatalf("writing siteConf: %v", err)
	}

	cfg := &config.Config{
		DeployDir: deployDir,
		Domains:   nil, // no domains explicitly listed
		TLS:       &config.TLS{Provider: "letsencrypt", Email: "admin@example.com"},
	}

	err := le.EnsureCertificates(mock, cfg)
	if err != nil {
		t.Fatalf("EnsureCertificates failed: %v", err)
	}

	foundDiscovered := false
	for _, call := range mock.sudoCalls {
		if strings.Contains(call, "discovered.example.com") {
			foundDiscovered = true
			break
		}
	}
	if !foundDiscovered {
		t.Errorf("expected EnsureCertificates to discover and bootstrap discovered.example.com from sites-available")
	}
}
