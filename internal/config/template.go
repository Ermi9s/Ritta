package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const DefaultConfigFile = "rittaConfig.yaml"  

const setupScriptTemplate = `#!/usr/bin/env bash

set -e

echo "Installing Ritta dependencies..."

sudo apt-get update

sudo apt-get install -y \
    git \
    nginx \
    certbot \
    python3-certbot-nginx

echo "Ritta dependencies installed."

echo
echo "Add application-specific dependencies below."
echo

# Example:
#
# sudo apt-get install -y docker.io
# sudo systemctl enable --now docker

echo "Setup complete."
`

const setupScriptName = "rittaScript.sh"

func CreateTemplate(path string) error {
	path = filepath.Clean(path)

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("configuration already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking configuration path: %w", err)
	}

	configDir := filepath.Dir(path)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating configuration directory: %w", err)
	}

	setupScriptPath := filepath.Join(configDir, setupScriptName)

	if _, err := os.Stat(setupScriptPath); err == nil {
		return fmt.Errorf("setup script already exists: %s", setupScriptPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking setup script: %w", err)
	}

	cfg := Config{
		RemoteProjectRoot: "~/srv/",
		LocalProjectRoot:  "./",

		Server: Server{
			Host: "",
			User: "deploy",
			Key:  "~/.ssh/id_ed25519",
			Port: 22,
		},

		SetupConfig: SetupConfig{
			PackageManager: "apt",
			Script:         "./rittaScript.sh",
		},

		Source: Source{
			Type:       "existing",
			Repository: "",
			Branch:     "",
		},

		ScanEnv: true,

		File: []File{},

		Build: &Command{
			Command: "",
		},

		Run: &Command{
			Command: "",
		},

		Health: &Health{
			Command: "",
		},

		Domains: []Domain{},

		Proxy: &Proxy{
			Provider: "Nginx",
		},

		TLS: &TLS{
			Provider: "letsencrypt",
			Email:    "",
		},
	}

	// Generate yaml
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshalling configuration: %w", err)
	}

	if err := os.WriteFile(path, data, 0644,); err != nil {
		return fmt.Errorf("writing configuration: %w", err)
	}

	// Generate rittaScript.sh.
	if err := os.WriteFile(setupScriptPath, []byte(setupScriptTemplate), 0755); err != nil {
		_ = os.Remove(path)

		return fmt.Errorf("writing setup script: %w", err)
	}
	return nil
}