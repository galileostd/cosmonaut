.PHONY: all build test clean dev fmt lint

all: build

build:
	@echo "🔨 Building all components..."
	$(MAKE) -C control-plane build || true
	$(MAKE) -C cli build || true

test:
	@echo "🧪 Running tests..."
	$(MAKE) -C control-plane test || true
	$(MAKE) -C cli test || true
	$(MAKE) -C sdk test || true

clean:
	@echo "🧹 Cleaning..."
	$(MAKE) -C control-plane clean || true
	$(MAKE) -C cli clean || true
	rm -rf ui/dist ui/.svelte-kit

dev:
	@echo "🚀 Starting dev environment..."
	./hack/dev-setup.sh

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

help:
	@echo "Cosmonaut Makefile targets:"
	@echo "  make build    - Build all components"
	@echo "  make test     - Run all tests"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make dev      - Start dev environment"
	@echo "  make fmt      - Format Go code"
	@echo "  make lint     - Run linter"
