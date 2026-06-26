#!/bin/bash
set -e

PROFILE="cosmonaut-dev"

echo "💀 Matando o cluster ${PROFILE}..."
minikube delete --profile=${PROFILE} --all --purge 2>/dev/null || true

echo "🧹 Limpando Docker..."
docker system prune -af --volumes 2>/dev/null || true

echo "🚀 Startando ${PROFILE}..."
minikube start \
  --profile=${PROFILE} \
  --driver=docker \
  --cpus=6 \
  --memory=18192 \
  --disk-size=50g

echo "🔌 Habilitando addons no ${PROFILE}..."
minikube addons enable ingress --profile=${PROFILE}
minikube addons enable dashboard --profile=${PROFILE}
minikube addons enable metrics-server --profile=${PROFILE}

echo "⏳ Esperando metrics-server ficar pronto..."
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=120s

echo "✅ ${PROFILE} pronto! Testa aí:"
echo "   minikube status --profile=${PROFILE}"
echo "   kubectl top nodes"
echo "   minikube dashboard --profile=${PROFILE}"