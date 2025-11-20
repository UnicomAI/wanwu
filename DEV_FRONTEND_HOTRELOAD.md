# Frontend Hot-Reload Development Guide

This guide shows you how to develop the frontend with automatic hot-reload, so you don't need to rebuild Docker images every time you make changes.

## Prerequisites

1. Docker Desktop must be running
2. The main app must be running (all backend services)

## Method 1: Watch Mode with Auto-Copy (Recommended)

This method automatically rebuilds and copies files to the nginx container whenever you save changes.

### Step 1: Start the main app

```bash
docker compose --env-file .env --env-file .env.image.arm64 up -d
```

### Step 2: Start the watch script

Open a new terminal and run:

```bash
cd web
npm run build -- --watch &
WATCH_PID=$!

# Watch for changes and auto-copy to nginx
while true; do
  inotifywait -r -e modify,create,delete ./dist 2>/dev/null || fswatch -o ./dist | while read; do
    echo "Changes detected, copying to nginx..."
    docker cp dist/. nginx-wanwu:/usr/share/nginx/html/aibase/
    docker exec nginx-wanwu nginx -s reload
    echo "✅ Frontend updated!"
  done
done
```

For Mac (using fswatch):
```bash
# Install fswatch if not already installed
brew install fswatch

# Run the watch script
cd /Users/mohankumarv/Desktop/SAFVR/wanwu
./scripts/dev-frontend-watch.sh
```

### Step 3: Develop

Just edit your Vue files in `web/src/` and save. The changes will automatically:
1. Trigger a rebuild
2. Copy to the nginx container
3. Reload nginx
4. Be visible in your browser (after refresh)

## Method 2: Development Server (Port 8080)

This is the fastest method but runs on a different port.

### Step 1: Start the main app (backend services)

```bash
docker compose --env-file .env --env-file .env.image.arm64 up -d
```

### Step 2: Start the dev server

```bash
cd web
npm run serve
```

### Step 3: Access the app

Open http://localhost:8080/aibase/ in your browser.

**Pros:**
- ⚡ Instant hot-reload (no manual refresh needed)
- 🔥 Fastest development experience
- 🐛 Better error messages

**Cons:**
- Different port (8080 instead of 8081)
- May have CORS issues with some APIs

## Method 3: Volume Mount (Requires Docker Restart)

This mounts your local `web/dist` folder directly into nginx.

### Step 1: Build the frontend once

```bash
cd web
npm run build
```

### Step 2: Start with dev override

```bash
docker compose --env-file .env --env-file .env.image.arm64 -f docker-compose.yaml -f docker-compose.dev.yaml up -d
```

### Step 3: Rebuild when you make changes

```bash
cd web
npm run build
# Changes are immediately visible (just refresh browser)
```

**Pros:**
- Same port (8081)
- No need to copy files

**Cons:**
- Still need to run `npm run build` after each change
- Slower than dev server

## Recommended Workflow

For active development:
1. Use **Method 2** (dev server on 8080) for fast iteration
2. Test on **Method 1 or 3** (port 8081) before committing

For production deployment:
1. Build the Docker image: `make docker-image-frontend`
2. Update `.env` with the new image tag
3. Restart nginx: `docker compose restart nginx`

## Troubleshooting

### Changes not showing up?
1. Hard refresh browser: `Cmd+Shift+R` (Mac) or `Ctrl+Shift+R` (Windows/Linux)
2. Clear browser cache
3. Check if nginx reloaded: `docker logs nginx-wanwu`

### Build errors?
1. Clear node_modules: `cd web && rm -rf node_modules && npm install`
2. Clear dist: `rm -rf web/dist`
3. Rebuild: `cd web && npm run build`

### Docker issues?
1. Check Docker is running: `docker ps`
2. Check logs: `docker logs nginx-wanwu`
3. Restart nginx: `docker restart nginx-wanwu`
