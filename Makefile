.PHONY: build test clean example lint

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Binary names
EXAMPLE_BINARY=zkexample

# Build directories
BUILD_DIR=bin

all: test build

build: 
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(EXAMPLE_BINARY) ./cmd/example

# Build for multiple platforms
build-all: build-linux build-windows build-arm

build-linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(EXAMPLE_BINARY)-linux-amd64 ./cmd/example

build-windows:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(EXAMPLE_BINARY)-windows-amd64.exe ./cmd/example

build-arm:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(EXAMPLE_BINARY)-linux-arm64 ./cmd/example

test:
	$(GOTEST) -v -race ./...

test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

fmt:
	$(GOFMT) ./...

tidy:
	$(GOMOD) tidy

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run example with device IP
run:
	@if [ -z "$(IP)" ]; then \
		echo "Usage: make run IP=192.168.1.201"; \
	else \
		$(GOCMD) run ./cmd/example $(IP); \
	fi

# Integration test with real device
test-integration:
	@if [ -z "$(ZKTECO_IP)" ]; then \
		echo "Set ZKTECO_IP environment variable"; \
		exit 1; \
	fi
	$(GOTEST) -v -tags=integration ./...

help:
	@echo "Available commands:"
	@echo "  make build          - Build example binary"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make test           - Run unit tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linter"
	@echo "  make fmt            - Format code"
	@echo "  make tidy           - Tidy go modules"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make run IP=x.x.x.x - Run example with device"
