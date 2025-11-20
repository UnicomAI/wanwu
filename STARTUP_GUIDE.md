# Wanwu Platform - Complete Startup Guide

## Prerequisites

- Docker Desktop installed and running
- At least 8GB RAM available
- Ports 8081, 3306, 9200, 9092, 6379, 9000 available

## Quick Start (After Computer Restart)

### 1. Start Docker Desktop

Make sure Docker Desktop is running. Check with:
```bash
docker ps
```

If you get an error, open Docker Desktop application.

### 2. Start All Services

```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# For ARM64 (M1/M2/M3 Mac)
docker compose --env-file .env --env-file .env.image.arm64 up -d

# For AMD64 (Intel Mac/Linux)
docker compose --env-file .env --env-file .env.image.amd64 up -d
```

### 3. Wait for Services to be Healthy

Check status:
```bash
docker ps
```

Wait until all services show `(healthy)` status. This may take 2-5 minutes.

### 4. Access the Platform

Open your browser and go to:
- **Main App**: http://localhost:8081/aibase/
- **Default Login**:
  - Username: `admin`
  - Password: `Wanwu123456`

---

## Development Workflow

### Option A: Quick Development (Recommended for Fast Iteration)

Use the development server on port 8080 with instant hot-reload:

```bash
cd web
npm run serve
```

Then open http://localhost:8080/aibase/

**Benefits:**
- ⚡ Instant hot-reload (no manual refresh)
- 🔥 Fastest development experience
- 🐛 Better error messages

### Option B: Production-Like Development (Port 8081)

#### Simple Method (Manual Rebuild)

After making frontend changes:

```bash
./scripts/dev-frontend-simple.sh
```

This will:
1. Build the frontend
2. Copy to nginx container
3. Reload nginx

Then refresh your browser at http://localhost:8081/aibase/

#### Advanced Method (Auto-Watch)

For automatic rebuilding and deployment:

```bash
./scripts/dev-frontend-watch.sh
```

This will:
1. Build the frontend initially
2. Watch for changes
3. Automatically rebuild and deploy
4. Reload nginx

Just edit your files and save - changes will be deployed automatically!

Press `Ctrl+C` to stop the watch script.

---

## Service Management

### Check Service Status

```bash
docker ps
```

### View Logs

```bash
# All services
docker compose logs -f

# Specific service
docker logs -f nginx-wanwu
docker logs -f bff-service
docker logs -f workflow-wanwu
```

### Restart a Service

```bash
# Restart nginx (frontend)
docker restart nginx-wanwu

# Restart backend
docker restart bff-service

# Restart all services
docker compose restart
```

### Stop All Services

```bash
# For ARM64
docker compose --env-file .env --env-file .env.image.arm64 down

# For AMD64
docker compose --env-file .env --env-file .env.image.amd64 down
```

### Stop and Remove All Data

```bash
# WARNING: This will delete all data!
docker compose --env-file .env --env-file .env.image.arm64 down -v
```

---

## Troubleshooting

### Docker Not Running

**Error**: `Cannot connect to the Docker daemon`

**Solution**: Start Docker Desktop application

### Port Already in Use

**Error**: `port is already allocated`

**Solution**: 
```bash
# Find what's using the port (e.g., 8081)
lsof -i :8081

# Kill the process
kill -9 <PID>
```

### Services Not Healthy

**Check logs**:
```bash
docker compose logs <service-name>
```

**Common issues**:
- MySQL not ready: Wait 1-2 minutes
- Elasticsearch not ready: Wait 2-3 minutes
- Out of memory: Increase Docker memory limit in Docker Desktop settings

### Frontend Changes Not Showing

1. **Hard refresh browser**: `Cmd+Shift+R` (Mac) or `Ctrl+Shift+R`
2. **Clear browser cache**
3. **Check if files were deployed**:
   ```bash
   docker exec nginx-wanwu ls -la /usr/share/nginx/html/aibase/
   ```
4. **Rebuild and deploy**:
   ```bash
   ./scripts/dev-frontend-simple.sh
   ```

### Backend Changes Not Working

If you modified backend code:

```bash
# Restart the backend service
docker restart bff-service

# Or rebuild if you changed Go code
make build-bff-arm64  # or build-bff-amd64
docker restart bff-service
```

---

## Development Tips

### 1. Keep Docker Desktop Running

Always keep Docker Desktop running in the background during development.

### 2. Use the Right Port

- **Port 8080**: Development server (hot-reload, fastest)
- **Port 8081**: Production nginx (requires rebuild)

### 3. Check Service Health

Before debugging, always check if all services are healthy:
```bash
docker ps
```

### 4. Monitor Logs

Keep a terminal open with logs:
```bash
docker compose logs -f
```

### 5. Regular Cleanup

Clean up unused Docker resources weekly:
```bash
docker system prune -a
```

---

## Quick Reference Commands

```bash
# Start everything
docker compose --env-file .env --env-file .env.image.arm64 up -d

# Stop everything
docker compose --env-file .env --env-file .env.image.arm64 down

# Restart nginx (frontend)
docker restart nginx-wanwu

# Restart backend
docker restart bff-service

# View logs
docker logs -f nginx-wanwu

# Deploy frontend changes (simple)
./scripts/dev-frontend-simple.sh

# Deploy frontend changes (auto-watch)
./scripts/dev-frontend-watch.sh

# Dev server (hot-reload)
cd web && npm run serve

# Check status
docker ps
```

---

## Next Steps

1. ✅ Start Docker Desktop
2. ✅ Run `docker compose up -d`
3. ✅ Wait for services to be healthy
4. ✅ Access http://localhost:8081/aibase/
5. ✅ Choose your development workflow (Option A or B)
6. 🚀 Start coding!

For more details, see:
- [DEV_FRONTEND_HOTRELOAD.md](./DEV_FRONTEND_HOTRELOAD.md) - Detailed frontend development guide
- [README.md](./README.md) - Full platform documentation
