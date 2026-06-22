#!/usr/bin/env bash
# Starts a local Minikube cluster for Cosmonaut development.
set -euo pipefail

CLUSTER_NAME="cosmonaut-dev"
K8S_VERSION="v1.29.0"
CPUS=4
MEMORY="8g"

echo "==> Starting Minikube cluster: ${CLUSTER_NAME}"

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
