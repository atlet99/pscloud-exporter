# Project-specific variables
BINARY_NAME := pscloud-exporter
OUTPUT_DIR := bin
CMD_DIR := cmd/pscloud-exporter
TAG_NAME ?= $(shell head -n 1 .release-version 2>/dev/null || echo "v0.0.0")
VERSION ?= $(shell head -n 1 .release-version 2>/dev/null || echo "dev")
BUILD ?= $(shell if tail -n 1 .release-version 2>/dev/null | grep -q "build"; then tail -n 1 .release-version | sed -E 's/.*\(build ([0-9]+)\).*/\1/'; else echo "unknown"; fi)
GO_FILES := $(wildcard $(CMD_DIR)/*.go)
GOLANGCI_LINT_VERSION := v1.57.2
GOLANGCI_LINT_PATH := $(shell go env GOPATH)/bin/golangci-lint

# Ensure the output directory exists
$(OUTPUT_DIR):
	@mkdir -p $(OUTPUT_DIR)

# Default target
.PHONY: default
default: fmt vet lint build quicktest

# Run the application locally
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	go run ./$(CMD_DIR)

# Install project dependencies
.PHONY: install-deps
install-deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod vendor
	@echo "Installing golangci-lint..."
	@if ! command -v $(GOLANGCI_LINT_PATH) > /dev/null; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
		echo "golangci-lint installed successfully"; \
	else \
		echo "golangci-lint is already installed"; \
	fi

# Clean up vendor
.PHONY: clean-deps
clean-deps:
	@echo "Cleaning up vendor dependencies..."
	rm -rf vendor

# Update dependencies to latest versions
.PHONY: update-deps update-deps-major

update-deps:
	@echo "Updating dependencies to latest minor/patch versions..."
	go get -u ./...
	go mod tidy
	@echo "Don't forget to run 'make test' to verify the updates"

update-deps-major:
	@echo "Updating dependencies to latest major versions (may include breaking changes)..."
	go get -u -t ./...
	go mod tidy
	@echo "WARNING: Major version updates may include breaking changes!"
	@echo "Please run 'make test' to verify the updates"

# Build binary for current OS/Arch
.PHONY: build
build: $(OUTPUT_DIR)
	@echo "Building $(BINARY_NAME) version $(VERSION) build $(BUILD)..."
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-X 'main.Version=$(VERSION)' -X 'main.Build=$(BUILD)'" -o $(OUTPUT_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Build binaries for multiple platforms
.PHONY: build-cross
build-cross: $(OUTPUT_DIR)
	@echo "Building cross-platform binaries..."
	GOOS=linux   GOARCH=amd64   go build -ldflags="-X 'main.Version=$(VERSION)' -X 'main.Build=$(BUILD)'" -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=darwin  GOARCH=arm64   go build -ldflags="-X 'main.Version=$(VERSION)' -X 'main.Build=$(BUILD)'" -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64   go build -ldflags="-X 'main.Version=$(VERSION)' -X 'main.Build=$(BUILD)'" -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "Cross-platform binaries are available in $(OUTPUT_DIR):"
	@ls -1 $(OUTPUT_DIR)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(OUTPUT_DIR)

# Run tests with coverage and race detection
.PHONY: test
test:
	@echo "Running all tests with race detection and coverage..."
	go test -v -race -cover ./...

# Quick unit tests
.PHONY: quicktest
quicktest:
	@echo "Running quick tests..."
	go test ./...

# Coverage report
.PHONY: test-coverage
test-coverage:
	@echo "Generating coverage report..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run go fmt
.PHONY: fmt
fmt:
	@echo "Running go fmt..."
	go fmt ./...

# Run golangci-lint
.PHONY: lint
lint:
	@echo "Running linter..."
	@if ! command -v $(GOLANGCI_LINT_PATH) > /dev/null; then \
		echo "golangci-lint is not installed. Running 'make install-deps' to install it..."; \
		make install-deps; \
	fi
	$(GOLANGCI_LINT_PATH) run

# Run all checks
.PHONY: check
check: fmt vet lint test

# Create release
.PHONY: release
release: check build-cross
	@echo "Creating release $(TAG_NAME)..."
	@echo "Release $(TAG_NAME) created successfully"

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  default         - Run fmt, vet, lint, build, and quicktest"
	@echo "  run             - Run the application locally"
	@echo "  install-deps    - Install project dependencies"
	@echo "  clean-deps      - Clean up vendor dependencies"
	@echo "  update-deps     - Update dependencies to latest minor/patch versions"
	@echo "  update-deps-major - Update dependencies to latest major versions"
	@echo "  build           - Build binary for current OS/Arch"
	@echo "  build-cross     - Build binaries for multiple platforms"
	@echo "  clean           - Clean build artifacts"
	@echo "  test            - Run all tests with race detection and coverage"
	@echo "  quicktest       - Run quick tests"
	@echo "  test-coverage   - Generate coverage report"
	@echo "  vet             - Run go vet"
	@echo "  fmt             - Run go fmt"
	@echo "  lint            - Run golangci-lint"
	@echo "  check           - Run all checks (fmt, vet, lint, test)"
	@echo "  release         - Create release"
	@echo "  help            - Show this help message"