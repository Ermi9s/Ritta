package config

import (
	"os"
	"path/filepath"
	"testing"
	"strings"
)

func TestCreateTemplate(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "deploy-target")

	// 1. First run creates both files
	if err := CreateTemplate(targetDir); err != nil {
		t.Fatalf("CreateTemplate failed on initial run: %v", err)
	}

	configPath := filepath.Join(targetDir, "rittaConfig.yaml")
	scriptPath := filepath.Join(targetDir, "rittaScript.sh")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("expected %s to be created", configPath)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Errorf("expected %s to be created", scriptPath)
	}

	// Verify generated config is valid syntax
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("generated config failed to load: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config loaded from generated template")
	}

	// 2. Second run does not fail or overwrite modified content
	customContent := "# Custom configuration\nlocal_project_root: /custom\n"
	if err := os.WriteFile(configPath, []byte(customContent), 0644); err != nil {
		t.Fatalf("failed writing custom content: %v", err)
	}

	if err := CreateTemplate(targetDir); err != nil {
		t.Fatalf("CreateTemplate failed on second run: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed reading config after re-run: %v", err)
	}
	if string(content) != customContent {
		t.Errorf("CreateTemplate overwrote existing config file")
	}
}

func TestCreateTemplate_ScaffoldsDeployDirectoryAndGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	err := CreateTemplate(tmpDir)
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(tmpDir, "rittaConfig.yaml"),
		filepath.Join(tmpDir, "rittaScript.sh"),
		filepath.Join(tmpDir, "deploy", "nginx", "conf.d", "global.conf"),
		filepath.Join(tmpDir, "deploy", "nginx", "sites-available", "app.example.com.conf"),
		filepath.Join(tmpDir, "deploy", ".env.deploy.example"),
		filepath.Join(tmpDir, ".gitignore"),
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("expected file %s to be created, but it was not", f)
		}
	}

	gitignoreContent, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
	if err != nil {
		t.Fatalf("reading .gitignore: %v", err)
	}
	if !strings.Contains(string(gitignoreContent), ".env.deploy") {
		t.Errorf("expected .gitignore to contain .env.deploy")
	}

	globalConf, err := os.ReadFile(filepath.Join(tmpDir, "deploy", "nginx", "conf.d", "global.conf"))
	if err != nil {
		t.Fatalf("reading global.conf: %v", err)
	}
	if strings.Contains(string(globalConf), "server_tokens") {
		t.Errorf("global.conf should not contain server_tokens directive to avoid duplicate directive collisions")
	}
}
