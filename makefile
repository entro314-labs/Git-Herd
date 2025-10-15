# git-herd Makefile

# Variables
BINARY_NAME=git-herd
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION=$(shell go version | cut -d' ' -f3)
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_TIME} -X main.builtBy=make"

# Default target
.PHONY: all
all: clean build

# Build the binary
.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o ${BINARY_NAME} ./cmd/git-herd

# Build with race detection for development
.PHONY: build-dev
build-dev:
	@echo "Building ${BINARY_NAME} with race detection..."
	go build -race ${LDFLAGS} -o ${BINARY_NAME}-dev ./cmd/git-herd

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linting
.PHONY: lint
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -f ${BINARY_NAME}
	rm -f ${BINARY_NAME}-*
	rm -f coverage.out coverage.html

# Install the binary to /usr/local/bin
.PHONY: install
install: build
	@echo "Installing ${BINARY_NAME} to /usr/local/bin..."
	sudo cp ${BINARY_NAME} /usr/local/bin/
	@echo "${BINARY_NAME} installed successfully!"

# Uninstall the binary from /usr/local/bin
.PHONY: uninstall
uninstall:
	@echo "Removing ${BINARY_NAME} from /usr/local/bin..."
	sudo rm -f /usr/local/bin/${BINARY_NAME}
	@echo "${BINARY_NAME} uninstalled successfully!"

# Cross-platform builds
.PHONY: build-all
build-all: clean
	@echo "Building for all platforms..."

	@echo "Building for macOS (Intel)..."
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-darwin-amd64 ./cmd/git-herd

	@echo "Building for macOS (Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${BINARY_NAME}-darwin-arm64 ./cmd/git-herd

	@echo "Building for Linux (x64)..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux-amd64 ./cmd/git-herd

	@echo "Building for Linux (ARM64)..."
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux-arm64 ./cmd/git-herd

	@echo "Building for Windows (x64)..."
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-windows-amd64.exe ./cmd/git-herd

	@echo "Cross-platform builds completed!"

# Create release packages
.PHONY: package
package: build-all
	@echo "Creating release packages..."
	mkdir -p dist

	# macOS packages
	tar -czf dist/${BINARY_NAME}-${VERSION}-darwin-amd64.tar.gz ${BINARY_NAME}-darwin-amd64 README.md
	tar -czf dist/${BINARY_NAME}-${VERSION}-darwin-arm64.tar.gz ${BINARY_NAME}-darwin-arm64 README.md

	# Linux packages
	tar -czf dist/${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz ${BINARY_NAME}-linux-amd64 README.md
	tar -czf dist/${BINARY_NAME}-${VERSION}-linux-arm64.tar.gz ${BINARY_NAME}-linux-arm64 README.md

	# Windows package
	zip -j dist/${BINARY_NAME}-${VERSION}-windows-amd64.zip ${BINARY_NAME}-windows-amd64.exe README.md

	@echo "Release packages created in dist/"

# Development workflow
.PHONY: dev
dev: deps fmt lint test build

# Run the application with example arguments
.PHONY: run
run: build
	./${BINARY_NAME} -v -n

# Run with custom path
.PHONY: run-path
run-path: build
	@read -p "Enter path to scan: " path; \
	./${BINARY_NAME} -v -n "$$path"

# Show help
.PHONY: help
help:
	@echo "git-herd Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build        Build the binary"
	@echo "  build-dev    Build with race detection"
	@echo "  build-all    Cross-platform builds"
	@echo "  clean        Clean build artifacts"
	@echo "  deps         Install dependencies"
	@echo "  dev          Full development workflow (deps, fmt, lint, test, build)"
	@echo "  fmt          Format code"
	@echo "  help         Show this help"
	@echo "  install      Install binary to /usr/local/bin"
	@echo "  lint         Run linters"
	@echo "  package      Create release packages"
	@echo "  run          Run the application with default args"
	@echo "  run-path     Run with custom path input"
	@echo "  test         Run tests"
	@echo "  test-coverage Run tests with coverage"
	@echo "  uninstall    Remove binary from /usr/local/bin"
	@echo ""
	@echo "Version: ${VERSION}"
	@echo "Go Version: ${GO_VERSION}"

# Development tools installation
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Development tools installed!"

# Quick development setup
.PHONY: setup
setup: install-tools deps
	@echo "Development environment setup complete!"