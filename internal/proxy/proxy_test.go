package proxy

import (
	"strings"
	"testing"

	"ritta/internal/config"
	proxyproviders "ritta/internal/proxy/proxy.providers"
)

func TestNewReverseProxy(t *testing.T) {
	tests := []struct {
		provider string
		wantNil  bool
	}{
		{"nginx", false},
		{"Nginx", false},
		{"NGINX", false},
		{"apache", true},
		{"", true},
	}

	for _, tt := range tests {
		cfg := &config.Config{
			Proxy: &config.Proxy{Provider: tt.provider},
		}
		res := NewReverseProxy(nil, cfg)
		if (res == nil) != tt.wantNil {
			t.Errorf("NewReverseProxy(%q) nil = %v, wantNil %v", tt.provider, res == nil, tt.wantNil)
		}
	}

	// Test nil proxy config
	if res := NewReverseProxy(nil, &config.Config{}); res != nil {
		t.Errorf("expected nil for empty proxy config, got %v", res)
	}
}

func TestNewTLSProvider(t *testing.T) {
	tests := []struct {
		provider string
		wantErr  bool
	}{
		{"letsencrypt", false},
		{"LetsEncrypt", false},
		{"certbot", false},
		{"Certbot", false},
		{"unknown", true},
		{"", true},
	}

	for _, tt := range tests {
		cfg := &config.Config{
			TLS: &config.TLS{Provider: tt.provider},
		}
		_, err := NewTLSProvider(nil, cfg)
		if (err != nil) != tt.wantErr {
			t.Errorf("NewTLSProvider(%q) err = %v, wantErr %v", tt.provider, err, tt.wantErr)
		}
	}
}

func TestGenerateNginxConfig(t *testing.T) {
	domain := config.Domain{
		Host: "example.com",
		Port: 8080,
	}

	conf := proxyproviders.GenerateConfig(domain)

	if !strings.Contains(conf, "server_name example.com;") {
		t.Errorf("expected server_name example.com, got:\n%s", conf)
	}
	if !strings.Contains(conf, "proxy_pass http://127.0.0.1:8080;") {
		t.Errorf("expected proxy_pass http://127.0.0.1:8080, got:\n%s", conf)
	}
	if !strings.Contains(conf, "listen 80;") {
		t.Errorf("expected listen 80, got:\n%s", conf)
	}
}
