.PHONY: help build install clean test fmt vet lint run

BINARY_DIR := bin
BINARY_NAME := slack2md
CMD_PATH := cmd/slack2md

help:
	@echo "Available targets:"
	@echo "  make build    - Build the slack2md binary"
	@echo "  make install  - Install slack2md to GOPATH/bin"
	@echo "  make run      - Run slack2md (use ARGS= to pass arguments)"
	@echo "  make test     - Run all tests"
	@echo "  make fmt      - Format all Go code"
	@echo "  make vet      - Run go vet on all packages"
	@echo "  make lint     - Run golangci-lint (requires installation)"
	@echo "  make clean    - Remove build artifacts"

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_DIR)/$(BINARY_NAME) ./$(CMD_PATH)
	@echo "Built: $(BINARY_DIR)/$(BINARY_NAME)"

install:
	@echo "Installing $(BINARY_NAME)..."
	@go install ./$(CMD_PATH)
	@echo "Installed $(BINARY_NAME) to GOPATH/bin"

run:
	@go run ./$(CMD_PATH) $(ARGS)

test:
	@echo "Running tests..."
	@go test -v ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINARY_DIR)
	@echo "Clean complete"
