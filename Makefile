# DevOpsMaestro Makefile
# Professional build and installation system

# Build variables
BINARY_NAME=dvm
DVT_BINARY_NAME=dvt
NVP_BINARY_NAME=nvp
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)"
DVT_LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)"
NVP_LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)"

# Installation paths (standard for Homebrew compatibility)
PREFIX?=/usr/local
BINDIR=$(PREFIX)/bin
DATADIR=$(HOME)/.devopsmaestro

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

.PHONY: all build clean test install uninstall dev help sync-migrations build-dvt build-nvp

# Default target
all: test build build-dvt build-nvp

## help: Show this help message
help:
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build: sync-migrations
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v

## sync-migrations: Sync migrations to cmd/dvt and cmd/nvp for embedding
sync-migrations:
	@echo "Syncing migrations for dvt and nvp embedding..."
	@if [ -d cmd/dvt/migrations ]; then rm -rf cmd/dvt/migrations; fi
	@cp -r db/migrations cmd/dvt/migrations
	@if [ -d cmd/nvp/migrations ]; then rm -rf cmd/nvp/migrations; fi
	@cp -r db/migrations cmd/nvp/migrations

## build-dvt: Build the dvt binary
build-dvt: sync-migrations
	@echo "Building $(DVT_BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(DVT_LDFLAGS) -o $(DVT_BINARY_NAME) -v ./cmd/dvt/

## build-nvp: Build the nvp binary
build-nvp: sync-migrations
	@echo "Building $(NVP_BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(NVP_LDFLAGS) -o $(NVP_BINARY_NAME) -v ./cmd/nvp/

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(DVT_BINARY_NAME)
	rm -f $(NVP_BINARY_NAME)
	rm -rf cmd/dvt/migrations
	rm -rf cmd/nvp/migrations

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## install: Install dvm to $(BINDIR) (may require sudo)
install: build
	@echo "Installing $(BINARY_NAME) to $(BINDIR)..."
	@mkdir -p $(BINDIR) 2>/dev/null || sudo mkdir -p $(BINDIR)
	@mkdir -p $(DATADIR)
	@if [ -w $(BINDIR) ]; then \
		install -m 0755 $(BINARY_NAME) $(BINDIR)/$(BINARY_NAME); \
	else \
		echo "  (requires sudo for $(BINDIR))"; \
		sudo install -m 0755 $(BINARY_NAME) $(BINDIR)/$(BINARY_NAME); \
	fi
	@echo ""
	@echo "✓ $(BINARY_NAME) installed successfully!"
	@echo ""
	@echo "Location: $(BINDIR)/$(BINARY_NAME)"
	@echo "Version:  $(VERSION)"
	@echo "Data dir: $(DATADIR)"
	@echo ""
	@echo "You can now run: dvm --version"
	@echo ""
	@if ! command -v dvm >/dev/null 2>&1; then \
		echo "⚠ 'dvm' command not found in PATH"; \
		echo "Make sure $(BINDIR) is in your PATH"; \
		echo "Add this to your ~/.zshrc or ~/.bashrc:"; \
		echo "  export PATH=\"$(BINDIR):\$$PATH\""; \
	fi

## uninstall: Remove dvm from $(BINDIR) (may require sudo)
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(BINDIR)..."
	@if [ -w $(BINDIR)/$(BINARY_NAME) ]; then \
		rm -f $(BINDIR)/$(BINARY_NAME); \
	else \
		sudo rm -f $(BINDIR)/$(BINARY_NAME); \
	fi
	@echo "✓ $(BINARY_NAME) uninstalled"
	@echo ""
	@echo "Note: Data directory $(DATADIR) was not removed"
	@echo "To remove data: rm -rf $(DATADIR)"

## dev: Build and run in development mode
dev: build
	@echo "Running in development mode..."
	./$(BINARY_NAME)

## version: Show version information
version:
	@echo "Version:    $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Commit:     $(COMMIT)"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## lint: Run linters (requires golangci-lint)
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install: brew install golangci-lint"; exit 1)
	golangci-lint run

## release: Build release binaries for multiple platforms
release:
	@echo "Building release binaries..."
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64
	@echo "✓ Release binaries created in dist/"

## install-dev: Install with PREFIX=~/.local (for development)
install-dev:
	@$(MAKE) install PREFIX=$(HOME)/.local
