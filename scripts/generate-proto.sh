#!/bin/bash

# Generate protobuf files for Universal MerkleSync

echo "Generating protobuf files..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed. Please install Protocol Buffers compiler."
    echo "On Ubuntu/Debian: sudo apt-get install protobuf-compiler"
    echo "On macOS: brew install protobuf"
    echo "On Windows: Download from https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

# Check if Go protobuf plugins are installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Generate Go code from protobuf
echo "Generating Go code from proto/merklesync.proto..."
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/merklesync.proto

if [ $? -eq 0 ]; then
    echo "✅ Protobuf files generated successfully!"
    echo "Generated files:"
    echo "  - proto/merklesync.pb.go"
    echo "  - proto/merklesync_grpc.pb.go"
else
    echo "❌ Failed to generate protobuf files"
    exit 1
fi
