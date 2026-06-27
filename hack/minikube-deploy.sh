#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="cosmonaut"
IMAGE_TAG="dev"
NAMESPACE="cosmonaut"
RELEASE_NAME="cosmonaut"
CLUSTER_NAME="cosmonaut-dev"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

step()  { echo -e "\n${GREEN}==>${NC} $*"; }
warn()  { echo -e "${YELLOW}WARN:${NC} $*"; }
die()   { echo -e "${RED}ERROR:${NC} $*" >&2; exit 1; }

SKIP_BUILD=false
for arg in "$@"; do
  case $arg in
    --skip-build) SKIP_BUILD=true ;;
  esac
done

command -v minikube >/dev/null 2>&1 || die "minikube não encontrado"
command -v kubectl  >/dev/null 2>&1 || die "kubectl não encontrado"
command -v helm     >/dev/null 2>&1 || die "helm não encontrado"
command -v docker   >/dev/null 2>&1 || die "docker não encontrado"

if ! minikube status --profile="${CLUSTER_NAME}" | grep -q "Running"; then
  die "Minikube '${CLUSTER_NAME}' não está rodando. Execute: minikube start --profile=${CLUSTER_NAME}"
fi

kubectl config use-context "${CLUSTER_NAME}"
step "Contexto kubectl: ${CLUSTER_NAME}"

# ── build ─────────────────────────────────────────────────────────────────────

if [ "${SKIP_BUILD}" = false ]; then
  # DELETA TUDO antes
  step "Copiando UI para internal/ui/build..."
  rm -rf internal/ui/build
  rm -rf ui/.svelte-kit
  rm -rf ui/node_modules/.vite
  rm -rf ui/node_modules/.cache

  # Builda limpo
  cd "${REPO_ROOT}/ui"
  npm install --silent
  npm run build
  cd "${REPO_ROOT}"

  step "Copiando UI para internal/ui/build..."
  rm -rf internal/ui/build
  mkdir -p internal/ui/build
  rm -rf internal/ui/build && cp -r ui/build/. internal/ui/build/

  step "Buildando binário Go..."
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X github.com/galileostd/cosmonaut/internal/api.BuildVersion=dev" \
    -o bin/cosmonaut ./cmd/server

  step "Buildando imagem Docker..."
  docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" "${REPO_ROOT}"

  step "Carregando imagem no Minikube..."
  minikube ssh --profile="${CLUSTER_NAME}" -- docker rmi -f docker.io/library/${IMAGE_NAME}:${IMAGE_TAG} 2>/dev/null || true
  minikube ssh --profile="${CLUSTER_NAME}" -- docker rmi -f docker.io/library/${IMAGE_NAME}:${IMAGE_TAG} 2>/dev/null || true
  minikube image load "${IMAGE_NAME}:${IMAGE_TAG}" --profile="${CLUSTER_NAME}"

  step "Removendo build temporário..."
  # rm -rf internal/ui/build
fi

# ── helm ──────────────────────────────────────────────────────────────────────

step "Atualizando dependências Helm..."
cd "${REPO_ROOT}"
helm repo add cockroachdb https://charts.cockroachdb.com/ 2>/dev/null || true
helm repo update cockroachdb
helm dependency update ./charts/cosmonaut

step "Deletando CRD órfão se existir..."

step "Fazendo deploy via Helm (sem --wait, vamos controlar manualmente)..."
helm upgrade --install "${RELEASE_NAME}" \
  ./charts/cosmonaut \
  -f ./charts/cosmonaut/values-minikube.yaml \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --timeout 10m

# ── cockroachdb init ──────────────────────────────────────────────────────────

step "Aguardando pod CockroachDB ficar Running (container started)..."
kubectl wait pod \
  -n "${NAMESPACE}" \
  -l "app.kubernetes.io/name=cockroachdb" \
  --for=condition=Initialized \
  --timeout=120s

# dá uns segundos pro processo iniciar antes do exec
sleep 10

step "Inicializando cluster CockroachDB single-node..."
if kubectl exec -n "${NAMESPACE}" "${RELEASE_NAME}-cockroachdb-0" -- \
    /cockroach/cockroach init --insecure 2>&1 | grep -qE "already initialized|Cluster successfully initialized"; then
  warn "CockroachDB já inicializado ou inicializado agora — ok"
else
  step "CockroachDB inicializado com sucesso"
fi

step "Aguardando CockroachDB ficar Ready..."
kubectl wait pod \
  -n "${NAMESPACE}" \
  "${RELEASE_NAME}-cockroachdb-0" \
  --for=condition=Ready \
  --timeout=120s

# ── cosmonaut rollout ─────────────────────────────────────────────────────────

step "Forçando rollout do Cosmonaut (agora que o banco está pronto)..."
kubectl rollout restart deployment/"${RELEASE_NAME}" -n "${NAMESPACE}"
kubectl rollout status deployment/"${RELEASE_NAME}" -n "${NAMESPACE}" --timeout=120s

# ── verificação final ─────────────────────────────────────────────────────────

step "Status final:"
kubectl get pods -n "${NAMESPACE}"

echo ""
step "Deploy completo!"
echo ""
echo "Para acessar a UI:"
echo "  kubectl port-forward -n ${NAMESPACE} svc/${RELEASE_NAME} 8080:8080"
echo "  Abra: http://localhost:8080"
echo ""
echo "Para ver os logs:"
echo "  kubectl logs -n ${NAMESPACE} -l app.kubernetes.io/name=cosmonaut -f"
echo ""
echo "Para testar a API:"
echo "  curl http://localhost:8080/api/v1/system/health"