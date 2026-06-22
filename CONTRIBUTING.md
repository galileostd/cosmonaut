# Contributing to Cosmonaut

Thank you for your interest in contributing. This document covers everything you need to get started.

---

## Ground Rules

- All code, comments, commit messages, issues, and pull requests must be in **English**
- Apache 2.0 license applies to all contributions
- The API is not stable until `v1.0.0` — breaking changes are expected in `v0.x`

---

## Development Setup

### Requirements

- Go 1.26+
- Node.js 20+ (for the UI)
- Docker
- `kubectl` + a local Kubernetes cluster (Minikube or kind)
- `helm` 3.x

### Clone and bootstrap

```bash
git clone https://github.com/galileostd/cosmonaut
cd cosmonaut
make dev-setup
```

This installs Go tools (`golangci-lint`, `controller-gen`, `ko`) and Node dependencies for the UI.

### Run locally

```bash
# start a local Minikube cluster with the base stack
make dev-cluster

# run the control plane against the local cluster
make run-control-plane

# run the UI in dev mode
make run-ui

# build and run the CLI
make build-cli
./bin/cosmo component list
```

---

## Project Structure

```
control-plane/    API server + K8s controllers
cli/              cosmo CLI binary
ui/               SvelteKit web interface
sdk/              Plugin interface — start here for integrations
plugins/          Native plugin implementations (reference)
charts/           Helm charts
docs/             Documentation
hack/             Dev scripts
```

---

## Writing a Plugin

The fastest way to contribute is to write a plugin for a tool that is not yet supported.

1. Read [sdk/plugin.go](./sdk/plugin.go) — the full interface is there
2. Look at [plugins/trino/](./plugins/trino/) as a reference implementation
3. Create `plugins/<your-tool>/`
4. Implement the `Plugin` interface
5. Add a `README.md` to your plugin directory explaining configuration
6. Open a PR

Plugin contributions do not require deep knowledge of the control plane internals.

---

## Pull Request Process

1. Fork the repository
2. Create a branch: `git checkout -b feat/your-feature`
3. Make your changes
4. Run checks: `make lint test`
5. Commit with a clear message (see below)
6. Open a PR against `main`

### Commit message format

```
<type>: <short description>

Types: feat, fix, docs, refactor, test, chore
```

Examples:
```
feat: add Trino plugin health check
fix: handle nil component in registry reconciler
docs: add plugin writing guide
```

---

## Reporting Issues

Use GitHub Issues. Include:

- Cosmonaut version (`cosmo version`)
- Kubernetes version (`kubectl version`)
- What you expected vs what happened
- Relevant logs

---

## Code of Conduct

Be direct. Be respectful. No tolerance for harassment.
