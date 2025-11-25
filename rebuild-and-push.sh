#!/bin/bash

# Rebuild and Push Script for Dokploy Deployment
# Usage: ./rebuild-and-push.sh [backend|frontend|rag|agent|all]

set -e

PROJECT_ROOT="/Users/mohankumarv/Desktop/SAFVR/wanwu"
DOCKER_USER="safvr"

cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to build backend
build_backend() {
    echo_info "Building backend image..."
    cp .dockerignore .dockerignore.backup
    cp .dockerignore.backend .dockerignore
    
    docker build --platform linux/amd64 \
        --build-arg WANWU_ARCH=amd64 \
        -f Dockerfile.backend \
        -t ${DOCKER_USER}/wanwu-backend:latest .
    
    mv .dockerignore.backup .dockerignore
    echo_info "✅ Backend image built successfully"
}

# Function to build frontend
build_frontend() {
    echo_info "Building frontend image..."
    cp .dockerignore .dockerignore.backup
    cp .dockerignore.frontend .dockerignore
    
    docker build --platform linux/amd64 \
        --build-arg WANWU_ARCH=amd64 \
        -f Dockerfile.frontend \
        -t ${DOCKER_USER}/wanwu-frontend:latest .
    
    mv .dockerignore.backup .dockerignore
    echo_info "✅ Frontend image built successfully"
}

# Function to build rag
build_rag() {
    echo_info "Building RAG image..."
    cp .dockerignore .dockerignore.backup
    cp .dockerignore.rag .dockerignore
    
    docker build --platform linux/amd64 \
        -f Dockerfile.rag \
        -t ${DOCKER_USER}/wanwu-rag:latest .
    
    mv .dockerignore.backup .dockerignore
    echo_info "✅ RAG image built successfully"
}

# Function to build agent
build_agent() {
    echo_info "Building agent image..."
    cp .dockerignore .dockerignore.backup
    cp .dockerignore.agent-base .dockerignore
    
    docker build --platform linux/amd64 \
        -f Dockerfile.agent \
        -t ${DOCKER_USER}/agent:latest .
    
    mv .dockerignore.backup .dockerignore
    echo_info "✅ Agent image built successfully"
}

# Function to push images
push_images() {
    local service=$1
    echo_info "Pushing ${service} image to Docker Hub..."
    
    case $service in
        backend)
            docker push ${DOCKER_USER}/wanwu-backend:latest
            ;;
        frontend)
            docker push ${DOCKER_USER}/wanwu-frontend:latest
            ;;
        rag)
            docker push ${DOCKER_USER}/wanwu-rag:latest
            ;;
        agent)
            docker push ${DOCKER_USER}/agent:latest
            ;;
        all)
            docker push ${DOCKER_USER}/wanwu-backend:latest
            docker push ${DOCKER_USER}/wanwu-frontend:latest
            docker push ${DOCKER_USER}/wanwu-rag:latest
            docker push ${DOCKER_USER}/agent:latest
            ;;
    esac
    
    echo_info "✅ ${service} image pushed successfully"
}

# Main script
SERVICE=${1:-all}

echo_info "Starting rebuild for: $SERVICE"
echo_info "Project root: $PROJECT_ROOT"
echo_info "Docker user: $DOCKER_USER"
echo ""

case $SERVICE in
    backend)
        build_backend
        push_images backend
        ;;
    frontend)
        build_frontend
        push_images frontend
        ;;
    rag)
        build_rag
        push_images rag
        ;;
    agent)
        build_agent
        push_images agent
        ;;
    all)
        build_backend
        build_frontend
        build_rag
        build_agent
        push_images all
        ;;
    *)
        echo_error "Invalid service: $SERVICE"
        echo "Usage: $0 [backend|frontend|rag|agent|all]"
        exit 1
        ;;
esac

echo ""
echo_info "=========================================="
echo_info "✅ Build and push completed successfully!"
echo_info "=========================================="
echo ""
echo_info "Next steps:"
echo "1. Login to your Dokploy dashboard"
echo "2. Navigate to your WanWu application"
echo "3. Click 'Redeploy' or 'Restart' for the updated services"
echo ""
echo "Or SSH to your AWS instance and run:"
echo "  docker compose pull"
echo "  docker compose up -d"
echo ""
