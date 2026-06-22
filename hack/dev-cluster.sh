#!/usr/bin/env bash
# Deletes any existing Minikube cluster and creates a fresh one for Cosmonaut development.
set -euo pipefail

CLUSTER_NAME="cosmonaut-dev"
K8S_VERSION="v1.29.0"
CPUS=4
MEMORY="8g"

echo "==> Deleting existing cluster (if any)"
minikube delete --profile="${CLUSTER_NAME}" 2>/dev/null || true

echo "==> Starting fresh Minikube cluster: ${CLUSTER_NAME}"
minikube start \
  --profile="${CLUSTER_NAME}" \
  --kubernetes-version="${K8S_VERSION}" \
  --cpus="${CPUS}" \
  --memory="${MEMORY}" \
  --driver=docker

echo "==> Setting kubectl context"
kubectl config use-context "${CLUSTER_NAME}"

echo "==> Cluster ready"
kubectl get nodes
