.PHONY: generate build build-cli build-cross run clean

BINARY_NAME ?= a2a-playground
VERSION ?= dev
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

# Generate TS and Go from proto (run from packages/a2a)
generate:
	cd packages/a2a && buf generate

# Build frontend for BFF
build-frontend:
	cd app && pnpm build

# Build the CLI binary (embeds app/dist directly; requires build-frontend first)
build-cli: generate build-frontend
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/a2a-playground/

# Full build: generate, frontend, embed, CLI
build: build-cli

# Cross-compile for all platforms (used by build.sh / CI)
# Usage: make build-cross VERSION=v1.0.0
build-cross:
	@./build.sh $(VERSION)

# Run the CLI (use --dev to serve from app/dist without embedding)
run: generate build-frontend
	go run ./cmd/a2a-playground/ --dev --agent-url=localhost:8080

clean:
	rm -f $(BINARY_NAME)
	rm -rf app/dist
	rm -rf dist
