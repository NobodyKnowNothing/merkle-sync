#!/bin/bash

# Universal MerkleSync System Test Script

set -e

echo "🚀 Universal MerkleSync System Test"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    print_status "Checking Go installation..."
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        echo "Download from: https://golang.org/dl/"
        exit 1
    fi
    
    GO_VERSION=$(go version | cut -d' ' -f3)
    print_success "Go is installed: $GO_VERSION"
}

# Check if Docker is installed
check_docker() {
    print_status "Checking Docker installation..."
    if ! command -v docker &> /dev/null; then
        print_warning "Docker is not installed. Some tests will be skipped."
        return 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_warning "Docker Compose is not installed. Some tests will be skipped."
        return 1
    fi
    
    print_success "Docker and Docker Compose are installed"
    return 0
}

# Generate protobuf files
generate_proto() {
    print_status "Generating protobuf files..."
    
    if ! command -v protoc &> /dev/null; then
        print_warning "protoc is not installed. Skipping protobuf generation."
        return 1
    fi
    
    # Install Go protobuf plugins if not present
    if ! command -v protoc-gen-go &> /dev/null; then
        print_status "Installing protoc-gen-go..."
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    fi
    
    if ! command -v protoc-gen-go-grpc &> /dev/null; then
        print_status "Installing protoc-gen-go-grpc..."
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    fi
    
    # Generate protobuf files
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           proto/merklesync.proto
    
    if [ $? -eq 0 ]; then
        print_success "Protobuf files generated successfully"
        return 0
    else
        print_error "Failed to generate protobuf files"
        return 1
    fi
}

# Run Go tests
run_go_tests() {
    print_status "Running Go tests..."
    
    # Download dependencies
    print_status "Downloading dependencies..."
    go mod tidy
    
    # Run tests (excluding examples directory)
    go test -v ./core/... ./server/... ./edge-client/...
    
    if [ $? -eq 0 ]; then
        print_success "All Go tests passed"
        return 0
    else
        print_error "Some Go tests failed"
        return 1
    fi
}

# Build all components
build_components() {
    print_status "Building all components..."
    
    # Create build directory
    mkdir -p build
    
    # Build server
    print_status "Building server..."
    go build -o build/merklesync-server ./cmd/server
    
    # Build PostgreSQL connector
    print_status "Building PostgreSQL connector..."
    go build -o build/postgresql-connector ./cmd/postgresql-connector
    
    # Build MongoDB connector
    print_status "Building MongoDB connector..."
    go build -o build/mongodb-connector ./cmd/mongodb-connector
    
    # Build edge client
    print_status "Building edge client..."
    go build -o build/edge-client ./cmd/edge-client
    
    print_success "All components built successfully"
}

# Test Docker setup
test_docker() {
    if ! check_docker; then
        return 1
    fi
    
    print_status "Testing Docker setup..."
    
    # Build Docker images using optimized script
    print_status "Building Docker images (optimized)..."
    chmod +x ./scripts/build-docker-optimized.sh
    ./scripts/build-docker-optimized.sh
    
    if [ $? -eq 0 ]; then
        print_success "Docker images built successfully"
    else
        print_error "Failed to build Docker images"
        return 1
    fi
    
    # Test Docker Compose configuration
    print_status "Validating Docker Compose configuration..."
    docker-compose config
    
    if [ $? -eq 0 ]; then
        print_success "Docker Compose configuration is valid"
    else
        print_error "Docker Compose configuration is invalid"
        return 1
    fi
}

# Run integration test
run_integration_test() {
    print_status "Running integration test..."
    
    go run examples/integration_demo.go
    
    if [ $? -eq 0 ]; then
        print_success "Integration test passed"
        return 0
    else
        print_error "Integration test failed"
        return 1
    fi
}

# Main test execution
main() {
    local test_results=()
    
    # Check prerequisites
    check_go
    
    # Generate protobuf files (optional)
    if generate_proto; then
        test_results+=("protobuf:✅")
    else
        test_results+=("protobuf:⚠️")
    fi
    
    # Run Go tests
    if run_go_tests; then
        test_results+=("go-tests:✅")
    else
        test_results+=("go-tests:❌")
    fi
    
    # Build components
    if build_components; then
        test_results+=("build:✅")
    else
        test_results+=("build:❌")
    fi
    
    # Test Docker (optional)
    if test_docker; then
        test_results+=("docker:✅")
    else
        test_results+=("docker:⚠️")
    fi
    
    # Run integration test
    if run_integration_test; then
        test_results+=("integration:✅")
    else
        test_results+=("integration:❌")
    fi
    
    # Print summary
    echo ""
    echo "📊 Test Summary"
    echo "==============="
    for result in "${test_results[@]}"; do
        echo "  $result"
    done
    
    # Check if all critical tests passed
    if [[ "${test_results[*]}" == *"go-tests:❌"* ]] || [[ "${test_results[*]}" == *"build:❌"* ]]; then
        print_error "Critical tests failed. Please fix the issues above."
        exit 1
    else
        print_success "All critical tests passed! 🎉"
        echo ""
        echo "🚀 Universal MerkleSync is ready to use!"
        echo ""
        echo "Next steps:"
        echo "  1. Start the system: docker-compose up -d"
        echo "  2. View logs: docker-compose logs -f"
        echo "  3. Run individual components: make dev-server"
        echo ""
    fi
}

# Run main function
main "$@"
