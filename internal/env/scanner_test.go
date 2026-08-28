package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
	tempDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tempDir, ".env"), []byte("PORT=3000"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, ".env.production"), []byte("PORT=8080"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# App"), 0644)

	gitDir := filepath.Join(tempDir, ".git")
	_ = os.MkdirAll(gitDir, 0755)
	_ = os.WriteFile(filepath.Join(gitDir, ".env.secret"), []byte("SECRET=123"), 0644)

	nodeModules := filepath.Join(tempDir, "node_modules", "pkg")
	_ = os.MkdirAll(nodeModules, 0755)
	_ = os.WriteFile(filepath.Join(nodeModules, ".env"), []byte("IGNORED=1"), 0644)

	configDir := filepath.Join(tempDir, "config")
	_ = os.MkdirAll(configDir, 0755)
	_ = os.WriteFile(filepath.Join(configDir, ".env.local"), []byte("LOCAL=true"), 0644)

	files, err := Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	foundMap := make(map[string]bool)
	for _, f := range files {
		foundMap[f] = true
	}

	if !foundMap[".env"] {
		t.Errorf("expected .env to be found")
	}
	if !foundMap[".env.production"] {
		t.Errorf("expected .env.production to be found")
	}
	if !foundMap[filepath.Join("config", ".env.local")] {
		t.Errorf("expected config/.env.local to be found")
	}
	if foundMap[filepath.Join(".git", ".env.secret")] {
		t.Errorf("expected .git directory to be skipped")
	}
	if foundMap[filepath.Join("node_modules", "pkg", ".env")] {
		t.Errorf("expected node_modules directory to be skipped")
	}
}
