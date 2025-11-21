#!/bin/bash
# Permanent fix for WanWu startup - ensures services start in correct order
# and English is set as default language

set -e

echo "🚀 Starting WanWu Application with English as default..."

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Check architecture
ARCH=$(grep "WANWU_ARCH=" .env | cut -d'=' -f2)
echo "📌 Architecture: $ARCH"

# Stop all containers
echo "🛑 Stopping all containers..."
docker-compose -f docker-compose.yaml --env-file .env.image.$ARCH --env-file .env down 2>/dev/null || true

# Start infrastructure services first
echo "📦 Starting infrastructure (MySQL, Redis, MinIO, Kafka, Elasticsearch)..."
docker-compose -f docker-compose.yaml --env-file .env.image.$ARCH --env-file .env up -d mysql redis minio

# Wait for MySQL to be healthy
echo "⏳ Waiting for MySQL to be healthy..."
for i in {1..60}; do
    health=$(docker inspect mysql-wanwu --format='{{.State.Health.Status}}' 2>/dev/null || echo "starting")
    if [ "$health" = "healthy" ]; then
        echo "✅ MySQL is healthy!"
        break
    fi
    echo "  [$i/60] MySQL status: $health"
    sleep 2
done

# Start remaining infrastructure
echo "📦 Starting Kafka and Elasticsearch..."
docker-compose -f docker-compose.yaml --env-file .env.image.$ARCH --env-file .env up -d 2>/dev/null || docker-compose -f docker-compose.yaml --env-file .env.image.$ARCH --env-file .env up -d

# Wait a bit for Kafka/ES
sleep 10

# Start all backend services
echo "🔧 Starting backend services..."
docker-compose -f docker-compose.yaml --env-file .env.image.$ARCH --env-file .env up -d

# Wait for services to stabilize
echo "⏳ Waiting for all services to be healthy..."
sleep 30

# Restart nginx to ensure it has fresh DNS resolution
echo "🔄 Restarting nginx..."
docker restart nginx-wanwu
sleep 5

# Show status
echo ""
echo "📊 Service Status:"
docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "(mysql|bff|nginx|iam|knowledge|app)"

echo ""
echo "✅ WanWu started successfully!"
echo "🌐 Access at: http://localhost:8081/aibase/"
echo "🇬🇧 Default language: English"
echo ""
echo "📝 To clear browser cache and force English:"
echo "   1. Open DevTools (F12)"
echo "   2. Console: localStorage.clear(); localStorage.setItem('locale', 'en'); location.reload();"
