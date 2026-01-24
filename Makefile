.PHONY: build test clean install fmt vet generate help

BINARY_NAME=peb
BUILD_DIR=bin
CMD_DIR=cmd/peb

# Default target
build: generate
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

generate:
	go generate ./...

test: generate
	@echo "Running tests..."
	go test ./...

test-coverage: generate
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

install: generate
	@echo "Installing $(BINARY_NAME)..."
	go install ./$(CMD_DIR)

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet: generate
	@echo "Running go vet..."
	go vet ./...

mod-tidy:
	@echo "Tidying go.mod..."
	go mod tidy

deps:
	@echo "Downloading dependencies..."
	go mod download

help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install the binary to GOPATH/bin"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  generate       - Run go generate"
	@echo "  mod-tidy       - Tidy go.mod"
	@echo "  deps           - Download dependencies"
	@echo "  help           - Show this help message"
