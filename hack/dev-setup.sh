#!/bin/bash
# dev-setup.sh - Set up the development environment

set -e

echo "🚀 Setting up Cosmonaut development environment..."

# Check dependencies
command -v go >/dev/null 2>&1 || { echo "❌ Go not found. Install: https://go.dev/dl/"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "❌ Node.js not found. Install: https://nodejs.org/"; exit 1; }
command -v kubectl >/dev/null 2>&1 || echo "⚠️  kubectl not found (optional)"
command -v helm >/dev/null 2>&1 || echo "⚠️  helm not found (optional)"

echo "📦 Installing Go dependencies..."
for dir in control-plane cli sdk plugins/*; do
    if [ -f "$dir/go.mod" ]; then
        (cd "$dir" && go mod tidy)
    fi
done

echo "📦 Installing UI dependencies..."
cd ui && npm install && cd ..

echo "🔧 Building CLI..."
cd cli && make build && cd ..

echo "✅ Environment ready!"
echo ""
echo "To start the control-plane:"
echo "  cd control-plane && go run cmd/main.go"
echo ""
echo "To use the CLI:"
echo "  ./cli/cosmo version"
echo ""
echo "To start the UI in dev mode:"
echo "  cd ui && npm run dev"
