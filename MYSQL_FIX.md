# MySQL Stability Fix

## Problem
MySQL container restarts every 2 minutes, causing:
- Login fails with i18n panic error
- Chinese UI (can't load English config from DB)
- Backend services lose database connection

## Root Cause
Docker resource constraints or health check issues causing MySQL to restart.

## Solution Options

### Option 1: Increase Docker Resources (Recommended)
1. Open Docker Desktop
2. Go to Settings → Resources
3. Increase:
   - **Memory**: At least 8GB (currently need more)
   - **CPU**: At least 4 cores
   - **Swap**: 2GB
4. Click "Apply & Restart"
5. Restart WanWu: `./scripts/start-wanwu.sh`

### Option 2: Use External MySQL (Production-Ready)
Instead of Docker MySQL, use a dedicated MySQL instance:

1. Install MySQL 8.0 locally or use cloud MySQL
2. Update `.env`:
```bash
WANWU_MYSQL_HOST=localhost  # or your MySQL host
WANWU_MYSQL_PORT=3306
WANWU_MYSQL_PASSWORD=your_password
```

3. Initialize database:
```bash
mysql -u root -p < configs/middleware/mysql/initdb.d/init.sql
```

4. Comment out MySQL in docker-compose.yaml
5. Start services: `./scripts/start-wanwu.sh`

### Option 3: Disable Health Check Temporarily
Edit `docker-compose.yaml`, find the mysql section and comment out healthcheck:

```yaml
mysql:
  restart: always
  # healthcheck:
  #   test: ["CMD-SHELL", "mysqladmin ping..."]
  #   ...
```

Then restart: `./scripts/start-wanwu.sh`

## Verify MySQL is Stable

```bash
# Watch MySQL status (should stay "Up" without restarting)
watch -n 2 'docker ps --filter name=mysql --format "{{.Names}}: {{.Status}}"'

# Check restart count (should stay at 0)
docker inspect mysql-wanwu --format='{{.RestartCount}} restarts'
```

## After MySQL is Stable

1. **Test Backend API:**
```bash
curl http://localhost:8081/user/api/v1/base/custom
# Should return JSON, not HTML
```

2. **Clear Browser Cache:**
- Open http://localhost:8081/aibase/
- Press F12 → Console
- Run:
```javascript
localStorage.clear();
sessionStorage.clear();  
localStorage.setItem('locale', 'en');
location.reload(true);
```

3. **Login** - Should work with English UI

## Frontend English Changes (Already Applied)

These changes are complete and will work once MySQL is stable:

- `web/src/lang/index.js`: Default locale = 'en', forces English on startup
- `web/src/main.js`: Overrides Chinese locale to English
- All Chinese text translated
- Built and deployed to nginx

The issue is **not** the frontend - it's MySQL stability affecting the backend.
