# Makefile for miko-manifest

# Variables
BINARY_NAME=miko-manifest
VERSION?=dev
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build targets
.PHONY: all build clean clean-all test test-coverage test-race test-verbose deps fmt lint docker help

all: test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

clean-all: clean
	@echo "Cleaning all artifacts and directories..."
	rm -rf config templates output
	@echo "Cleaned: binary, config/, templates/, and output/ directories"

# Test targets
test:
	$(GOTEST) -v ./...

test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

test-race:
	$(GOTEST) -v -race ./...

test-verbose:
	$(GOTEST) -v -count=1 ./...

test-lib:
	$(GOTEST) -v ./pkg/mikomanifest/...

test-cmd:
	$(GOTEST) -v ./cmd/...

test-bench:
	$(GOTEST) -v -bench=. ./...

test-short:
	$(GOTEST) -v -short ./...

# Development targets
deps:
	$(GOMOD) tidy
	$(GOMOD) download

fmt:
	$(GOFMT) ./...

lint:
	golangci-lint run

# Docker targets
docker:
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-run:
	docker run --rm $(BINARY_NAME):$(VERSION) --help

# Installation targets
install:
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) .

# CI targets
ci-test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

ci-build:
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(LDFLAGS) -a -installsuffix cgo -o $(BINARY_NAME) .

# Development workflow
dev-setup:
	$(GOMOD) tidy
	$(GOFMT) ./...
	$(GOTEST) -v ./...

dev-test-watch:
	find . -name "*.go" | entr -r make test

# Example usage
example-init:
	./$(BINARY_NAME) init

example-build:
	./$(BINARY_NAME) build --env dev --output-dir example-output

example-check:
	./$(BINARY_NAME) check --config-dir config

example-lint:
	./$(BINARY_NAME) lint --dir example-output

# Help
help:
	@echo "Available targets:"
	@echo "  all           - Run tests and build"
	@echo "  build         - Build the binary"
	@echo "  clean         - Clean build artifacts"
	@echo "  clean-all     - Clean build artifacts and remove config/, templates/, output/ directories"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-race     - Run tests with race detector"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-lib      - Run library tests only"
	@echo "  test-cmd      - Run command tests only"
	@echo "  test-bench    - Run benchmarks"
	@echo "  test-short    - Run short tests only"
	@echo "  deps          - Download dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  docker        - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  install       - Install binary to GOPATH"
	@echo "  ci-test       - Run CI tests"
	@echo "  ci-build      - Build for CI"
	@echo "  dev-setup     - Set up development environment"
	@echo "  example-*     - Run example commands"
	@echo "  help          - Show this help message"
