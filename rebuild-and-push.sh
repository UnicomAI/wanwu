#!/bin/bash

# Smart Rebuild and Push Script with Change Detection
# Usage: ./rebuild-and-push.sh [backend|frontend|rag|agent|all|auto]
# Use 'auto' to detect changes against origin/main

set -e

PROJECT_ROOT="/Users/mohankumarv/Desktop/SAFVR/wanwu"
DOCKER_USER="safvr"
WANWU_ARCH="amd64"
WANWU_VERSION="v0.2.7"

cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

echo_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# Function to detect changed files
detect_changes() {
    echo_info "Detecting changed files..."
    
    # Get the diff between current branch and origin/main
    CHANGED_FILES=$(git diff --name-only origin/main...HEAD 2>/dev/null || git diff --name-only HEAD~1 2>/dev/null || echo "")
    
    if [ -z "$CHANGED_FILES" ]; then
        echo_warn "No changes detected from origin/main, checking uncommitted changes..."
        CHANGED_FILES=$(git diff --name-only)
        if [ -z "$CHANGED_FILES" ]; then
            echo_warn "No uncommitted changes either"
            return
        fi
    fi
    
    echo_debug "Changed files:"
    echo "$CHANGED_FILES" | while read file; do
        echo "  - $file"
    done
}

# Function to check if specific paths changed
should_build() {
    local service=$1
    local patterns=("$@")
    
    CHANGED_FILES=$(git diff --name-only origin/main...HEAD 2>/dev/null || git diff --name-only HEAD~1 2>/dev/null || git diff --name-only)
    
    for pattern in "${patterns[@]:1}"; do
        if echo "$CHANGED_FILES" | grep -q "^$pattern"; then
            return 0
        fi
    done
    return 1
}

# Function to build and push a service
build_and_push() {
    local service=$1
    local dockerfile=$2
    local dockerignore=$3
    local image_name=$4
    local build_args=${5:-""}
    
    echo_info "Building $service image..."
    cp .dockerignore .dockerignore.backup
    cp "$dockerignore" .dockerignore
    
    BUILD_CMD="docker build --platform linux/$WANWU_ARCH \
        -f $dockerfile \
        -t ${DOCKER_USER}/${image_name}:latest \
        -t ${DOCKER_USER}/${image_name}:$WANWU_VERSION"
    
    if [ -n "$build_args" ]; then
        BUILD_CMD="$BUILD_CMD $build_args"
    fi
    
    BUILD_CMD="$BUILD_CMD ."
    
    eval "$BUILD_CMD"
    
    mv .dockerignore.backup .dockerignore
    
    echo_info "Pushing ${service} image to Docker Hub..."
    docker push ${DOCKER_USER}/${image_name}:latest
    docker push ${DOCKER_USER}/${image_name}:$WANWU_VERSION
    
    echo_info "✅ $service image built and pushed successfully"
}

# Main logic
SERVICE=${1:-auto}
SERVICES_TO_BUILD=()

if [ "$SERVICE" = "auto" ]; then
    echo_info "Auto-detecting changes..."
    detect_changes
    
    if should_build "backend" "internal" "cmd" "pkg" "proto" "api" "Dockerfile.backend" "go.mod" "go.sum"; then
        echo_warn "Backend changes detected"
        SERVICES_TO_BUILD+=("backend")
    else
        echo_info "No backend changes detected"
    fi
    
    if should_build "frontend" "web" "Dockerfile.frontend" "configs/middleware/nginx/conf.d"; then
        echo_warn "Frontend changes detected"
        SERVICES_TO_BUILD+=("frontend")
    else
        echo_info "No frontend changes detected"
    fi
    
    if should_build "rag" "rag" "Dockerfile.rag" "Dockerfile.rag-base"; then
        echo_warn "RAG changes detected"
        SERVICES_TO_BUILD+=("rag")
    else
        echo_info "No RAG changes detected"
    fi
    
    if should_build "agent" "agent" "Dockerfile.agent" "Dockerfile.agent-base"; then
        echo_warn "Agent changes detected"
        SERVICES_TO_BUILD+=("agent")
    else
        echo_info "No agent changes detected"
    fi
    
    if [ ${#SERVICES_TO_BUILD[@]} -eq 0 ]; then
        echo_warn "No changes detected for any service"
        exit 0
    fi
else
    # Use explicitly specified services
    case $SERVICE in
        backend|frontend|rag|agent|all)
            if [ "$SERVICE" = "all" ]; then
                SERVICES_TO_BUILD=("backend" "frontend" "rag" "agent")
            else
                SERVICES_TO_BUILD=("$SERVICE")
            fi
            ;;
        *)
            echo_error "Invalid service: $SERVICE"
            echo "Usage: $0 [backend|frontend|rag|agent|all|auto]"
            exit 1
            ;;
    esac
fi

echo ""
echo_info "=========================================="
echo_info "Services to build:"
for svc in "${SERVICES_TO_BUILD[@]}"; do
    echo "  - $svc"
done
echo_info "=========================================="
echo ""

# Build services
for svc in "${SERVICES_TO_BUILD[@]}"; do
    case $svc in
        backend)
            build_and_push "backend" "Dockerfile.backend" ".dockerignore.backend" "wanwu-backend" "--build-arg WANWU_ARCH=$WANWU_ARCH --build-arg WANWU_VERSION=$WANWU_VERSION"
            ;;
        frontend)
            build_and_push "frontend" "Dockerfile.frontend" ".dockerignore.frontend" "wanwu-frontend" "--build-arg WANWU_ARCH=$WANWU_ARCH"
            ;;
        rag)
            echo_info "Building RAG base image first..."
            build_and_push "rag-base" "Dockerfile.rag-base" ".dockerignore.rag" "rag-base"
            build_and_push "rag" "Dockerfile.rag" ".dockerignore.rag" "wanwu-rag"
            ;;
        agent)
            echo_info "Building agent base image first..."
            build_and_push "agent-base" "Dockerfile.agent-base" ".dockerignore.agent-base" "agent-base"
            build_and_push "agent" "Dockerfile.agent" ".dockerignore.agent-base" "agent" "--build-arg WANWU_VERSION=$WANWU_VERSION"
            ;;
    esac
done

echo ""
echo_info "=========================================="
echo_info "✅ Build and push completed successfully!"
echo_info "=========================================="
echo ""
echo_info "Next steps:"
echo "1. Push changes to remote: git push origin <branch>"
echo "2. GitHub Actions will automatically trigger the CI/CD pipeline"
echo "3. Or manually trigger from: https://github.com/UnicomAI/wanwu/actions"
echo ""
