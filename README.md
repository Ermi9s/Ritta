<div align="center">

# Ritta 2.0

**A centrally-configured, SSH-powered deployment tool built with Go.**

Deploy your applications to any remote server simply, repeatably, and with style.

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://golang.org/)
[![Version](https://img.shields.io/badge/Release-v2.0.1-blue.svg)](https://github.com/Ermi9s/Ritta/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)

<p align="center">
  <img src="asset/ritta(transparent).png" alt="Ritta Demo" width="800" />
</p>

</div>

---

## Table of Contents

- [What is Ritta?](#what-is-ritta)
- [What's New in Ritta 2.0?](#whats-new-in-ritta-20)
- [Why Ritta?](#why-ritta)
- [Deployment Lifecycle](#deployment-lifecycle)
- [Supported Providers](#supported-providers)
  - [Reverse Proxy Providers](#reverse-proxy-providers)
  - [TLS Providers](#tls-providers)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Quick Start](#quick-start)
- [CLI Reference](#cli-reference)
  - [ritta init](#ritta-init)
  - [ritta validate](#ritta-validate)
  - [ritta deploy](#ritta-deploy)
  - [ritta nginx sync](#ritta-nginx-sync)
- [Configuration Reference (`rittaConfig.yaml`)](#configuration-reference-rittaconfigyaml)
- [Deployment Directory & Nginx Structure](#deployment-directory--nginx-structure)
- [Environment & Credential Management](#environment--credential-management)
- [Terminal UI & Scrollable Logs](#terminal-ui--scrollable-logs)
- [Safety, Rollback & Locking](#safety-rollback--locking)
- [Contributing](#contributing)
- [License](#license)

---

## What is Ritta?

Ritta is a lightweight, modern CLI deployment tool. Describe **how** your application should be deployed in a single YAML configuration file, then let Ritta connect to your remote server over SSH, provision dependencies, deploy environment files, build and run your services, configure modular Nginx reverse proxies with automatic Let's Encrypt SSL bootstrapping, and monitor health—all while streaming real-time status in an interactive terminal UI.

---

## What's New in Ritta 2.0?

Ritta 2.0 is a major upgrade focused on production-grade web infrastructure, modular proxy management, and deployment reliability:

- 🚀 **Modular Nginx Directory Sync**: Full synchronization of local `deploy/nginx/` configs directly to remote `/etc/nginx/conf.d/` and `/etc/nginx/sites-available/` with automated symlinking to `sites-enabled/`.
- 🌐 **Dedicated Nginx CLI Subcommand**: Run `ritta nginx sync` to update, validate, and reload reverse proxy configurations independently without triggering a full code deployment.
- 🎯 **Environment Scoping (`--env`)**: Target specific deployment tiers (`--env staging`, `--env production`, or `--env all`) to activate matching Nginx virtual hosts and application scopes.
- 🔒 **Two-Phase ACME/SSL Bootstrapping**: Completely solves the classic chicken-and-egg SSL deadlock on fresh servers. Bootstraps an HTTP challenge server, issues certificates via Certbot webroot mode, and seamlessly upgrades to full HTTPS.
- 🛡️ **Multi-Tenant Safe Archival**: Scoped configuration archiving that targets only this project's domains—never disturbing co-hosted applications on the same VPS.
- 🔄 **Automatic Nginx Validation & Rollback**: Validates new Nginx configurations with `nginx -t` before reloading; automatically reverts symlinks if syntax validation fails.
- ⚙️ **Dynamic Variable Expansion & Smart Credential Fallbacks**: Support for `${VAR}` and `${VAR:-default}` syntax, precedence-based `.env.deploy` loading, and automatic fallbacks to standard VPS environment variables (`VPS_HOST`, `VPS_USER`, `VPS_PORT`, `VPS_KEY`).
- 📜 **Interactive Scrollable TUI**: View real-time logs with keyboard scrolling (`↑`/`↓`, `j`/`k`, `PgUp`/`PgDn`, `Home`/`End`), powered by thread-safe logging.

---

## Why Ritta?

- **No Heavy CI/CD Agents Required** – Deploy straight from your terminal or CI runner over standard SSH.
- **Single Source of Truth** – Declarative `rittaConfig.yaml` controls setup, source, files, build, run, proxy, and TLS.
- **Automated Reverse Proxy & TLS** – Out-of-the-box reverse proxy management and automated Let's Encrypt certificate provisioning.
- **Zero Server Footprint** – No daemon or agent needs to stay running on your server.
- **Automated Rollbacks & Concurrency Locks** – Atomic deployments with remote lockfiles (`/tmp/ritta-<hash>.lock`) and git rollbacks on failure.

---

## Deployment Lifecycle

When you run `ritta deploy`, Ritta executes the following pipeline:

| Step | Phase | Description |
|:----:|:------|:------------|
| **1** | **Connect** | Establishes SSH connection with key-based authentication (passphrase-supported) |
| **2** | **Lock** | Acquires a remote deployment mutex (`/tmp/ritta-<hash>.lock`) to prevent concurrent runs |
| **3** | **Authenticate** | Authenticates `sudo` credentials if non-root (bypasses prompt for `root`) |
| **4** | **Setup** | Uploads and executes your server provisioning script (`rittaScript.sh`) |
| **5** | **Source** | Clones/pulls the target Git branch or verifies existing remote application directory |
| **6** | **Environment & Files** | Uploads `.env` files and configured files via SFTP |
| **7** | **Build** | Runs your application build command remotely |
| **8** | **Run** | Starts your application (supports background daemons, systemd, PM2, Docker, nohup) |
| **9** | **Health Check** | Polls your health check command until healthy or timeout |
| **10** | **TLS Bootstrapping** | Pre-provisions Let's Encrypt SSL certificates via webroot HTTP challenges |
| **11** | **Reverse Proxy** | Syncs Nginx site configs, validates with `nginx -t`, and reloads the proxy |

---

## Supported Providers

### Reverse Proxy Providers

Ritta features a modular reverse proxy architecture handling virtual hosts, upstream routing, syntax validation, and zero-downtime reloads.

| Provider | Value in `rittaConfig.yaml` | Features & Behavior |
|:---------|:---------------------------|:--------------------|
| **Nginx** | `nginx` | • Syncs `deploy/nginx/conf.d/*.conf` $\rightarrow$ `/etc/nginx/conf.d/`<br>• Syncs `deploy/nginx/sites-available/*.conf` $\rightarrow$ `/etc/nginx/sites-available/`<br>• Selectively symlinks matching sites to `/etc/nginx/sites-enabled/` based on `--env`<br>• Validates with `sudo nginx -t` and reloads via `systemctl reload nginx`<br>• Automatically rolls back symlinks if validation fails |

---

### TLS Providers

Ritta automates SSL/TLS certificate issuance, challenge routing, and renewal setup.

| Provider | Value in `rittaConfig.yaml` | Features & Behavior |
|:---------|:---------------------------|:--------------------|
| **Let's Encrypt / Certbot** | `letsencrypt` or `certbot` | • Bootstraps temporary HTTP server blocks to answer ACME challenges<br>• Obtains certificates non-interactively via `certbot certonly --webroot -w /var/www/certbot`<br>• Discovers domains from `domains` list or directly from `sites-available/*.conf`<br>• Supports `--email` or `--register-unsafely-without-email` |

---

## Getting Started

### Prerequisites

- **Local Machine**:
  - Go **1.22+** (if compiling from source)
  - Optional audio player for music: `mpv` or `pw-play` (PipeWire) `:)`
- **Remote Server**:
  - Linux server (Ubuntu 20.04/22.04/24.04 or Debian recommended)
  - SSH access (private key authentication)
  - Sudo or root privileges

### Installation

Download the pre-compiled binary from the `bin` directory or releases, or build from source:

```bash
git clone https://github.com/your-username/ritta.git
cd ritta
go build -o bin/ritta cmd/main.go
sudo cp bin/ritta /usr/local/bin/ritta
```

---

### Quick Start

#### 1. Initialize a new project

Run `ritta init` in your project root:

```bash
ritta init
```

This scaffolds:
- `rittaConfig.yaml` – Central deployment configuration.
- `rittaScript.sh` – Remote server bootstrap script.
- `deploy/.env.deploy.example` – Deployment credentials template.
- `deploy/nginx/conf.d/global.conf` – Shared Nginx HTTP-context directives.
- `deploy/nginx/sites-available/app.example.com.conf` – Starter reverse proxy virtual host.

#### 2. Configure deployment credentials

Copy the example file and fill in your remote VPS credentials:

```bash
cp deploy/.env.deploy.example .env.deploy
```

```env
VPS_HOST=203.0.113.10
VPS_USER=root
VPS_PORT=22
VPS_KEY=~/.ssh/id_ed25519
```

> [!TIP]
> `.env.deploy` is automatically added to `.gitignore` by `ritta init` to protect your secrets.

#### 3. Customize `rittaConfig.yaml`

```yaml
local_project_root: ./
remote_project_root: /srv/myapp
deploy_dir: ./deploy

setup_config:
  script: ./rittaScript.sh

source:
  type: git
  repository: "git@github.com:you/myapp.git"
  branch: main

server:
  host: ${VPS_HOST}
  user: ${VPS_USER:-root}
  port: ${VPS_PORT:-22}
  key: ${VPS_KEY}

build:
  command: "npm ci && npm run build"

run:
  command: "pm2 start dist/index.js --name myapp"

health:
  command: "curl -sf http://localhost:3000/health"

proxy:
  provider: nginx

domains:
  - host: app.example.com
    port: 3000
    tls: true
    env: production

tls:
  provider: letsencrypt
  email: admin@example.com
```

#### 4. Validate and Deploy

```bash
# Validate syntax and rules
ritta validate

# Deploy application
ritta deploy

# Or deploy specifically to an environment tier
ritta deploy --env production
```

---

## CLI Reference

### Summary of Commands

| Command | Syntax | Description |
|:--------|:-------|:------------|
| **`ritta`** | `ritta [flags]` | Root command; displays version and help |
| **`ritta init`** | `ritta init [path] [flags]` | Scaffolds config, setup script, and `deploy/` templates |
| **`ritta validate`** | `ritta validate [path] [flags]` | Validates configuration files against schema rules |
| **`ritta deploy`** | `ritta deploy [path] [flags]` | Executes the full remote deployment pipeline |
| **`ritta nginx sync`** | `ritta nginx sync [path] [flags]` | Syncs, validates, and reloads remote Nginx configurations |

---

### `ritta init`

Initializes Ritta in the current directory or target path:

```bash
ritta init [path]
```

**Flags:**
- `-p, --path <string>` – Target directory (default: `./`).
- `-d, --dir <string>` – Alias for `--path`.

---

### `ritta validate`

Validates required fields, port boundaries, provider types, and syntax:

```bash
ritta validate [path]
```

**Flags:**
- `-f, --file <string>` – Path to configuration file (default: `./`).

---

### `ritta deploy`

Executes the remote deployment pipeline:

```bash
ritta deploy [path] [flags]
```

**Flags:**
- `-f, --file <string>` – Path to configuration file (default: `./`).
- `-e, --env <string>` – Target environment filter (`all`, `staging`, `production`) (default: `all`).
- `-s, --scan-env` – Automatically scan and upload discovered `.env` files.

**Examples:**
```bash
# Standard deployment
ritta deploy

# Target production environment only
ritta deploy --env production

# Deploy a specific config file and scan environment files
ritta deploy staging.yaml --env staging --scan-env
```

---

### `ritta nginx sync`

Synchronizes local Nginx configurations to the server without running application builds or restarts:

```bash
ritta nginx sync [config] [flags]
```

**Flags:**
- `-e, --env <string>` – Target environment filter (`all`, `staging`, `production`) (default: `all`).

**Examples:**
```bash
# Sync all Nginx sites
ritta nginx sync

# Sync only staging virtual hosts
ritta nginx sync --env staging
```

---

## Configuration Reference (`rittaConfig.yaml`)

| Field | Type | Required | Default | Description |
|:------|:----:|:--------:|:-------:|:------------|
| `local_project_root` | String | ✅ Yes | `.` | Root directory of the local project. |
| `remote_project_root` | String | ✅ Yes | — | Remote destination path on the server. |
| `deploy_dir` | String | ❌ No | `./deploy` | Path to local deployment directory containing Nginx configs and templates. |
| `setup_config.script` | String | ❌ No | `./rittaScript.sh` | Remote provisioning shell script executed with `sudo`. |
| `source.type` | String | ✅ Yes | — | `"git"` (pulls repository) or `"existing"` (uses current remote directory). |
| `source.repository` | String | Conditional | — | Repository URL (required if `source.type` is `"git"`). |
| `source.branch` | String | Conditional | `main` | Git branch to checkout. |
| `server.host` | String | ✅ Yes | — | Server hostname or IP (supports `${VAR}` expansion). |
| `server.user` | String | ✅ Yes | `root` | SSH user (must have sudo privileges; bypasses sudo prompt if `root`). |
| `server.port` | Integer | ❌ No | `22` | SSH port (1–65535). |
| `server.key` | String | ✅ Yes | `~/.ssh/id_ed25519` | Path to local SSH private key. |
| `file` | Array | ❌ No | `[]` | Explicit files/directories to transfer via SFTP. |
| `file[].from` | String | Required in item | — | Source path relative to `local_project_root`. |
| `file[].to` | String | Required in item | — | Destination path relative to `remote_project_root`. |
| `build.command` | String | ❌ No | `""` | Remote application build command. |
| `run.command` | String | ✅ Yes | `""` | Remote application startup command (auto-daemonized in nohup if needed). |
| `health.command` | String | ❌ No | `""` | Command to verify application health remotely. |
| `proxy.provider` | String | ❌ No | `""` | Reverse proxy provider (e.g. `nginx`). |
| `domains` | Array | ❌ No | `[]` | Domain routing definitions. |
| `domains[].host` | String | Required in item | — | Host domain (e.g. `app.example.com`). |
| `domains[].port` | Integer | Required in item | — | Upstream port to proxy traffic to. |
| `domains[].tls` | Boolean | ❌ No | `false` | Whether to issue an SSL certificate for this domain. |
| `domains[].env` | String | ❌ No | `""` | Environment scope (`staging`, `production`). |
| `tls.provider` | String | ❌ No | `""` | TLS provider (e.g. `letsencrypt`, `certbot`). |
| `tls.email` | String | Conditional | `""` | Email for Let's Encrypt registration and notifications. |

---

## Deployment Directory & Nginx Structure

When using `deploy_dir` (default: `./deploy`), Ritta manages Nginx modularly:

```text
deploy/
├── .env.deploy.example
└── nginx/
    ├── conf.d/
    │   └── global.conf              # Uploaded to /etc/nginx/conf.d/global.conf
    └── sites-available/
        ├── app.example.com.conf     # Uploaded to /etc/nginx/sites-available/
        └── staging.example.com.conf # Uploaded to /etc/nginx/sites-available/
```

### Environment Scoping Rules
- Site files containing `staging` in their filename are activated during `--env staging` or `--env all`.
- Non-staging site files are activated during `--env production` or `--env all`.
- Symlinks are created at `/etc/nginx/sites-enabled/<name>.conf` pointing to `/etc/nginx/sites-available/<name>.conf`.

---

## Environment & Credential Management

### Variable Expansion
`rittaConfig.yaml` supports dynamic expansion:
```yaml
server:
  host: ${VPS_HOST}
  user: ${VPS_USER:-root}
  port: ${VPS_PORT:-22}
  key: ${VPS_KEY}
```

### Loading Order
Values are sourced in the following order of precedence:
1. Environment variables exported in your current shell
2. `.env.deploy` in project root
3. `deploy/.env.deploy`
4. `.env` in project root
5. Built-in smart fallbacks (`VPS_HOST` $\rightarrow$ `SERVER_HOST`, `VPS_KEY` $\rightarrow$ `SERVER_SSH_KEY_PATH`)

---

## Terminal UI & Scrollable Logs

Ritta provides an interactive BubbleTea terminal UI during deployments:

- **Scrollable Output**: Use `↑` / `↓` or `k` / `j` to scroll line-by-line.
- **Page Scrolling**: Use `PgUp` / `PgDn` to navigate large log outputs.
- **Top / Bottom**: Use `Home` / `g` to jump to the top, and `End` / `G` to jump to latest logs.
- **Thread-Safe**: Logs from SSH commands, environment scanners, and proxy providers are synchronized to prevent visual corruption.

---

## Safety, Rollback & Locking

### Concurrency Protection
Before deploying, Ritta obtains an atomic remote lock at `/tmp/ritta-<hash>.lock`. If another deployment is active for the same remote project root, Ritta aborts immediately to avoid corrupted states. The lock is cleanly released upon completion or failure.

### Automatic Nginx Rollback
If `nginx -t` validation fails during directory synchronization or domain configuration, Ritta automatically rolls back newly created symlinks in `/etc/nginx/sites-enabled/`, leaving your existing production proxy configurations intact.

### Automatic Git Rollback
If a failure occurs during the **build**, **run**, or **health check** phases of a Git-based deployment, Ritta automatically rolls back to the previous stable commit and restores the working state.

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) to learn how to get started, report issues, and submit pull requests.

---

## License

Ritta is open-source software released under the [MIT License](LICENSE).
