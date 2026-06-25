.PHONY: all build test clean dev fmt lint ui copy-ui clean-ui

all: ui copy-ui
	@$(MAKE) build || true
	@$(MAKE) clean-ui

build:
	@echo "🔨 Building Go binary..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/cosmonaut ./cmd/server

ui:
	@echo "📦 Building UI..."
	cd ui && npm install && npm run build

copy-ui:
	@echo "📋 Copying UI build to internal/ui/build..."
	@rm -rf internal/ui/build
	@cp -r ui/build internal/ui/build

clean-ui:
	@echo "🧹 Removing temporary UI build from internal/ui/build..."
	@rm -rf internal/ui/build

test:
	@echo "🧪 Running tests..."
	go test ./...

clean:
	@echo "🧹 Cleaning..."
	rm -rf bin/
	rm -rf internal/ui/build
	rm -rf ui/build ui/.svelte-kit

dev:
	@echo "🚀 Starting dev environment..."
	./hack/dev-setup.sh

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

help:
	@echo "Cosmonaut Makefile targets:"
	@echo "  make all      - Build UI, copy to internal, build Go, clean temp (always cleans)"
	@echo "  make build    - Build Go binary only (requires UI in internal/ui/build)"
	@echo "  make ui       - Build UI only"
	@echo "  make copy-ui  - Copy UI build to internal/ui/build"
	@echo "  make clean-ui - Remove temporary UI build from internal/ui/build"
	@echo "  make test     - Run all tests"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make dev      - Start dev environment"
	@echo "  make fmt      - Format Go code"
	@echo "  make lint     - Run linter"