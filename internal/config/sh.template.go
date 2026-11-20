package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

const globalNginxTemplate = `limit_req_zone $binary_remote_addr zone=app_general:10m rate=20r/s;
limit_req_zone $binary_remote_addr zone=app_auth:10m rate=5r/s;
limit_conn_zone $binary_remote_addr zone=app_conn:10m;

map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}
`

const siteNginxTemplate = `server {
    listen 80;
    listen [::]:80;
    server_name app.example.com;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    listen [::]:443 ssl;
    server_name app.example.com;

    ssl_certificate     /etc/letsencrypt/live/app.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/app.example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL_app:10m;
    ssl_session_timeout 1d;

    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    client_max_body_size 50m;

    location / {
        limit_req zone=app_general burst=40 nodelay;
        limit_conn app_conn 20;

        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
`

const envDeployExampleTemplate = `# Deployment Server Credentials
# Copy this file to .env.deploy and fill in your server details.
# DO NOT commit real secrets to Git.
VPS_HOST=your-server-ip-or-domain
VPS_USER=root
VPS_PORT=22
VPS_KEY=~/.ssh/id_ed25519
`

const setupScriptName = "rittaScript.sh"
const filename = "rittaConfig.yaml"

func ensureGitignore(path string) error {
	gitignorePath := filepath.Join(path, ".gitignore")
	var existingContent string
	if data, err := os.ReadFile(gitignorePath); err == nil {
		existingContent = string(data)
	}

	var toAdd []string
	if !strings.Contains(existingContent, ".env.deploy") {
		toAdd = append(toAdd, ".env.deploy")
	}
	if !strings.Contains(existingContent, ".env\n") && !strings.HasSuffix(existingContent, ".env") {
		toAdd = append(toAdd, ".env")
	}

	if len(toAdd) > 0 {
		var newLines string
		if len(existingContent) > 0 && !strings.HasSuffix(existingContent, "\n") {
			newLines = "\n"
		}
		newLines += strings.Join(toAdd, "\n") + "\n"
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.WriteString(newLines); err != nil {
			return err
		}
	}
	return nil
}

func CreateTemplate(path string) error {
	path = filepath.Clean(path)

	configPath := filepath.Join(path, filename)
	scriptPath := filepath.Join(path, setupScriptName)

	deployDir := filepath.Join(path, "deploy")
	confDDir := filepath.Join(deployDir, "nginx", "conf.d")
	sitesAvailDir := filepath.Join(deployDir, "nginx", "sites-available")
	globalConfPath := filepath.Join(confDDir, "global.conf")
	siteConfPath := filepath.Join(sitesAvailDir, "app.example.com.conf")
	envExamplePath := filepath.Join(deployDir, ".env.deploy.example")

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("%s", ui.ErrorStyle.Render(fmt.Sprintf("creating configuration directory: %v", err)))
	}
	if err := os.MkdirAll(confDDir, 0755); err != nil {
		return fmt.Errorf("creating nginx conf.d directory: %w", err)
	}
	if err := os.MkdirAll(sitesAvailDir, 0755); err != nil {
		return fmt.Errorf("creating nginx sites-available directory: %w", err)
	}

	// 1. rittaConfig.yaml
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, []byte(DefaultConfig), 0644); err != nil {
			return fmt.Errorf("%s", ui.ErrorStyle.Render(fmt.Sprintf("creating %s: %v", configPath, err)))
		}
	}

	// 2. rittaScript.sh
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		if err := os.WriteFile(scriptPath, []byte(setupScriptTemplate), 0755); err != nil {
			return fmt.Errorf("%s", ui.ErrorStyle.Render(fmt.Sprintf("writing setup script: %v", err)))
		}
	}

	// 3. deploy/nginx/conf.d/global.conf
	if _, err := os.Stat(globalConfPath); os.IsNotExist(err) {
		if err := os.WriteFile(globalConfPath, []byte(globalNginxTemplate), 0644); err != nil {
			return fmt.Errorf("writing global nginx config: %w", err)
		}
	}

	// 4. deploy/nginx/sites-available/app.example.com.conf
	if _, err := os.Stat(siteConfPath); os.IsNotExist(err) {
		if err := os.WriteFile(siteConfPath, []byte(siteNginxTemplate), 0644); err != nil {
			return fmt.Errorf("writing sample site nginx config: %w", err)
		}
	}

	// 5. deploy/.env.deploy.example
	if _, err := os.Stat(envExamplePath); os.IsNotExist(err) {
		if err := os.WriteFile(envExamplePath, []byte(envDeployExampleTemplate), 0644); err != nil {
			return fmt.Errorf("writing env.deploy.example: %w", err)
		}
	}

	// 6. Ensure .gitignore protects secrets
	if err := ensureGitignore(path); err != nil {
		return fmt.Errorf("updating .gitignore: %w", err)
	}

	return nil
}
