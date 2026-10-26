package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"ritta/internal/config"
)

func TestLoadConfig_EnvDeployPrecedenceAndExpansion(t *testing.T) {
	tmpDir := t.TempDir()

	envDeployContent := "VPS_HOST=192.168.1.100\nVPS_USER=deployer\nVPS_PORT=2222\nVPS_KEY=/tmp/id_ed25519\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.deploy"), []byte(envDeployContent), 0644); err != nil {
		t.Fatalf("writing .env.deploy: %v", err)
	}

	yamlContent := `
remote_project_root: /srv/app
local_project_root: ./
deploy_dir: ./custom_deploy
server:
  host: ${VPS_HOST}
  user: ${VPS_USER}
  port: ${VPS_PORT}
  key: ${VPS_KEY}
domains:
  - host: app.example.com
    port: 3000
    tls: true
    env: production
`
	configPath := filepath.Join(tmpDir, "rittaConfig.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("writing rittaConfig.yaml: %v", err)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.DeployDir != "./custom_deploy" {
		t.Errorf("expected DeployDir ./custom_deploy, got %q", cfg.DeployDir)
	}
	if cfg.Server.Host != "192.168.1.100" {
		t.Errorf("expected Server.Host 192.168.1.100, got %q", cfg.Server.Host)
	}
	if cfg.Server.User != "deployer" {
		t.Errorf("expected Server.User deployer, got %q", cfg.Server.User)
	}
	if cfg.Server.Port != 2222 {
		t.Errorf("expected Server.Port 2222, got %d", cfg.Server.Port)
	}
	if cfg.Server.Key != "/tmp/id_ed25519" {
		t.Errorf("expected Server.Key /tmp/id_ed25519, got %q", cfg.Server.Key)
	}
	if len(cfg.Domains) != 1 || cfg.Domains[0].Env != "production" {
		t.Errorf("expected domain env 'production', got %+v", cfg.Domains)
	}
}

func TestLoadConfig_ServerFallbacks(t *testing.T) {
	tmpDir := t.TempDir()

	envDeployContent := "SERVER_HOST=10.0.0.5\nSERVER_USER=ubuntu\nSERVER_PORT=22\nSERVER_SSH_KEY_PATH=~/.ssh/test_key\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.deploy"), []byte(envDeployContent), 0644); err != nil {
		t.Fatalf("writing .env.deploy: %v", err)
	}

	yamlContent := `
remote_project_root: /srv/app
local_project_root: ./
`
	configPath := filepath.Join(tmpDir, "rittaConfig.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("writing rittaConfig.yaml: %v", err)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Server.Host != "10.0.0.5" {
		t.Errorf("expected fallback Host 10.0.0.5, got %q", cfg.Server.Host)
	}
	if cfg.Server.User != "ubuntu" {
		t.Errorf("expected fallback User ubuntu, got %q", cfg.Server.User)
	}
	if cfg.Server.Port != 22 {
		t.Errorf("expected fallback Port 22, got %d", cfg.Server.Port)
	}
	if cfg.Server.Key != "~/.ssh/test_key" {
		t.Errorf("expected fallback Key ~/.ssh/test_key, got %q", cfg.Server.Key)
	}
	if cfg.DeployDir != "./deploy" {
		t.Errorf("expected default DeployDir ./deploy, got %q", cfg.DeployDir)
	}
}
