#!/usr/bin/env bash
# Runs a full manual test of the CRD + health controller against a local Minikube cluster.
#
# Prerequisites:
#   - Minikube installed
#   - Go 1.22+
#   - kubectl
#
# Usage:
#   ./hack/test-manual.sh
set -euo pipefail

CLUSTER_NAME="cosmonaut-dev"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CRD_MANIFEST="${REPO_ROOT}/charts/cosmonaut/templates/crd-cosmocomponent.yaml"
EXAMPLES_DIR="${REPO_ROOT}/docs/examples"

# ─── colors ────────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

step()  { echo -e "\n${GREEN}==>${NC} $*"; }
warn()  { echo -e "${YELLOW}WARN:${NC} $*"; }
die()   { echo -e "${RED}ERROR:${NC} $*"; exit 1; }

# ─── 1. cluster ────────────────────────────────────────────────────────────────
step "Setting up Minikube cluster"

if minikube status --profile="${CLUSTER_NAME}" &>/dev/null; then
  warn "Cluster '${CLUSTER_NAME}' already exists — reusing it."
  warn "Run './hack/dev-cluster.sh' first if you want a clean slate."
else
  "${REPO_ROOT}/hack/dev-cluster.sh"
fi

kubectl config use-context "${CLUSTER_NAME}"

# ─── 2. install CRD ────────────────────────────────────────────────────────────
step "Installing CosmoComponent CRD"
kubectl apply -f "${CRD_MANIFEST}"

echo "Waiting for CRD to be established..."
kubectl wait \
  --for=condition=Established \
  --timeout=30s \
  crd/cosmocomponents.cosmonaut.galileostd.io

# ─── 3. create namespace ───────────────────────────────────────────────────────
step "Creating cosmonaut namespace"
kubectl create namespace cosmonaut --dry-run=client -o yaml | kubectl apply -f -

# ─── 4. build control plane ────────────────────────────────────────────────────
step "Building control plane"
cd "${REPO_ROOT}"
cd "${REPO_ROOT}/control-plane"
go build -o /tmp/cosmonaut-cp ./cmd/server
cd "${REPO_ROOT}"
echo "Binary: /tmp/cosmonaut-cp"

# ─── 5. run control plane in background ───────────────────────────────────────
step "Starting control plane (background)"

# controller-runtime needs a kubeconfig to talk to the cluster
export COSMONAUT_DEV="true"

/tmp/cosmonaut-cp &
CP_PID=$!
echo "Control plane PID: ${CP_PID}"

# ensure we kill the control plane when the script exits
trap "echo 'Stopping control plane...'; kill ${CP_PID} 2>/dev/null || true" EXIT

echo "Waiting for control plane to start..."
sleep 3

# verify it's alive via the health probe
if curl -sf http://localhost:8081/healthz > /dev/null; then
  echo "Control plane is healthy"
else
  die "Control plane health check failed — check the logs above"
fi

# ─── 6. apply a CosmoComponent ────────────────────────────────────────────────
step "Applying a test CosmoComponent (Trino — unreachable, expected unhealthy)"

cat <<MANIFEST | kubectl apply -f -
apiVersion: cosmonaut.galileostd.io/v1
kind: CosmoComponent
metadata:
  name: trino-test
  namespace: cosmonaut
spec:
  plugin: trino
  type: query-engine
  endpoint: http://trino-does-not-exist:8080
  healthCheckIntervalSeconds: 10
MANIFEST

# ─── 7. watch the status ──────────────────────────────────────────────────────
step "Watching status updates (Ctrl+C to stop)"
echo ""
echo "Expected: health=unhealthy (Trino endpoint does not exist)"
echo ""

# poll every 2 seconds for 30 seconds
for i in $(seq 1 15); do
  echo -n "[${i}/15] "
  kubectl get cosmocomponent trino-test -n cosmonaut \
    -o custom-columns='NAME:.metadata.name,PLUGIN:.spec.plugin,HEALTH:.status.health,MESSAGE:.status.message' \
    --no-headers 2>/dev/null || echo "(not yet reconciled)"
  sleep 2
done

echo ""
step "Full status object"
kubectl get cosmocomponent trino-test -n cosmonaut -o yaml

# ─── 8. summary ───────────────────────────────────────────────────────────────
echo ""
step "Test complete"
echo ""
echo "What to verify:"
echo "  1. CRD was installed:       kubectl get crd cosmocomponents.cosmonaut.galileostd.io"
echo "  2. Component was created:   kubectl get cosmocomponents -n cosmonaut"
echo "  3. Health was reconciled:   status.health should be 'unhealthy' or 'unknown'"
echo "  4. Conditions are present:  status.conditions[0].type == 'Healthy'"
echo ""
echo "To inspect the full object:"
echo "  kubectl get cosmocomponent trino-test -n cosmonaut -o yaml"
echo ""
echo "To clean up:"
echo "  kubectl delete cosmocomponent trino-test -n cosmonaut"
echo "  minikube delete --profile=${CLUSTER_NAME}"
