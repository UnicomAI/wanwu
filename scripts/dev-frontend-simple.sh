#!/bin/bash

# Simple Frontend Development Script
# Builds and deploys frontend to nginx container

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
WEB_DIR="$PROJECT_DIR/web"
NGINX_CONTAINER="nginx-wanwu"

echo "🚀 Frontend Development - Build & Deploy"
echo ""

# Check if Docker is running
if ! docker ps >/dev/null 2>&1; then
    echo "❌ Error: Docker is not running. Please start Docker Desktop."
    exit 1
fi

# Check if nginx container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${NGINX_CONTAINER}$"; then
    echo "❌ Error: nginx-wanwu container is not running."
    echo "Please start it with: docker compose --env-file .env --env-file .env.image.arm64 up -d"
    exit 1
fi

# Build
echo "🔨 Building frontend..."
cd "$WEB_DIR"
npm run build

# Deploy
echo "📦 Deploying to nginx container..."
docker cp dist/. "${NGINX_CONTAINER}:/usr/share/nginx/html/aibase/"
docker exec "$NGINX_CONTAINER" nginx -s reload

echo ""
echo "✅ Frontend deployed successfully!"
echo "🌐 Open http://localhost:8081/aibase/ in your browser"
echo ""
echo "💡 To rebuild and deploy again, just run this script:"
echo "   ./scripts/dev-frontend-simple.sh"
