#!/bin/bash
# minikube-setup.sh - Set up Minikube for development

set -e

echo "🚀 Setting up Minikube for Cosmonaut..."

command -v minikube >/dev/null 2>&1 || { echo "❌ Minikube not found. Install: https://minikube.sigs.k8s.io/docs/start/"; exit 1; }

minikube start --cpus=4 --memory=8192 --disk-size=50g
minikube addons enable ingress
minikube addons enable dashboard
minikube addons enable metrics-server

echo "✅ Minikube configured!"
echo ""
echo "To access the dashboard:"
echo "  minikube dashboard"
