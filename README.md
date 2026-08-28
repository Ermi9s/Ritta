<div align="center">

# 🚀 Ritta

**A centrally-configured, SSH-powered deployment tool built with Go.**

Deploy your applications to any remote server — simply, repeatably, and with style.

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)

</div>

---

## What is Ritta?

Ritta is a CLI deployment tool that lets you describe **how** your application
should be deployed in a single YAML configuration file — then deploys it to a
remote server over SSH with one command.

It handles the full deployment lifecycle:

| Step | What Ritta does |
|------|----------------|
| 🔑 **Connect** | Opens an SSH connection to your server |
| ⚙️ **Setup** | Runs your custom setup/bootstrap script |
| 📦 **Source** | Clones/pulls your repository on the server |
| 🌍 **Environment** | Uploads and applies environment variables |
| 🔨 **Build** | Executes your build command on the server |
| ▶️ **Run** | Starts your application |
| ❤️ **Health check** | Verifies the application came up healthy |
| 🔀 **Reverse proxy** | Configures your reverse proxy (e.g., Nginx) |
| 🔒 **TLS** | Provisions and installs TLS certificates |

---

## Why Ritta?

- **No CI/CD infrastructure required** — deploy directly from your terminal.
- **Single config file** — one `rittaConfig.yaml` describes your entire deployment.
- **Pluggable providers** — swap out proxy and TLS providers without changing your workflow.
- **Beautiful TUI** — real-time deployment logs powered by [Bubbletea](https://github.com/charmbracelet/bubbletea).
- **Zero runtime dependencies on the server** — Ritta uses plain SSH + shell scripts.

---

## Getting Started

### Prerequisites

- Go **1.22+**
- SSH access to your remote server (key-based auth recommended)

### Installation

```bash
go install ritta@latest
```

Or build from source:

```bash
git clone https://github.com/your-username/ritta.git
cd ritta
go build -o ritta ./cmd/
```

### Quick Start

**1. Initialize a new deployment config:**

```bash
ritta init --path ./
```

This creates two files in the current directory:

```
rittaConfig.yaml    # Your deployment configuration
rittaScript.sh      # Your server bootstrap/setup script
```

**2. Edit `rittaConfig.yaml`:**

```yaml
local_project_root: "./myapp"
remote_project_root: "/opt/myapp"

source:
  type: git
  repository: "https://github.com/your-username/myapp.git"
  branch: main

server:
  host: "your-server.example.com"
  user: "ubuntu"
  port: 22
  key: "~/.ssh/id_rsa"

setup_config:
  script: "./rittaScript.sh"

build:
  command: "npm run build"

run:
  command: "node dist/index.js"

health:
  command: "curl -sf http://localhost:3000/health"

proxy:
  provider: nginx

domains:
  - host: "myapp.example.com"
    port: 3000
    tls: true

tls:
  provider: certbot
  email: "you@example.com"
```

**3. Validate your configuration:**

```bash
ritta validate --file ./rittaConfig.yaml
```

**4. Deploy:**

```bash
ritta deploy --file ./rittaConfig.yaml
```

Ritta will prompt for your sudo password, then handle the rest.

---

## CLI Reference

| Command | Description |
|---------|-------------|
| `ritta init [--path <dir>]` | Scaffold a new deployment configuration |
| `ritta validate [--file <path>]` | Validate a `rittaConfig.yaml` before deploying |
| `ritta deploy [--file <path>]` | Run the full deployment pipeline |

---

## Configuration Reference

| Field | Required | Description |
|-------|----------|-------------|
| `local_project_root` | ✅ | Path to your project on the local machine |
| `remote_project_root` | ✅ | Deployment root directory on the server |
| `source.type` | ✅ | Source type (e.g., `git`) |
| `source.repository` | ✅ | Repository URL |
| `source.branch` | ✅ | Branch to deploy |
| `server.host` | ✅ | Server hostname or IP |
| `server.user` | ✅ | SSH username |
| `server.port` | ✅ | SSH port (default: `22`) |
| `server.key` | ✅ | Path to private SSH key |
| `setup_config.script` | ✅ | Path to bootstrap shell script |
| `scan_env` | ❌ | Scan and upload local `.env` variables |
| `build.command` | ❌ | Command to build the application |
| `run.command` | ❌ | Command to start the application |
| `health.command` | ❌ | Command to verify application health |
| `proxy.provider` | ❌ | Reverse proxy provider (`nginx`, etc.) |
| `domains` | ❌ | List of domain-to-port mappings |
| `tls.provider` | ❌ | TLS provider (`certbot`, etc.) |
| `tls.email` | ❌ | Email for TLS certificate registration |

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) to
learn how to get started, report bugs, suggest features, and submit pull requests.

---

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md) (v2.1).
By participating, you agree to uphold these standards.

---

## License

Ritta is released under the [MIT License](LICENSE).
