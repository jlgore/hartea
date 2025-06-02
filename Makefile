# Build info
APP_NAME := hartea
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILT_BY := $(shell whoami)

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOVET := $(GOCMD) vet
GOFMT := gofmt

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -X main.builtBy=$(BUILT_BY) -s -w"
BUILD_FLAGS := -a -installsuffix cgo

# Directories
BUILD_DIR := build
CMD_DIR := cmd
INTERNAL_DIR := internal

# Binary names
BINARY_NAME := $(APP_NAME)
BINARY_UNIX := $(BINARY_NAME)_unix

# Docker
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := latest

.PHONY: all build build-linux build-darwin build-windows clean test coverage lint vet fmt check help install deps tidy vendor docker docker-build docker-run dev

## Build

all: clean deps test build ## Run clean, deps, test and build

build: ## Build the binary for current platform
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)/main.go

build-linux: ## Build for Linux (amd64 and arm64)
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)/main.go

build-darwin: ## Build for macOS (amd64 and arm64)
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)/main.go

build-windows: ## Build for Windows (amd64)
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)/main.go

build-all: build-linux build-darwin build-windows ## Build for all platforms

## Development

dev: ## Run the application with example HAR file
	@echo "Running development version..."
	$(GOCMD) run ./$(CMD_DIR)/main.go example.har

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) ./$(CMD_DIR)/main.go

## Dependencies

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

tidy: ## Tidy module dependencies
	@echo "Tidying module dependencies..."
	$(GOMOD) tidy

vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	$(GOMOD) vendor

verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GOMOD) verify

## Testing

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -race -v ./...

test-short: ## Run short tests
	@echo "Running short tests..."
	$(GOTEST) -short -v ./...

coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Code Quality

lint: ## Run linters
	@echo "Running linters..."
	@if command -v staticcheck > /dev/null; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not found, installing..."; \
		$(GOCMD) install honnef.co/go/tools/cmd/staticcheck@latest; \
		staticcheck ./...; \
	fi
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, installing..."; \
		$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run ./...; \
	fi

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -s -w .

fmt-check: ## Check if code is formatted
	@echo "Checking code formatting..."
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "Code is not formatted. Run 'make fmt' to format."; \
		$(GOFMT) -l .; \
		exit 1; \
	fi

check: fmt-check vet lint ## Run all code quality checks

## Docker

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -it $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: ## Push Docker image (requires login)
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

## Cleanup

clean: ## Clean build files
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

clean-deps: ## Clean module cache
	@echo "Cleaning module cache..."
	$(GOCMD) clean -modcache

## Release

release: clean check test build-all ## Prepare release (clean, check, test, build for all platforms)
	@echo "Release build complete. Binaries in $(BUILD_DIR)/"

## Info

version: ## Show version info
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"
	@echo "Built by: $(BUILT_BY)"

help: ## Show this help message
	@echo "$(APP_NAME) Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)