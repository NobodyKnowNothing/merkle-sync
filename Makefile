# Universal MerkleSync Makefile

.PHONY: build test lint clean docker-build docker-up docker-down proto

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
SERVER_BINARY=merklesync-server
POSTGRESQL_CONNECTOR_BINARY=postgresql-connector
MONGODB_CONNECTOR_BINARY=mongodb-connector
EDGE_CLIENT_BINARY=edge-client

# Build directories
BUILD_DIR=build
CMD_DIR=cmd

# Default target
all: build

# Build all binaries
build: proto
	@echo "Building all binaries..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(SERVER_BINARY) ./$(CMD_DIR)/server
	$(GOBUILD) -o $(BUILD_DIR)/$(POSTGRESQL_CONNECTOR_BINARY) ./$(CMD_DIR)/postgresql-connector
	$(GOBUILD) -o $(BUILD_DIR)/$(MONGODB_CONNECTOR_BINARY) ./$(CMD_DIR)/mongodb-connector
	$(GOBUILD) -o $(BUILD_DIR)/$(EDGE_CLIENT_BINARY) ./$(CMD_DIR)/edge-client
	@echo "Build complete!"

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@if command -v protoc >/dev/null 2>&1; then \
		protoc --go_out=. --go-grpc_out=. proto/merklesync.proto; \
	else \
		echo "protoc not found, skipping protobuf generation"; \
	fi

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v ./core/... ./server/... ./edge-client/...

# Run integration test
test-integration:
	@echo "Running integration test..."
	$(GOCMD) run integration_test.go

# Run system test
test-system:
	@echo "Running system test..."
	@if [ -f "./scripts/test-system.sh" ]; then \
		chmod +x ./scripts/test-system.sh && ./scripts/test-system.sh; \
	else \
		echo "System test script not found"; \
	fi

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run linting
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, running go vet and go fmt"; \
		$(GOCMD) vet ./...; \
		$(GOCMD) fmt ./...; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Docker commands
docker-build:
	@echo "Building Docker images..."
	docker-compose build

docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-down:
	@echo "Stopping services..."
	docker-compose down

docker-logs:
	@echo "Showing Docker Compose logs..."
	docker-compose logs -f

# Development commands
dev-server:
	@echo "Starting development server..."
	$(GOCMD) run ./$(CMD_DIR)/server -port 50051

dev-postgresql-connector:
	@echo "Starting PostgreSQL connector..."
	$(GOCMD) run ./$(CMD_DIR)/postgresql-connector \
		-db "postgres://user:password@localhost:5432/merklesync?sslmode=disable" \
		-grpc "localhost:50051"

dev-mongodb-connector:
	@echo "Starting MongoDB connector..."
	$(GOCMD) run ./$(CMD_DIR)/mongodb-connector \
		-db "mongodb://localhost:27017" \
		-database "merklesync" \
		-grpc "localhost:50051"

dev-edge-client:
	@echo "Starting edge client..."
	$(GOCMD) run ./$(CMD_DIR)/edge-client \
		-grpc "localhost:50051" \
		-cache "./cache"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u google.golang.org/protobuf/cmd/protoc-gen-go
	$(GOGET) -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

# Help
help:
	@echo "Available targets:"
	@echo "  build              - Build all binaries"
	@echo "  test               - Run all tests"
	@echo "  test-unit          - Run unit tests only"
	@echo "  test-integration   - Run integration test"
	@echo "  test-system        - Run system test"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  lint               - Run linters"
	@echo "  fmt                - Format code"
	@echo "  clean              - Clean build artifacts"
	@echo "  deps               - Download dependencies"
	@echo "  proto              - Generate protobuf code"
	@echo "  docker-build       - Build Docker images"
	@echo "  docker-up          - Start services with Docker Compose"
	@echo "  docker-down        - Stop services"
	@echo "  docker-logs        - Show Docker Compose logs"
	@echo "  dev-server         - Start development server"
	@echo "  dev-postgresql-connector - Start PostgreSQL connector"
	@echo "  dev-mongodb-connector    - Start MongoDB connector"
	@echo "  dev-edge-client    - Start edge client"
	@echo "  install-tools      - Install development tools"
	@echo "  help               - Show this help"
