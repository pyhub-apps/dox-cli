# Makefile for pyhub-documents-cli

# Variables
BINARY_NAME=pyhub-documents-cli
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-s -w -X github.com/pyhub/pyhub-documents-cli/cmd.version=${VERSION} -X github.com/pyhub/pyhub-documents-cli/cmd.commit=${COMMIT} -X github.com/pyhub/pyhub-documents-cli/cmd.date=${DATE}"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Directories
PKG_DIR=./pkg/...
CMD_DIR=./cmd/...
INTERNAL_DIR=./internal/...

.PHONY: all build clean test coverage fmt vet lint help

## help: Display this help message
help:
	@echo "Available targets:"
	@grep -E '^##' Makefile | sed 's/## //'

## all: Format, vet, test, and build
all: fmt vet test build

## build: Build the binary for current platform
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

## build-windows: Build Windows executable
build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME).exe .
	@echo "Windows build complete: $(BINARY_NAME).exe"

## build-darwin: Build macOS executable
build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .
	@echo "macOS build complete"

## build-linux: Build Linux executable
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	@echo "Linux build complete"

## build-all: Build for all platforms
build-all: build-windows build-darwin build-linux
	@echo "All platform builds complete"

## clean: Remove build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out coverage.html
	@echo "Cleanup complete"

## test: Run tests
test:
	$(GOTEST) -v -race $(PKG_DIR) $(CMD_DIR) $(INTERNAL_DIR)

## test-short: Run short tests
test-short:
	$(GOTEST) -v -short $(PKG_DIR) $(CMD_DIR) $(INTERNAL_DIR)

## coverage: Generate test coverage report
coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out $(PKG_DIR) $(CMD_DIR) $(INTERNAL_DIR)
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## benchmark: Run benchmarks
benchmark:
	$(GOTEST) -bench=. -benchmem $(PKG_DIR) $(CMD_DIR) $(INTERNAL_DIR)

## fmt: Format code
fmt:
	$(GOFMT) ./...
	@echo "Code formatted"

## vet: Run go vet
vet:
	$(GOVET) ./...
	@echo "Vet complete"

## lint: Run linter (requires golangci-lint)
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

## deps: Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

## install: Install the binary to GOPATH/bin
install: build
	$(GOCMD) install

## run: Run the application
run:
	$(GOCMD) run . $(ARGS)

## docker-build: Build Docker image
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

## ci: Run CI pipeline locally
ci: fmt vet lint test build
	@echo "CI pipeline complete"

# Development shortcuts
.PHONY: r t b

## r: Quick run
r: run

## t: Quick test
t: test

## b: Quick build
b: build