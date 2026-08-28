package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Valid(t *testing.T) {
	tempDir := t.TempDir()
	configContent := `
local_project_root: "./myapp"
remote_project_root: "/opt/myapp"
source:
  type: git
  repository: "https://github.com/test/app.git"
  branch: main
server:
  host: "192.168.1.100"
  user: "deploy"
  port: 2222
  key: "~/.ssh/id_rsa"
setup_config:
  script: "./setup.sh"
file:
  - from: ".env.production"
    to: ".env"
build:
  command: "go build"
run:
  command: "./app"
health:
  command: "curl -sf http://localhost:8080/health"
proxy:
  provider: "nginx"
domains:
  - host: "app.example.com"
    port: 8080
    tls: true
tls:
  provider: "letsencrypt"
  email: "admin@example.com"
`
	configPath := filepath.Join(tempDir, "rittaConfig.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.LocalProjectRoot != "./myapp" {
		t.Errorf("expected local_project_root './myapp', got %q", cfg.LocalProjectRoot)
	}
	if cfg.RemoteProjectRoot != "/opt/myapp" {
		t.Errorf("expected remote_project_root '/opt/myapp', got %q", cfg.RemoteProjectRoot)
	}
	if len(cfg.File) != 1 || cfg.File[0].From != ".env.production" || cfg.File[0].To != ".env" {
		t.Errorf("expected file entry to be parsed, got %+v", cfg.File)
	}
	if len(cfg.Domains) != 1 || cfg.Domains[0].Host != "app.example.com" || cfg.Domains[0].Port != 8080 {
		t.Errorf("expected domain entry to be parsed, got %+v", cfg.Domains)
	}
	if cfg.Server.Port != 2222 {
		t.Errorf("expected server port 2222, got %d", cfg.Server.Port)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(cfg *Config)
		wantErr bool
	}{
		{
			name:    "nil config",
			modify:  nil,
			wantErr: true,
		},
		{
			name: "valid config",
			modify: func(cfg *Config) {
				// baseline is valid
			},
			wantErr: false,
		},
		{
			name: "missing remote project root",
			modify: func(cfg *Config) {
				cfg.RemoteProjectRoot = ""
			},
			wantErr: true,
		},
		{
			name: "missing local project root",
			modify: func(cfg *Config) {
				cfg.LocalProjectRoot = ""
			},
			wantErr: true,
		},
		{
			name: "missing server host",
			modify: func(cfg *Config) {
				cfg.Server.Host = ""
			},
			wantErr: true,
		},
		{
			name: "missing server user",
			modify: func(cfg *Config) {
				cfg.Server.User = ""
			},
			wantErr: true,
		},
		{
			name: "invalid server port",
			modify: func(cfg *Config) {
				cfg.Server.Port = 70000
			},
			wantErr: true,
		},
		{
			name: "git source missing repository",
			modify: func(cfg *Config) {
				cfg.Source.Type = "git"
				cfg.Source.Repository = ""
			},
			wantErr: true,
		},
		{
			name: "unsupported source type",
			modify: func(cfg *Config) {
				cfg.Source.Type = "svn"
			},
			wantErr: true,
		},
		{
			name: "domain with invalid port",
			modify: func(cfg *Config) {
				cfg.Domains = []Domain{{Host: "example.com", Port: 0}}
			},
			wantErr: true,
		},
		{
			name: "domain with empty host",
			modify: func(cfg *Config) {
				cfg.Domains = []Domain{{Host: "", Port: 3000}}
			},
			wantErr: true,
		},
		{
			name: "unsupported proxy provider",
			modify: func(cfg *Config) {
				cfg.Proxy = &Proxy{Provider: "traefik"}
			},
			wantErr: true,
		},
		{
			name: "unsupported tls provider",
			modify: func(cfg *Config) {
				cfg.TLS = &TLS{Provider: "cloudflare"}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg *Config
			if tt.modify != nil {
				cfg = &Config{
					LocalProjectRoot:  "./",
					RemoteProjectRoot: "/opt/app",
					Source:            Source{Type: "git", Repository: "git@github.com:test/app.git"},
					Server:            Server{Host: "example.com", User: "ubuntu", Port: 22},
				}
				tt.modify(cfg)
			}

			err := Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
