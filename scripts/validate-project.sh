#!/bin/bash

# Universal MerkleSync Project Validation Script

set -e

echo "🔍 Universal MerkleSync Project Validation"
echo "=========================================="

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

# Check project structure
validate_structure() {
    print_status "Validating project structure..."
    
    local required_dirs=(
        "core"
        "server"
        "connectors/postgresql"
        "connectors/mongodb"
        "edge-client"
        "cmd/server"
        "cmd/postgresql-connector"
        "cmd/mongodb-connector"
        "cmd/edge-client"
        "proto"
        "scripts"
    )
    
    local required_files=(
        "go.mod"
        "README.md"
        "LICENSE"
        "CONTRIBUTING.md"
        "Makefile"
        "docker-compose.yml"
        "proto/merklesync.proto"
        "core/merkle_tree.go"
        "core/merkle_tree_test.go"
        "server/grpc_server.go"
        "server/grpc_server_test.go"
        "connectors/postgresql/postgresql_connector.go"
        "connectors/mongodb/mongodb_connector.go"
        "edge-client/go_client.go"
        "edge-client/go_client_test.go"
        "cmd/server/main.go"
        "cmd/postgresql-connector/main.go"
        "cmd/mongodb-connector/main.go"
        "cmd/edge-client/main.go"
        "Dockerfile.server"
        "Dockerfile.postgresql-connector"
        "Dockerfile.mongodb-connector"
        "Dockerfile.edge-client"
        "integration_test.go"
        "TEST_RESULTS.md"
    )
    
    local missing_dirs=()
    local missing_files=()
    
    # Check directories
    for dir in "${required_dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            missing_dirs+=("$dir")
        fi
    done
    
    # Check files
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            missing_files+=("$file")
        fi
    done
    
    if [ ${#missing_dirs[@]} -eq 0 ] && [ ${#missing_files[@]} -eq 0 ]; then
        print_success "Project structure is complete"
        return 0
    else
        print_error "Project structure issues found:"
        if [ ${#missing_dirs[@]} -gt 0 ]; then
            echo "  Missing directories:"
            for dir in "${missing_dirs[@]}"; do
                echo "    - $dir"
            done
        fi
        if [ ${#missing_files[@]} -gt 0 ]; then
            echo "  Missing files:"
            for file in "${missing_files[@]}"; do
                echo "    - $file"
            done
        fi
        return 1
    fi
}

# Check Go module
validate_go_module() {
    print_status "Validating Go module..."
    
    if [ ! -f "go.mod" ]; then
        print_error "go.mod not found"
        return 1
    fi
    
    # Check if go.sum exists (optional, created by go mod tidy)
    if [ ! -f "go.sum" ]; then
        print_warning "go.sum not found (will be created by 'go mod tidy')"
    fi
    
    # Check if go.mod has required dependencies
    local required_deps=(
        "github.com/google/uuid"
        "github.com/lib/pq"
        "github.com/mongodb/mongo-go-driver"
        "github.com/syndtr/goleveldb"
        "google.golang.org/grpc"
        "google.golang.org/protobuf"
        "golang.org/x/crypto"
    )
    
    local missing_deps=()
    for dep in "${required_deps[@]}"; do
        if ! grep -q "$dep" go.mod; then
            missing_deps+=("$dep")
        fi
    done
    
    if [ ${#missing_deps[@]} -eq 0 ]; then
        print_success "Go module dependencies are complete"
        return 0
    else
        print_error "Missing Go dependencies:"
        for dep in "${missing_deps[@]}"; do
            echo "  - $dep"
        done
        return 1
    fi
}

# Check Docker configuration
validate_docker() {
    print_status "Validating Docker configuration..."
    
    local docker_files=(
        "docker-compose.yml"
        "Dockerfile.server"
        "Dockerfile.postgresql-connector"
        "Dockerfile.mongodb-connector"
        "Dockerfile.edge-client"
    )
    
    local missing_files=()
    for file in "${docker_files[@]}"; do
        if [ ! -f "$file" ]; then
            missing_files+=("$file")
        fi
    done
    
    if [ ${#missing_files[@]} -eq 0 ]; then
        print_success "Docker configuration is complete"
        return 0
    else
        print_error "Missing Docker files:"
        for file in "${missing_files[@]}"; do
            echo "  - $file"
        done
        return 1
    fi
}

# Check documentation
validate_documentation() {
    print_status "Validating documentation..."
    
    local doc_files=(
        "README.md"
        "LICENSE"
        "CONTRIBUTING.md"
        "TEST_RESULTS.md"
    )
    
    local missing_files=()
    for file in "${doc_files[@]}"; do
        if [ ! -f "$file" ]; then
            missing_files+=("$file")
        fi
    done
    
    if [ ${#missing_files[@]} -eq 0 ]; then
        print_success "Documentation is complete"
        return 0
    else
        print_error "Missing documentation files:"
        for file in "${missing_files[@]}"; do
            echo "  - $file"
        done
        return 1
    fi
}

# Check test coverage
validate_tests() {
    print_status "Validating test coverage..."
    
    local test_files=(
        "core/merkle_tree_test.go"
        "server/grpc_server_test.go"
        "edge-client/go_client_test.go"
        "integration_test.go"
    )
    
    local missing_files=()
    for file in "${test_files[@]}"; do
        if [ ! -f "$file" ]; then
            missing_files+=("$file")
        fi
    done
    
    if [ ${#missing_files[@]} -eq 0 ]; then
        print_success "Test coverage is complete"
        return 0
    else
        print_error "Missing test files:"
        for file in "${missing_files[@]}"; do
            echo "  - $file"
        done
        return 1
    fi
}

# Check scripts
validate_scripts() {
    print_status "Validating scripts..."
    
    local script_files=(
        "scripts/generate-proto.sh"
        "scripts/generate-proto.bat"
        "scripts/test-system.sh"
        "scripts/test-system.bat"
        "scripts/validate-project.sh"
    )
    
    local missing_files=()
    for file in "${script_files[@]}"; do
        if [ ! -f "$file" ]; then
            missing_files+=("$file")
        fi
    done
    
    if [ ${#missing_files[@]} -eq 0 ]; then
        print_success "Scripts are complete"
        return 0
    else
        print_error "Missing script files:"
        for file in "${missing_files[@]}"; do
            echo "  - $file"
        done
        return 1
    fi
}

# Main validation function
main() {
    local validation_results=()
    
    if validate_structure; then
        validation_results+=("structure:✅")
    else
        validation_results+=("structure:❌")
    fi
    
    if validate_go_module; then
        validation_results+=("go-module:✅")
    else
        validation_results+=("go-module:❌")
    fi
    
    if validate_docker; then
        validation_results+=("docker:✅")
    else
        validation_results+=("docker:❌")
    fi
    
    if validate_documentation; then
        validation_results+=("documentation:✅")
    else
        validation_results+=("documentation:❌")
    fi
    
    if validate_tests; then
        validation_results+=("tests:✅")
    else
        validation_results+=("tests:❌")
    fi
    
    if validate_scripts; then
        validation_results+=("scripts:✅")
    else
        validation_results+=("scripts:❌")
    fi
    
    # Print summary
    echo ""
    echo "📊 Validation Summary"
    echo "===================="
    for result in "${validation_results[@]}"; do
        echo "  $result"
    done
    
    # Check if all validations passed
    if [[ "${validation_results[*]}" == *"❌"* ]]; then
        print_error "Some validations failed. Please fix the issues above."
        exit 1
    else
        print_success "All validations passed! 🎉"
        echo ""
        echo "🚀 Universal MerkleSync project is complete and ready!"
        echo ""
        echo "Next steps:"
        echo "  1. Install Go 1.21+ if not already installed"
        echo "  2. Run: go mod tidy"
        echo "  3. Run: ./scripts/test-system.sh"
        echo "  4. Start development: make dev-server"
        echo ""
    fi
}

# Run main function
main "$@"
