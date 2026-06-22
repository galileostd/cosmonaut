# 🚀 Cosmonaut

> Navigating the cosmos of data with Kubernetes and open-source

Cosmonaut is a Kubernetes-native data platform that unifies orchestration, processing, observability, and governance into a single ecosystem.

## 🎯 Vision

To be the "cosmos" where your data comes to life — a pantheon of open-source tools orchestrated with mastery, where engineers, scientists, and analysts navigate with freedom and control.

## 🏗️ Architecture

```sh
cosmonaut/
├── control-plane/        → Go — main server (API, gRPC)
├── cli/                  → Go — `cosmo` binary
├── ui/                   → SvelteKit — unified portal
├── sdk/                  → Go — plugin interface
├── plugins/              → Plugins for each tool
│   ├── trino/
│   ├── spark/
│   ├── flink/
│   ├── airflow/
│   ├── polaris/
│   └── kafka/
├── charts/               → Helm charts (components)
├── docs/                 → Documentation
└── hack/                 → Dev scripts
```
