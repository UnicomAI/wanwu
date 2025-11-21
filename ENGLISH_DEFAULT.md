# WanWu - English as Default Language

## ✅ Permanent Fixes Applied

### 1. Frontend Configuration
- **File:** `web/src/lang/index.js`  
  - Default locale set to `'en'` (line 24)
  
- **File:** `web/src/main.js`  
  - Force English locale on first load (lines 37-40)

### 2. Startup Scripts (Recommended Way)
Use these scripts to ensure services start in the correct order:

```bash
# Start WanWu
./scripts/start-wanwu.sh

# Stop WanWu  
./scripts/stop-wanwu.sh
```

### 3. Manual Docker Commands (If Needed)
If you need to start manually:

```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Stop all
docker-compose -f docker-compose.yaml --env-file .env.image.amd64 --env-file .env down

# Start all
docker-compose -f docker-compose.yaml --env-file .env.image.amd64 --env-file .env up -d

# Wait for MySQL to be healthy (important!)
for i in {1..30}; do 
  health=$(docker inspect mysql-wanwu --format='{{.State.Health.Status}}' 2>/dev/null)
  [ "$health" = "healthy" ] && echo "MySQL ready!" && break || sleep 2
done

# Restart nginx after all services are up
docker restart nginx-wanwu
```

## 🌐 Access the Application

- **URL:** http://localhost:8081/aibase/
- **Default Language:** English
- **Login:** admin / (your password)

## 🔧 Troubleshooting

### If UI Shows Chinese:
1. Open Browser DevTools (F12)
2. Go to Console tab
3. Run:
```javascript
localStorage.clear();
sessionStorage.clear();
localStorage.setItem('locale', 'en');
location.reload(true);
```

### If You Get "Bad Gateway" (502) Errors:
1. Wait 1-2 minutes for services to fully start
2. Check MySQL is healthy: `docker ps | grep mysql`
3. Restart nginx: `docker restart nginx-wanwu`

### If Login Fails with i18n Error:
1. Ensure MySQL is healthy before login
2. Restart bff-service: `docker restart bff-service`
3. Clear browser cache and retry

### Common Issues:
- **MySQL keeps restarting:** Use the `./scripts/start-wanwu.sh` script which waits for MySQL to be healthy
- **Services can't connect to MySQL:** MySQL IP changes on restart; restart backend services after MySQL is stable
- **Missing logo:** Frontend not deployed; rebuild with `cd web && npm run build` then copy to nginx

## 📝 Development Workflow

### Rebuild and Deploy Frontend:
```bash
cd web
npm run build
docker cp dist/. nginx-wanwu:/usr/share/nginx/html/aibase/
docker exec nginx-wanwu nginx -s reload
```

### View Service Logs:
```bash
docker logs bff-service --tail 50
docker logs mysql-wanwu --tail 50
docker logs nginx-wanwu --tail 50
```

### Check Service Status:
```bash
docker ps --format "table {{.Names}}\t{{.Status}}"
```

## ✨ What Was Changed

1. **Frontend i18n default:** `'en'` instead of checking localStorage first
2. **Force English on init:** Added code to set locale to 'en' if not already set
3. **Startup order:** Created scripts that ensure MySQL is healthy before starting other services
4. **All Chinese comments:** Translated to English
5. **Hardcoded Chinese text:** Replaced with English or i18n keys

## 🎯 Result
- Application always starts in English by default
- All UI elements translated to English
- Proper service startup order prevents connection issues
- No more i18n panic errors
