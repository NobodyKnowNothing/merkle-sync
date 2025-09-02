@echo off
REM Universal MerkleSync System Test for Windows

echo 🚀 Universal MerkleSync System Test
echo ====================================

REM Check if Go is installed
echo [INFO] Checking Go installation...
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed. Please install Go 1.21 or later.
    echo Download from: https://golang.org/dl/
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo [SUCCESS] Go is installed: %GO_VERSION%

REM Check if Docker is installed
echo [INFO] Checking Docker installation...
where docker >nul 2>nul
if %errorlevel% neq 0 (
    echo [WARNING] Docker is not installed. Some tests will be skipped.
    set DOCKER_AVAILABLE=0
) else (
    where docker-compose >nul 2>nul
    if %errorlevel% neq 0 (
        echo [WARNING] Docker Compose is not installed. Some tests will be skipped.
        set DOCKER_AVAILABLE=0
    ) else (
        echo [SUCCESS] Docker and Docker Compose are installed
        set DOCKER_AVAILABLE=1
    )
)

REM Generate protobuf files
echo [INFO] Generating protobuf files...
where protoc >nul 2>nul
if %errorlevel% neq 0 (
    echo [WARNING] protoc is not installed. Skipping protobuf generation.
    set PROTO_AVAILABLE=0
) else (
    REM Install Go protobuf plugins if not present
    where protoc-gen-go >nul 2>nul
    if %errorlevel% neq 0 (
        echo [INFO] Installing protoc-gen-go...
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    )
    
    where protoc-gen-go-grpc >nul 2>nul
    if %errorlevel% neq 0 (
        echo [INFO] Installing protoc-gen-go-grpc...
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    )
    
    REM Generate protobuf files
    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/merklesync.proto
    
    if %errorlevel% equ 0 (
        echo [SUCCESS] Protobuf files generated successfully
        set PROTO_AVAILABLE=1
    ) else (
        echo [ERROR] Failed to generate protobuf files
        set PROTO_AVAILABLE=0
    )
)

REM Run Go tests
echo [INFO] Running Go tests...
echo [INFO] Downloading dependencies...
go mod tidy

go test -v ./core/... ./server/... ./edge-client/...
if %errorlevel% equ 0 (
    echo [SUCCESS] All Go tests passed
    set GO_TESTS_PASSED=1
) else (
    echo [ERROR] Some Go tests failed
    set GO_TESTS_PASSED=0
)

REM Build all components
echo [INFO] Building all components...
if not exist build mkdir build

echo [INFO] Building server...
go build -o build/merklesync-server.exe ./cmd/server

echo [INFO] Building PostgreSQL connector...
go build -o build/postgresql-connector.exe ./cmd/postgresql-connector

echo [INFO] Building MongoDB connector...
go build -o build/mongodb-connector.exe ./cmd/mongodb-connector

echo [INFO] Building edge client...
go build -o build/edge-client.exe ./cmd/edge-client

echo [SUCCESS] All components built successfully

REM Test Docker setup
if %DOCKER_AVAILABLE% equ 1 (
    echo [INFO] Testing Docker setup...
    echo [INFO] Building Docker images (optimized)...
    scripts\build-docker-optimized.bat
    
    if %errorlevel% equ 0 (
        echo [SUCCESS] Docker images built successfully
        set DOCKER_BUILD_PASSED=1
    ) else (
        echo [ERROR] Failed to build Docker images
        set DOCKER_BUILD_PASSED=0
    )
    
    echo [INFO] Validating Docker Compose configuration...
    docker-compose config
    
    if %errorlevel% equ 0 (
        echo [SUCCESS] Docker Compose configuration is valid
        set DOCKER_CONFIG_PASSED=1
    ) else (
        echo [ERROR] Docker Compose configuration is invalid
        set DOCKER_CONFIG_PASSED=0
    )
) else (
    set DOCKER_BUILD_PASSED=0
    set DOCKER_CONFIG_PASSED=0
)

REM Run integration test
echo [INFO] Running integration test...
go run examples/integration_demo.go
if %errorlevel% equ 0 (
    echo [SUCCESS] Integration test passed
    set INTEGRATION_TEST_PASSED=1
) else (
    echo [ERROR] Integration test failed
    set INTEGRATION_TEST_PASSED=0
)

REM Print summary
echo.
echo 📊 Test Summary
echo ===============
if %PROTO_AVAILABLE% equ 1 (
    echo   protobuf:✅
) else (
    echo   protobuf:⚠️
)

if %GO_TESTS_PASSED% equ 1 (
    echo   go-tests:✅
) else (
    echo   go-tests:❌
)

echo   build:✅

if %DOCKER_BUILD_PASSED% equ 1 (
    echo   docker:✅
) else (
    echo   docker:⚠️
)

if %INTEGRATION_TEST_PASSED% equ 1 (
    echo   integration:✅
) else (
    echo   integration:❌
)

REM Check if all critical tests passed
if %GO_TESTS_PASSED% equ 0 (
    echo [ERROR] Critical tests failed. Please fix the issues above.
    exit /b 1
) else (
    echo [SUCCESS] All critical tests passed! 🎉
    echo.
    echo 🚀 Universal MerkleSync is ready to use!
    echo.
    echo Next steps:
    echo   1. Start the system: docker-compose up -d
    echo   2. View logs: docker-compose logs -f
    echo   3. Run individual components: make dev-server
    echo.
)
