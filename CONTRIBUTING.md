# Contributing to Ritta

First off — **thank you** for considering contributing to Ritta! 🎉  
It's people like you that make Ritta a great tool for the whole community.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Submitting Code Changes](#submitting-code-changes)
- [Development Setup](#development-setup)
- [Coding Style](#coding-style)
- [Pull Request Process](#pull-request-process)
- [Getting Help](#getting-help)

---

## Code of Conduct

This project and everyone participating in it is governed by the
[Ritta Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected
to uphold this code. Please report unacceptable behavior by opening a GitHub issue.

---

## How Can I Contribute?

### Reporting Bugs

Before filing a bug report, please check the [existing issues](../../issues) to
avoid duplicates.

When you file a bug report, please include:

- **Ritta version** (`ritta --version`)
- **Go version** (`go version`)
- **Operating System** and version
- **Steps to reproduce** the problem
- **Expected behavior** vs **actual behavior**
- Any **relevant logs** or error messages

> **Tip:** Use the `--verbose` flag (if available) to capture more detailed output.

### Suggesting Features

Feature requests are welcome! Please open an issue with the label `enhancement`
and describe:

- **What problem** does this feature solve?
- **How would it work** from the user's perspective?
- **Are there alternatives** you've considered?

### Submitting Code Changes

We welcome pull requests for:

- Bug fixes
- New deployment providers (proxy, TLS)
- Improved error messages and UX
- Documentation improvements
- Test coverage

---

## Development Setup

### Prerequisites

- [Go](https://golang.org/dl/) **1.22+**
- SSH access to a test server (for integration testing)
- Git

### Clone and Build

```bash
git clone https://github.com/your-username/ritta.git
cd ritta

# Download dependencies
go mod download

# Build the binary
go build -o ritta ./cmd/

# Verify it works
./ritta --help
```

### Running Tests

```bash
go test ./...
```

> Unit tests do **not** require a live server. Integration tests that exercise SSH
> and deployment steps may need environment variables set — see `internal/config/`
> for the expected structure.

---

## Coding Style

- Follow standard Go conventions enforced by `gofmt` and `go vet`.
- Run `gofmt -w .` before committing.
- Keep functions focused and small.
- Prefer clear error wrapping with `fmt.Errorf("context: %w", err)`.
- Add comments to exported types, functions, and non-obvious logic.
- New CLI commands belong in `internal/cli/`.
- New deployment providers belong in `internal/proxy/proxy.providers/` or
  `internal/proxy/tls.providers/` as appropriate.

---

## Pull Request Process

1. **Fork** the repository and create your branch from `main`:
   ```bash
   git checkout -b feat/my-awesome-feature
   ```
2. Make your changes and ensure `go test ./...` passes.
3. Run `gofmt -w .` to format your code.
4. Write a clear commit message describing *what* and *why*.
5. Open a Pull Request against the `main` branch.
6. Fill in the PR template (if provided) and link any related issues.
7. A maintainer will review your PR — please be patient and responsive to feedback.
8. Once approved, your PR will be merged. 🎉

> **Breaking changes**: If your change affects the `rittaConfig.yaml` schema or
> CLI flags, document it clearly in the PR description.

---

## Getting Help

- Open a [GitHub Discussion](../../discussions) for general questions.
- Open an [Issue](../../issues) for bugs or feature requests.
- Read the [README](README.md) for usage documentation.

We're happy to help — don't hesitate to ask!
