<div align="center">

# Ritta

**A centrally-configured, SSH-powered deployment tool built with Go.**

Deploy your applications to any remote server simply, repeatably, and with style.

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)

<p align="center">
  <img src="asset/ritta(transparent).png" alt="Ritta Demo" width="800" />
</p>

</div>

---

## Table of Contents

- [What is Ritta?](#what-is-ritta)
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
- [Configuration Reference (`rittaConfig.yaml`)](#configuration-reference-rittaconfigyaml)
- [Environment & File Management](#environment--file-management)
- [Terminal UI & Music Player](#terminal-ui--music-player)
- [Safety & Rollback](#safety--rollback)
- [Contributing](#contributing)
- [License](#license)

---

## What is Ritta?

Ritta is a lightweight CLI deployment tool that lets you describe **how** your application should be deployed in a single YAML configuration file, then provisions, builds, configures reverse proxies and TLS, and deploys to a remote server over SSH with a single command, and she even plays music for you while you wait.

---

## Why Ritta?

- **No Heavy CI/CD Infrastructure Required** – Deploy production ready apps straight from your terminal.
- **Single Source of Truth** – One `rittaConfig.yaml` file configures everything centrally.
- **Automated Reverse Proxy & TLS** – Automatic Nginx configuration generation and Let's Encrypt / Certbot SSL provisioning.
- **Pluggable Architecture** – Clean provider interfaces for reverse proxies and TLS.
- **Zero Server Runtime Dependencies** – Works over standard SSH with shell execution, no custom agents required on your remote machine.
- **Automated Rollbacks & Concurrency Locks** – Safe deployments with git-based rollback on failure.

---

## Deployment Lifecycle

When you run `ritta deploy`, Ritta executes the following deployment jobs:

| Step | Phase | Description |
|:----:|:------|:------------|
| **1** | **Connect** | Establishes an SSH connection with key-based authentication and passphrase support |
| **2** | **Lock** | Acquires a deployment lock (`/tmp/ritta-<hash>.lock`) on the server to prevent concurrent runs |
| **3** | **Authenticate** | Authenticates `sudo` credentials securely via masked terminal input |
| **4** | **Setup** | Uploads and executes your custom server script (`rittaScript.sh`) |
| **5** | **Source** | Clones/pulls the target Git repository if you configured it to be remote or existing remote directory |
| **6** | **Environment & Files** | Uploads configured files and scans local `.env` files if enabled (`ritta deploy --scan-env`) |
| **7** | **Build** | Runs your application build command remotely |
| **8** | **Run** | Starts your application (supports daemons, background processes, systemd, PM2, Docker) |
| **9** | **Health check** | Polls your health check endpoint/command (if you have one) |
| **10** | **Reverse Proxy** | Automatically configures domain reverse proxy routing |
| **11** | **TLS** | Requests and installs SSL/TLS certificates and configures HTTPS redirection |

---

## Supported Providers

### Reverse Proxy Providers

Ritta features a pluggable reverse proxy architecture that handles virtual hosts, upstream routing, configuration testing, and service reloads.

| Provider | Value in `rittaConfig.yaml` | Features & Behavior |
|:---------|:---------------------------|:--------------------|
| **Nginx** | `nginx` | • Generates modular server blocks under `/etc/nginx/conf.d/ritta-<domain>.conf`<br>• Sets standard proxy headers (`Host`, `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`)<br>• Routes incoming requests on port 80 to upstream `http://127.0.0.1:<port>`<br>• Runs syntax validation with `sudo nginx -t`<br>• Enables and reloads the service via `systemctl enable --now nginx && systemctl reload nginx` |

> [!NOTE]
> Ensure Nginx is installed on your server (default `rittaScript.sh` includes `nginx`).

---

### TLS Providers

Ritta automates SSL/TLS certificate provisioning and certificate renewal setup.

| Provider | Value in `rittaConfig.yaml` | Features & Behavior |
|:---------|:---------------------------|:--------------------|
| **Let's Encrypt / Certbot** | `letsencrypt` or `certbot` | • Provisions certificates via `sudo certbot --nginx`<br>• Non-interactive registration with automatic Terms of Service agreement (`--agree-tos --non-interactive`)<br>• Registers using your configured `tls.email`<br>• Automatically configures HTTPS and HTTP-to-HTTPS redirect (`--redirect`) for all domains where `tls: true` |

> [!NOTE]
> Your server must have port 80 and 443 open to the public internet, and your domain DNS records must point to your server IP for Let's Encrypt domain validation to succeed.

---

## Getting Started

### Prerequisites

- **Local Machine**:
  - Go **1.22+** or download the binary from `bin`
  - Optional audio player for music: `mpv` or `pw-play` (PipeWire) `:)`
- **Remote Server**:
  - Linux server 
  - SSH access (private key authentication)
  - Sudo user privileges

### Installation

You can download the binary at `bin` directory for fast use.
OR
Clone the repository and build the binary:

```bash
git clone https://github.com/your-username/ritta.git
cd ritta
go build -o ritta ./cmd/
```

Optionally move the binary into your `$PATH`:

```bash
sudo mv ritta /usr/local/bin/
```

---

### Quick Start

#### 1. Initialize a new deployment configuration

Run `ritta init` in your project root or specify a target directory:

```bash
ritta init
```

This generates two files:
- `rittaConfig.yaml` – Main deployment configuration.
- `rittaScript.sh` – Dependency installation script (pre-configured with execute permissions, and for ubuntu).

#### 2. Configure `rittaConfig.yaml`

Edit `rittaConfig.yaml` to match your application:

```yaml
# Local project directory
local_project_root: ./

# Deployment directory on the remote server
remote_project_root: /opt/myapp

# Server provisioning script
setup_config:
  script: ./rittaScript.sh

# Source configuration
# type: "git" (clone/pull from repository) or "existing" (in-place on server)
source:
  type: git
  repository: "https://github.com/you/yourapp.git"
  branch: main

# SSH server connection details
server:
  host: "203.0.113.10"
  user: "ubuntu"
  port: 22
  key: "~/.ssh/id_ed25519"

# Specific files or directories to copy to the server (optional)
file:
  - from: ".env.production"
    to: ".env"

# Application build command executed on server (optional)
build:
  command: "npm ci && npm run build"

# Application start command executed on server
run:
  command: "pm2 start dist/main.js --name myapp"

# Application health check command (optional)
health:
  command: "curl -sf http://localhost:3000/health"

# Reverse proxy provider
proxy:
  provider: nginx

# Domain and routing mappings
domains:
  - host: "myapp.example.com"
    port: 3000
    tls: true

# TLS certificate provider
tls:
  provider: letsencrypt
  email: "you@example.com"
```

#### 3. Validate your configuration

Check for configuration errors, missing fields, or invalid port ranges before deploying:

```bash
ritta validate
```

#### 4. Deploy

Run the deployment pipeline:

```bash
ritta deploy
```

Ritta connects to your server, securely prompts for SSH key passphrase (if encrypted) and remote sudo password, and runs the deployment pipeline with real-time logs in the TUI.

---

## CLI Reference

### Summary of Commands

| Command | Syntax | Description |
|:--------|:-------|:------------|
| **`ritta`** | `ritta [flags]` | Root command; displays general help and version information |
| **`ritta init`** | `ritta init [path] [flags]` | Scaffolds a new `rittaConfig.yaml` and `rittaScript.sh` |
| **`ritta validate`** | `ritta validate [path] [flags]` | Validates a configuration file against schema rules |
| **`ritta deploy`** | `ritta deploy [path] [flags]` | Executes the complete remote deployment pipeline |

---

### `ritta init`

Creates a starter `rittaConfig.yaml` and executable `rittaScript.sh` in the specified directory. If files already exist, they will not be overwritten.

```bash
ritta init [path] [flags]
```

**Flags:**
- `-p, --path <string>` – Target directory for configuration files (default: `./`).
- `-d, --dir <string>` – Alias for `--path` (default: `./`).
- `-h, --help` – Help for `init`.

**Examples:**
```bash
# Initialize in the current directory
ritta init

# Initialize in a specific directory
ritta init ./deploy
ritta init --path /path/to/project
```

---

### `ritta validate`

Parses and validates a `rittaConfig.yaml` file, ensuring all required fields are present, source settings are valid, domain ports are between 1–65535, and configured proxy/TLS providers are supported.

```bash
ritta validate [path] [flags]
```

**Flags:**
- `-f, --file <string>` – Path to configuration file (default: `./`).
- `-p, --path <string>` – Alias for `--file` (default: `./`).
- `-h, --help` – Help for `validate`.

**Examples:**
```bash
# Validate rittaConfig.yaml in the current directory
ritta validate

# Validate a specific config file
ritta validate ./stagingConfig.yaml
ritta validate -f /etc/ritta/production.yaml
```

---

### `ritta deploy`

Connects to the target server via SSH, checks concurrency locks, runs setup scripts, prepares source code, uploads files, executes build & run commands, verifies health, configures reverse proxy & TLS, and streams live logs in a Bubbletea TUI.

```bash
ritta deploy [path] [flags]
```

**Flags:**
- `-f, --file <string>` – Path to configuration file (default: `./`).
- `-s, --scan-env` – Automatically scan and upload local `.env` files found in the project.
- `-h, --help` – Help for `deploy`.

**Examples:**
```bash
# Deploy using rittaConfig.yaml in current directory
ritta deploy

# Deploy a specific configuration file
ritta deploy ./rittaConfig.yaml
ritta deploy -f ./deploy/prod.yaml

# Deploy and automatically upload discovered .env files
ritta deploy --scan-env
```

---

## Configuration Reference (`rittaConfig.yaml`)

| Field | Type | Required | Default | Description |
|:------|:----:|:--------:|:-------:|:------------|
| `local_project_root` | String | ✅ Yes | `.` | Root directory of the project on the local machine. |
| `remote_project_root` | String | ✅ Yes | — | Absolute destination path on the remote server where the application resides. |
| `setup_config.script` | String | ✅ Yes | `./rittaScript.sh` | Local path to the server bootstrap shell script executed with `sudo`. |
| `source.type` | String | ✅ Yes | — | Source provider: `"git"` (pulls/clones repo) or `"existing"` (uses existing directory on server). |
| `source.repository` | String | Conditional | — | Git repository URL (required if `source.type` is `"git"`). |
| `source.branch` | String | Conditional | `main` | Git branch to checkout (used when `source.type` is `"git"`). |
| `server.host` | String | ✅ Yes | — | Remote server IP address or hostname. |
| `server.user` | String | ✅ Yes | — | SSH login user (must have sudo privileges). |
| `server.port` | Integer | ❌ No | `22` | SSH port (1–65535). |
| `server.key` | String | ✅ Yes | `~/.ssh/id_rsa` | Path to local SSH private key (supports `~` expansion). |
| `file` | Array | ❌ No | `[]` | List of individual files/directories to copy from local to remote. |
| `file[].from` | String | Required in item | — | Source path relative to `local_project_root`. |
| `file[].to` | String | Required in item | — | Destination path relative to `remote_project_root`. |
| `build.command` | String | ❌ No | `""` | Command executed on server to build the app (e.g. `npm run build`, `make`). |
| `run.command` | String | ❌ No | `""` | Command executed on server to launch the app (e.g. `pm2 start`, `docker compose up -d`). |
| `health.command` | String | ❌ No | `""` | Command executed on server to verify health (e.g. `curl -sf http://localhost:3000/health`). |
| `proxy.provider` | String | ❌ No | `""` | Reverse proxy provider (supported: `nginx`). |
| `domains` | Array | ❌ No | `[]` | List of domain routing configurations. |
| `domains[].host` | String | Required in item | — | Domain name (e.g. `api.example.com`). |
| `domains[].port` | Integer | Required in item | — | Local application port to proxy to (1–65535). |
| `domains[].tls` | Boolean | ❌ No | `false` | Whether to provision an SSL/TLS certificate for this domain. |
| `tls.provider` | String | ❌ No | `""` | TLS certificate provider (supported: `letsencrypt`, `certbot`). |
| `tls.email` | String | Conditional | `""` | Email address registered with Let's Encrypt for renewal notices. |

---

## Environment & File Management

### Explicit File Uploads

You can specify exact files or directories to copy from your local machine to the remote application directory using the `file` section in `rittaConfig.yaml`:

```yaml
file:
  - from: ".env.production"
    to: ".env"
  - from: "./config/production.json"
    to: "config/production.json"
```

### Automatic `.env` Discovery

Enable automatic environment file discovery by passing the `--scan-env` (`-s`) flag to `ritta deploy`.

Ritta walks your local directory and uploads all `.env*` files to matching relative paths on the server, while ignoring common build and dependency folders:
- `.git`
- `node_modules`
- `vendor`
- `target`
- `dist`
- `build`

> [!TIP]
> Explicit entries in `file` take precedence over automatically discovered files.

---


## Safety & Rollback

### Concurrency Protection
Before performing any action on the server, Ritta creates a deployment mutex directory `/tmp/ritta-<hash>.lock`. If another deployment is already running on the same remote path, Ritta aborts to prevent conflicts. The lock is automatically released upon completion.

### Automatic Git Rollbacks
If a deployment fails during the **build**, **run**, or **health check** phases:
1. Ritta captures the previous commit hash before updating the repository.
2. If any step fails, Ritta automatically checks out the previous stable commit.
3. Re-runs the build and run commands to restore your service without downtime.

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) to learn how to get started, report bugs, suggest features, and submit pull requests.

---

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md) (v2.1). By participating, you agree to uphold these standards.

---

## License

Ritta is open-source software released under the [MIT License](LICENSE).
