package config

import (
	"fmt"
	"os"
	"path/filepath"
	"ritta/internal/ui"
)

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
const filename = "rittaConfig.yaml"

func CreateTemplate(path string) error {
	path = filepath.Clean(path)

	configPath := filepath.Join(path, filename)
	scriptPath := filepath.Join(path, setupScriptName)
	configExists := false
	scriptExists := false

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("%s", ui.ErrorStyle.Render(fmt.Sprintf("creating configuration directory: %v", err)))
	}

	if _, err := os.Stat(configPath); err == nil {
		configExists = true
		ui.WarningStyle.Render(fmt.Sprintf("configuration already exists: %s", configPath))
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("%s", ui.ErrorStyle.Render(fmt.Sprintf("checking configuration path: %v", err)))
	}

	if _, err := os.Stat(scriptPath); err == nil {
		scriptExists = true
		ui.WarningStyle.Render(fmt.Sprintf("setup script already exists: %s", scriptPath))
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("%s", ui.ErrorStyle.Render(fmt.Sprintf("checking setup script: %v", err)))
	}

	if !configExists {
		err := os.WriteFile(configPath, []byte(DefaultConfig), 0644)
		if err != nil {
			return fmt.Errorf("%s", ui.ErrorStyle.Render(fmt.Sprintf("creating %s: %v", configPath, err)))
		}
	}

	if !scriptExists {
		// Generate rittaScript.sh.
		if err := os.WriteFile(scriptPath, []byte(setupScriptTemplate), 0755); err != nil {
			return fmt.Errorf("%s", ui.ErrorStyle.Render(fmt.Sprintf("writing setup script: %v", err)))
		}
	}
	return nil
}
