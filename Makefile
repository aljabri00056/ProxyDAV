# ProxyDAV Makefile

.PHONY: build run clean test test-coverage help deps fmt vet build-all clean-dist

# Version information
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD)
COMMIT ?= $(shell git rev-parse HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -w -s"

# Build the application
build:
	@echo "Building ProxyDAV..."
	@go build $(LDFLAGS) -o proxydav ./cmd/proxydav

# Run the application with default settings
run:
	@echo "Starting ProxyDAV server..."
	@go run ./cmd/proxydav

# Run with custom mapping file
run-config:
	@echo "Starting ProxyDAV server with custom mappings..."
	@go run ./cmd/proxydav -port 9000 -mappings files.json -cache-ttl 600

# Run in redirect mode
run-redirect:
	@echo "Starting ProxyDAV server in redirect mode..."
	@go run ./cmd/proxydav -redirect

# Run with authentication
run-auth:
	@echo "Starting ProxyDAV server with authentication..."
	@go run ./cmd/proxydav -auth -user admin -pass secret

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f proxydav
	@go clean ./...

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod tidy
	@go mod download

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet code for potential issues
vet:
	@echo "Vetting code..."
	@go vet ./...

# Run all checks (fmt, vet, test)
check: fmt vet test

# Install development tools
dev-tools:
	@echo "Installing development tools..."
	@go install golang.org/x/tools/cmd/goimports@latest

# Cross-platform builds with zip packaging
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 freebsd/amd64

build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		if [ "$$GOOS" = "windows" ]; then \
			BINARY_NAME="proxydav.exe"; \
			ZIP_NAME="proxydav-$(VERSION)-$$GOOS-$$GOARCH.zip"; \
		else \
			BINARY_NAME="proxydav"; \
			ZIP_NAME="proxydav-$(VERSION)-$$GOOS-$$GOARCH.zip"; \
		fi; \
		echo "Building $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 go build $(LDFLAGS) -o "dist/$$BINARY_NAME" ./cmd/proxydav; \
		cd dist && zip "$$ZIP_NAME" "$$BINARY_NAME" && rm "$$BINARY_NAME" && cd ..; \
	done
	@echo "All builds completed in dist/ directory"

# Clean distribution directory
clean-dist:
	@echo "Cleaning distribution directory..."
	@rm -rf dist/

# Show help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for all platforms and create zip files"
	@echo "  run          - Run with default settings"
	@echo "  run-config   - Run with custom mapping file"
	@echo "  run-redirect - Run in redirect mode"
	@echo "  run-auth     - Run with authentication"
	@echo "  clean        - Clean build artifacts"
	@echo "  clean-dist   - Clean distribution directory"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  deps         - Download dependencies"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code for potential issues"
	@echo "  check        - Run fmt, vet, and test"
	@echo "  dev-tools    - Install development tools"
	@echo "  help         - Show this help message"

# Default target
all: deps build
