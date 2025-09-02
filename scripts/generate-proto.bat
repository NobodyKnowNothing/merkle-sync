@echo off
REM Generate protobuf files for Universal MerkleSync

echo Generating protobuf files...

REM Check if protoc is installed
where protoc >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: protoc is not installed. Please install Protocol Buffers compiler.
    echo Download from: https://github.com/protocolbuffers/protobuf/releases
    exit /b 1
)

REM Check if Go protobuf plugins are installed
where protoc-gen-go >nul 2>nul
if %errorlevel% neq 0 (
    echo Installing protoc-gen-go...
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
)

where protoc-gen-go-grpc >nul 2>nul
if %errorlevel% neq 0 (
    echo Installing protoc-gen-go-grpc...
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
)

REM Generate Go code from protobuf
echo Generating Go code from proto/merklesync.proto...
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/merklesync.proto

if %errorlevel% equ 0 (
    echo ✅ Protobuf files generated successfully!
    echo Generated files:
    echo   - proto/merklesync.pb.go
    echo   - proto/merklesync_grpc.pb.go
) else (
    echo ❌ Failed to generate protobuf files
    exit /b 1
)
