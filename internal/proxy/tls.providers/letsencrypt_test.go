package tlsproviders

import (
	"errors"
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
	for _, call := range mock.sudoCalls {
		if strings.Contains(call, "certbot certonly --webroot -w /var/www/certbot") && strings.Contains(call, "app.example.com") {
			foundWebroot = true
			break
		}
	}
	if !foundWebroot {
		t.Errorf("expected certbot webroot command for app.example.com, got calls: %v", mock.sudoCalls)
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
