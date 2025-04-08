# Project-specific variables
BINARY_NAME := pscloud-exporter
OUTPUT_DIR := bin
CMD_DIR := cmd/pscloud-exporter
TAG_NAME ?= $(shell head -n 1 .release-version 2>/dev/null || echo "v0.0.0")
VERSION_RAW ?= $(shell tail -n 1 .release-version 2>/dev/null || echo "dev")
VERSION ?= $(VERSION_RAW)
GO_FILES := $(wildcard $(CMD_DIR)/*.go)

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

# Clean up vendor
.PHONY: clean-deps
clean-deps:
	@echo "Cleaning up vendor dependencies..."
	rm -rf vendor

# Build binary for current OS/Arch
.PHONY: build
build: $(OUTPUT_DIR)
	@echo "Building $(BINARY_NAME) with version $(VERSION)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-X 'main.Version=$(VERSION)'" -o $(OUTPUT_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Build binaries for multiple platforms
.PHONY: build-cross
build-cross: $(OUTPUT_DIR)
	@echo "Building cross-platform binaries..."
	GOOS=linux   GOARCH=amd64   go build -ldflags="-X 'main.Version=$(VERSION)'" -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=darwin  GOARCH=arm64   go build -ldflags="-X 'main.Version=$(VERSION)'" -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64   go build -ldflags="-X 'main.Version=$(VERSION)'" -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
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

# Code formatting
.PHONY: fmt
fmt:
	@echo "Running go fmt..."
	go fmt ./...

# Install and run golangci-lint
.PHONY: install-lint lint lint-fix

install-lint:
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	@echo "Running linter..."
	@golangci-lint run

lint-fix:
	@echo "Running linter with auto-fix..."
	@golangci-lint run --fix

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  default        - fmt + vet + lint + build + quicktest"
	@echo "  run            - Run the application"
	@echo "  install-deps   - Tidy and vendor go modules"
	@echo "  clean-deps     - Clean up vendor dependencies"
	@echo "  build          - Build the binary"
	@echo "  build-cross    - Cross-platform build"
	@echo "  clean          - Remove binaries"
	@echo "  test           - Run tests with race & coverage"
	@echo "  test-coverage  - Generate HTML coverage report"
	@echo "  quicktest      - Run quick tests"
	@echo "  fmt            - Format Go code"
	@echo "  vet            - Run go vet"
	@echo "  lint           - Run golangci-lint"
	@echo "  lint-fix       - Run golangci-lint with auto-fix"
	@echo "  help           - Show this help"