# Quick Start After Computer Restart

## 🚀 3-Step Startup

### Step 1: Start Docker Desktop
Open Docker Desktop application and wait for it to be ready.

### Step 2: Start All Services
```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu
docker compose --env-file .env --env-file .env.image.arm64 up -d
```

### Step 3: Access the Platform
Wait 2-3 minutes for services to be healthy, then open:
- http://localhost:8081/aibase/
- Login: `admin` / `Wanwu123456`

---

## 🔥 Frontend Development (Choose One)

### Option A: Hot-Reload Dev Server (Fastest)
```bash
cd web
npm run serve
```
Then open http://localhost:8080/aibase/
- ⚡ Instant hot-reload
- 🔥 No manual refresh needed

### Option B: Production Mode (Port 8081)

**Manual rebuild after changes:**
```bash
./scripts/dev-frontend-simple.sh
```

**Auto-rebuild (watch mode):**
```bash
./scripts/dev-frontend-watch.sh
```

---

## 📋 Quick Commands

```bash
# Check status
docker ps

# View logs
docker logs -f nginx-wanwu

# Restart frontend
docker restart nginx-wanwu

# Stop everything
docker compose --env-file .env --env-file .env.image.arm64 down
```

---

## 🆘 Troubleshooting

**Docker not running?**
→ Open Docker Desktop

**Port already in use?**
→ `lsof -i :8081` then `kill -9 <PID>`

**Changes not showing?**
→ Hard refresh: `Cmd+Shift+R`
→ Or run: `./scripts/dev-frontend-simple.sh`

---

For detailed documentation, see:
- [STARTUP_GUIDE.md](./STARTUP_GUIDE.md) - Complete guide
- [DEV_FRONTEND_HOTRELOAD.md](./DEV_FRONTEND_HOTRELOAD.md) - Frontend development
