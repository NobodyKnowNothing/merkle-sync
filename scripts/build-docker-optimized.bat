@echo off
echo 🐳 Building Docker Images (Optimized)
echo =====================================

REM Set Docker build options
set DOCKER_BUILDKIT=1
set DOCKER_DEFAULT_PLATFORM=linux/amd64

echo [INFO] Building server image...
echo [DEBUG] Command: docker build --no-cache -f Dockerfile.server -t epic-merklesync-server:latest .
docker build --no-cache -f Dockerfile.server -t epic-merklesync-server:latest .
if %errorlevel% neq 0 (
    echo [ERROR] Server build failed
    echo [DEBUG] Check Docker logs and processes
    docker ps -a
    exit /b 1
)

echo [INFO] Server image built successfully!
echo [INFO] Building edge client image...
docker build --no-cache -f Dockerfile.edge-client -t epic-edge-client:latest .
if %errorlevel% neq 0 (
    echo [ERROR] Edge client build failed
    exit /b 1
)

echo [INFO] Edge client image built successfully!
echo [INFO] Building PostgreSQL connector image...
docker build --no-cache -f Dockerfile.postgresql-connector -t epic-postgresql-connector:latest .
if %errorlevel% neq 0 (
    echo [ERROR] PostgreSQL connector build failed
    exit /b 1
)

echo [INFO] PostgreSQL connector image built successfully!
echo [INFO] Building MongoDB connector image...
docker build --no-cache -f Dockerfile.mongodb-connector -t epic-mongodb-connector:latest .
if %errorlevel% neq 0 (
    echo [ERROR] MongoDB connector build failed
    exit /b 1
)

echo [SUCCESS] All Docker images built successfully!
echo.
echo Next steps:
echo   1. Start services: docker-compose up -d
echo   2. Check status: docker-compose ps
echo   3. View logs: docker-compose logs -f
