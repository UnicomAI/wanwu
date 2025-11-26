# Deploy Static Assets 404 Fix

## Quick Deploy Instructions

Follow these steps to fix the static assets 404 error on your Dokploy deployment:

### Step 1: Start Docker (if needed)
```bash
# On macOS, ensure Docker Desktop is running
open -a Docker

# Wait for Docker to start (about 30 seconds)
# You can verify with:
docker ps
```

### Step 2: Build and Push the Backend Image

**Option A: Using the rebuild script (Recommended)**
```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu
./rebuild-and-push.sh backend
```

**Option B: Manual build and push**
```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Copy the correct dockerignore
cp .dockerignore .dockerignore.backup
cp .dockerignore.backend .dockerignore

# Build the image
docker build --platform linux/amd64 \
  -f Dockerfile.backend \
  -t safvr/wanwu-backend:latest \
  -t safvr/wanwu-backend:v0.2.7 \
  --build-arg WANWU_ARCH=amd64 \
  --build-arg WANWU_VERSION=v0.2.7 \
  .

# Restore original dockerignore
mv .dockerignore.backup .dockerignore

# Push to Docker Hub
docker push safvr/wanwu-backend:latest
docker push safvr/wanwu-backend:v0.2.7
```

### Step 3: Commit and Push Changes

```bash
git add Dockerfile.backend STATIC_ASSETS_404_FIX.md DEPLOY_STATIC_ASSETS_FIX.md
git commit -m "Fix: Include all configs in backend image for static asset serving

- Modified Dockerfile.backend to copy entire configs directory
- Ensures static files (icons, logos) are included in the image
- Fixes 404 errors for /user/api/v1/static/* resources on Dokploy
- Static files now work regardless of volume mount configuration"

git push origin docs/translation-plan
# Or push to your main branch
```

### Step 4: Deploy to Dokploy

You have two options:

#### Option A: Automatic Deployment via CI/CD (Recommended)

If your GitHub repo is connected to Dokploy with auto-deploy:

1. **Merge to main branch** (if not already there):
   ```bash
   git checkout main
   git merge docs/translation-plan
   git push origin main
   ```

2. **Wait for GitHub Actions**:
   - Go to: https://github.com/YOUR_REPO/actions
   - Watch the build and deployment pipeline
   - It will automatically rebuild and deploy to Dokploy

#### Option B: Manual Dokploy Deployment

1. **Login to Dokploy**:
   - Open your Dokploy dashboard
   - Navigate to your compose deployment

2. **Trigger Redeploy**:
   - Click on "Redeploy" or "Rebuild"
   - Dokploy will pull the new `safvr/wanwu-backend:latest` image
   - Wait for deployment to complete

3. **Alternative - SSH to server**:
   ```bash
   ssh ubuntu@34.238.142.181
   cd /path/to/your/dokploy/project
   
   # Pull the new image
   docker pull safvr/wanwu-backend:latest
   
   # Restart the service
   docker-compose restart bff-service
   
   # Or full restart if needed
   docker-compose down bff-service
   docker-compose up -d bff-service
   ```

### Step 5: Verify the Fix

#### 5.1 Check Container Has Files
```bash
# SSH into your Dokploy server
ssh ubuntu@34.238.142.181

# Check if static files exist in container
docker exec bff-service ls -la /app/configs/microservice/bff-service/static/icon/
docker exec bff-service ls -la /app/configs/microservice/bff-service/static/logo/
```

Expected output:
```
-rwxr-xr-x  user-default-icon.png  (5032 bytes)
-rwxr-xr-x  workflow-default-icon.png  (1422 bytes)
# ... more icons

-rw-r--r--  title_logo.png  (386431 bytes)
-rw-r--r--  login_logo.png  (68197 bytes)
# ... more logos
```

#### 5.2 Test API Directly
```bash
# From your local machine
curl -I https://app.safvr.com/user/api/v1/static/icon/user-default-icon.png

# Should return:
HTTP/2 200 
content-type: image/png
content-length: 5032
```

#### 5.3 Test in Browser
1. Open: https://app.safvr.com
2. Clear cache: Ctrl+Shift+R (Windows/Linux) or Cmd+Shift+R (Mac)
3. Open DevTools Console (F12)
4. Check Network tab - no 404 errors for static assets
5. Verify images display correctly

### Step 6: Restart nginx if needed

If images still don't load after backend restart, restart nginx:

```bash
# On Dokploy server
docker-compose restart nginx-wanwu

# Or just nginx service
docker restart nginx-wanwu
```

## Troubleshooting

### Issue: Docker build fails

**Error**: "Cannot connect to Docker daemon"
```bash
# Solution: Start Docker Desktop
open -a Docker

# Wait 30 seconds, then retry
./rebuild-and-push.sh backend
```

**Error**: "no space left on device"
```bash
# Clean up Docker
docker system prune -a
docker volume prune

# Then retry build
```

### Issue: Push to Docker Hub fails

**Error**: "denied: requested access to the resource is denied"
```bash
# Login to Docker Hub
docker login

# Enter your Docker Hub credentials:
# Username: safvr
# Password: [your token]

# Then retry push
docker push safvr/wanwu-backend:latest
```

### Issue: Dokploy doesn't pull new image

```bash
# SSH to Dokploy server
ssh ubuntu@34.238.142.181

# Force pull new image
docker pull safvr/wanwu-backend:latest --no-cache

# Check image creation time
docker image inspect safvr/wanwu-backend:latest | grep Created

# Restart with new image
docker-compose up -d --force-recreate bff-service
```

### Issue: Still getting 404 after deployment

1. **Verify image was updated**:
   ```bash
   docker exec bff-service ls -la /app/configs/
   # Should show: microservice, middleware (both directories)
   ```

2. **Check bff-service logs**:
   ```bash
   docker logs bff-service | tail -100
   ```

3. **Test from inside nginx container**:
   ```bash
   docker exec nginx-wanwu curl -I http://bff-service:6668/v1/static/icon/user-default-icon.png
   # Should return: HTTP/1.1 200 OK
   ```

4. **Restart both services**:
   ```bash
   docker-compose restart bff-service nginx-wanwu
   ```

### Issue: Volume mount still overriding

If you want to completely remove volume mount dependency:

1. **Edit docker-compose.yaml on server**:
   ```yaml
   bff-service:
     volumes:
       # Comment out this line:
       # - ./configs:/app/configs:ro
       
       # Keep these:
       - ${WANWU_PROJECT_DIR}/bff-service/log:/app/log
       - ${WANWU_PROJECT_DIR}/bff-service/tmp:/app/tmp
   ```

2. **Redeploy**:
   ```bash
   docker-compose up -d bff-service
   ```

## Expected Results

After successful deployment:

✅ **No 404 errors** for static assets in browser console
✅ **User icon loads**: `https://app.safvr.com/user/api/v1/static/icon/user-default-icon.png`
✅ **Title logo loads**: `https://app.safvr.com/user/api/v1/static/logo/title_logo.png`
✅ **All other static assets load** without errors
✅ **Container has files**: `/app/configs/microservice/bff-service/static/*`

## Quick Command Summary

```bash
# 1. Start Docker
open -a Docker

# 2. Build and push
cd /Users/mohankumarv/Desktop/SAFVR/wanwu
./rebuild-and-push.sh backend

# 3. Commit and push
git add Dockerfile.backend *.md
git commit -m "Fix: Include all configs in backend image for static assets"
git push

# 4. Deploy on Dokploy server
ssh ubuntu@34.238.142.181
docker pull safvr/wanwu-backend:latest
docker-compose restart bff-service nginx-wanwu

# 5. Test
curl -I https://app.safvr.com/user/api/v1/static/icon/user-default-icon.png
```

## Need Help?

Check these files:
- `STATIC_ASSETS_404_FIX.md` - Detailed explanation of the issue and fix
- `Dockerfile.backend` - The modified Dockerfile
- `docker-compose.yaml` - Volume mount configuration
- `configs/middleware/nginx/conf.d/aibase.conf` - Nginx proxy configuration

Contact: [Your contact info]

