#!/bin/bash

# Frontend Hot-Reload Watch Script
# This script watches for changes in web/dist and automatically copies them to nginx container

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
WEB_DIR="$PROJECT_DIR/web"
DIST_DIR="$WEB_DIR/dist"
NGINX_CONTAINER="nginx-wanwu"

echo "🚀 Starting Frontend Hot-Reload Watch Script"
echo "📁 Project Dir: $PROJECT_DIR"
echo "📁 Web Dir: $WEB_DIR"
echo "📁 Dist Dir: $DIST_DIR"
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

# Check if fswatch is installed
if ! command -v fswatch &> /dev/null; then
    echo "⚠️  fswatch is not installed. Installing via Homebrew..."
    if command -v brew &> /dev/null; then
        brew install fswatch
    else
        echo "❌ Error: Homebrew is not installed. Please install fswatch manually:"
        echo "   brew install fswatch"
        exit 1
    fi
fi

# Initial build
echo "🔨 Building frontend for the first time..."
cd "$WEB_DIR"
npm run build

# Initial copy
echo "📦 Copying initial build to nginx..."
docker cp "$DIST_DIR/." "${NGINX_CONTAINER}:/usr/share/nginx/html/aibase/"
docker exec "$NGINX_CONTAINER" nginx -s reload
echo "✅ Initial deployment complete!"
echo ""

# Start build in watch mode
echo "👀 Starting build in watch mode..."
npm run build -- --watch &
BUILD_PID=$!

# Cleanup function
cleanup() {
    echo ""
    echo "🛑 Stopping watch script..."
    kill $BUILD_PID 2>/dev/null || true
    echo "✅ Cleanup complete"
    exit 0
}

trap cleanup SIGINT SIGTERM

# Watch for changes in dist directory
echo "👀 Watching for changes in $DIST_DIR..."
echo "💡 Press Ctrl+C to stop"
echo ""

fswatch -o "$DIST_DIR" | while read change; do
    echo "🔄 Changes detected at $(date '+%H:%M:%S')"
    echo "📦 Copying to nginx container..."
    
    if docker cp "$DIST_DIR/." "${NGINX_CONTAINER}:/usr/share/nginx/html/aibase/"; then
        if docker exec "$NGINX_CONTAINER" nginx -s reload 2>/dev/null; then
            echo "✅ Frontend updated successfully!"
        else
            echo "⚠️  Warning: nginx reload failed, but files were copied"
        fi
    else
        echo "❌ Error: Failed to copy files to nginx"
    fi
    echo ""
done
