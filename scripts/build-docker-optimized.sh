#!/bin/bash

echo "🐳 Building Docker Images (Optimized)"
echo "====================================="

# Set Docker build options
export DOCKER_BUILDKIT=1
export DOCKER_DEFAULT_PLATFORM=linux/amd64

echo "[INFO] Building server image..."
docker build --no-cache -f Dockerfile.server -t epic-merklesync-server:latest .
if [ $? -ne 0 ]; then
    echo "[ERROR] Server build failed"
    exit 1
fi

echo "[INFO] Building edge client image..."
docker build --no-cache -f Dockerfile.edge-client -t epic-edge-client:latest .
if [ $? -ne 0 ]; then
    echo "[ERROR] Edge client build failed"
    exit 1
fi

echo "[INFO] Building PostgreSQL connector image..."
docker build --no-cache -f Dockerfile.postgresql-connector -t epic-postgresql-connector:latest .
if [ $? -ne 0 ]; then
    echo "[ERROR] PostgreSQL connector build failed"
    exit 1
fi

echo "[INFO] Building MongoDB connector image..."
docker build --no-cache -f Dockerfile.mongodb-connector -t epic-mongodb-connector:latest .
if [ $? -ne 0 ]; then
    echo "[ERROR] MongoDB connector build failed"
    exit 1
fi

echo "[SUCCESS] All Docker images built successfully!"
echo ""
echo "Next steps:"
echo "  1. Start services: docker-compose up -d"
echo "  2. Check status: docker-compose ps"
echo "  3. View logs: docker-compose logs -f"
