.PHONY: build clean install test lint help

# Build variables
BINARY_NAME=burh
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Default target
all: build

# Build the application
build:
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o ${BINARY_NAME} .

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux-amd64 .

build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-darwin-amd64 .

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-windows-amd64.exe .

# Install the application
install:
	@echo "Installing ${BINARY_NAME}..."
	go install ${LDFLAGS} .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f ${BINARY_NAME}
	rm -f ${BINARY_NAME}-*
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Update dependencies
deps:
	@echo "Updating dependencies..."
	go mod tidy
	go mod download

# Create release
release: clean build-all
	@echo "Creating release..."
	mkdir -p release
	cp ${BINARY_NAME}-* release/
	@echo "Release files created in release/ directory"

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  build-all      - Build for Linux, macOS, and Windows"
	@echo "  build-linux    - Build for Linux"
	@echo "  build-darwin   - Build for macOS"
	@echo "  build-windows  - Build for Windows"
	@echo "  install        - Install the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  deps           - Update dependencies"
	@echo "  release        - Create release builds"
	@echo "  help           - Show this help message"

# Development helpers
dev: build
	@echo "Running in development mode..."
	./${BINARY_NAME}

# Quick test
quick-test:
	@echo "Running quick test..."
	./${BINARY_NAME} --help
	./${BINARY_NAME} create --help
	./${BINARY_NAME} list --help
	./${BINARY_NAME} search --help
