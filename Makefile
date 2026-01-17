.PHONY: build run test clean install fmt vet help

BINARY_NAME=peb
BUILD_DIR=bin
CMD_DIR=cmd/peb

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	@echo "Running tests..."
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

install:
	@echo "Installing $(BINARY_NAME)..."
	go install ./$(CMD_DIR)

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
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
	@echo "  run            - Build and run the binary"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install the binary to GOPATH/bin"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  mod-tidy       - Tidy go.mod"
	@echo "  deps           - Download dependencies"
	@echo "  help           - Show this help message"
