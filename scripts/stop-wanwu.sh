#!/bin/bash
# Stop WanWu Application

set -e

echo "🛑 Stopping WanWu Application..."

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Check architecture
ARCH=$(grep "WANWU_ARCH=" .env | cut -d'=' -f2)

# Stop all containers
docker-compose -f docker-compose.yaml --env-file .env.image.$ARCH --env-file .env down

echo "✅ WanWu stopped successfully!"
