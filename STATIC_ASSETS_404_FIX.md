# Static Assets 404 Fix

## Problem

After deploying to Dokploy, static image assets are returning 404 errors:

```
GET https://app.safvr.com/user/api/v1/static/icon/user-default-icon.png 404 (Not Found)
GET https://app.safvr.com/user/api/v1/static/logo/title_logo.png 404 (Not Found)
```

## Root Cause

The issue is caused by a mismatch between the Docker image build and docker-compose volume mounts:

1. **Dockerfile.backend** originally copied only `configs/microservice`:
   ```dockerfile
   COPY configs/microservice ./configs/microservice
   ```

2. **docker-compose.yaml** has a volume mount that overrides the image:
   ```yaml
   volumes:
     - ./configs:/app/configs:ro
   ```

3. **On Dokploy deployments**:
   - The compose file uses a relative path `./configs`
   - This path doesn't exist or is incomplete on the Dokploy server
   - The volume mount overrides the image contents, making static files inaccessible

## How Static Files Should Work

### Request Flow:
1. Frontend request: `https://app.safvr.com/user/api/v1/static/icon/user-default-icon.png`
2. Nginx receives: `/user/api/v1/static/icon/user-default-icon.png`
3. Nginx proxy config strips `/user/api/` and forwards to: `http://bff-service:6668/v1/static/icon/user-default-icon.png`
4. Backend (Go) route: `/v1` group with `/static` handler
5. Backend serves from: `./configs/microservice/bff-service/static/icon/user-default-icon.png`

### Backend Routing Configuration:
```go
// internal/bff-service/server/http/handler/router/v1/guest.go
func registerGuest(apiV1 *gin.RouterGroup) {
    apiV1.Static("/static", "./configs/microservice/bff-service/static")
    // ...
}

// internal/bff-service/server/http/handler/init.go
v1.Register(r.Group("/v1"))
```

So the full path is: `/v1/static/*` → `./configs/microservice/bff-service/static/*`

## Solution

Modified **Dockerfile.backend** to copy the entire `configs` directory:

```dockerfile
# Before:
COPY configs/microservice ./configs/microservice

# After:
# Copy entire configs directory to ensure static files are included
COPY configs ./configs
```

This ensures:
- All static files are baked into the Docker image
- The application works even if volume mounts fail or are missing
- No dependency on host filesystem structure in Dokploy

## Verification of Files

Files confirmed to exist locally:
```bash
$ ls -la configs/microservice/bff-service/static/icon/
-rwxr-xr-x  user-default-icon.png  (5032 bytes)
-rwxr-xr-x  workflow-default-icon.png  (1422 bytes)
# ... other icons

$ ls -la configs/microservice/bff-service/static/logo/
-rw-r--r--  title_logo.png  (386431 bytes)
-rw-r--r--  login_logo.png  (68197 bytes)
# ... other logos
```

## Deployment Steps

### 1. Rebuild Backend Image
```bash
./rebuild-and-push.sh backend
```

Or manually:
```bash
docker build --platform linux/amd64 \
  -f Dockerfile.backend \
  -t safvr/wanwu-backend:latest \
  -t safvr/wanwu-backend:v0.2.7 \
  --build-arg WANWU_ARCH=amd64 \
  --build-arg WANWU_VERSION=v0.2.7 \
  .

docker push safvr/wanwu-backend:latest
docker push safvr/wanwu-backend:v0.2.7
```

### 2. Deploy to Dokploy

#### Option A: Automatic via GitHub Actions
```bash
git add Dockerfile.backend STATIC_ASSETS_404_FIX.md
git commit -m "Fix: Include all configs in backend image for static asset serving"
git push origin main
```

GitHub Actions will:
1. Build the new backend image
2. Push to Docker Hub
3. Trigger Dokploy deployment

#### Option B: Manual Dokploy Deployment
1. Log into Dokploy dashboard
2. Navigate to your compose deployment
3. Click "Redeploy" or "Rebuild"
4. Dokploy will pull the new `safvr/wanwu-backend:latest` image

### 3. Verify the Fix

#### Check if container has the files:
```bash
# SSH into Dokploy server
ssh ubuntu@your-server

# Check inside the container
docker exec bff-service ls -la /app/configs/microservice/bff-service/static/icon/
docker exec bff-service ls -la /app/configs/microservice/bff-service/static/logo/
```

Should show all files including:
- `user-default-icon.png`
- `title_logo.png`
- etc.

#### Test in browser:
1. Clear browser cache (Ctrl+Shift+R)
2. Navigate to: https://app.safvr.com
3. Check browser console - no 404 errors for static assets
4. Images should load properly

#### Test API directly:
```bash
curl -I https://app.safvr.com/user/api/v1/static/icon/user-default-icon.png
# Should return: HTTP/2 200
```

## Alternative Solutions Considered

### Option 1: Remove volume mount (rejected)
```yaml
# Remove this line from docker-compose.yaml:
- ./configs:/app/configs:ro
```
**Issue**: Breaks local development and flexibility to update configs without rebuilding

### Option 2: Create configs directory on Dokploy server (rejected)
**Issue**: Adds manual deployment steps, error-prone, not automated

### Option 3: Use environment variable to specify configs path (rejected)
**Issue**: Requires code changes, complicates configuration

**Chosen solution** (Option 4): Copy entire configs directory into image - provides best balance of:
- No code changes needed
- Works in all environments (local, Dokploy, K8s)
- Volume mount can still override for local development
- Self-contained Docker image

## Files Changed

- `Dockerfile.backend` - Changed COPY instruction to include all configs
- `STATIC_ASSETS_404_FIX.md` - This documentation

## Testing Checklist

- [ ] Backend image rebuilt and pushed
- [ ] Deployed to Dokploy
- [ ] Verified files exist in container: `/app/configs/microservice/bff-service/static/`
- [ ] Browser test: No 404 errors for static assets
- [ ] Images load correctly on: https://app.safvr.com
- [ ] User icon displays properly
- [ ] Logo displays correctly

## Related Files

- `internal/bff-service/server/http/handler/router/v1/guest.go` - Static file route configuration
- `internal/bff-service/server/http/handler/init.go` - Router group setup
- `configs/middleware/nginx/conf.d/aibase.conf` - Nginx proxy configuration
- `docker-compose.yaml` - Volume mount configuration

## Nginx Configuration (for reference)

```nginx
# configs/middleware/nginx/conf.d/aibase.conf
location ^~ /user/api/ {
    proxy_pass       http://bff-service:6668/;
    proxy_buffering  off;
    proxy_cache      off;
}
```

This configuration:
- Matches requests starting with `/user/api/`
- Strips `/user/api/` prefix and forwards the rest to `bff-service:6668`
- Example: `/user/api/v1/static/icon/user.png` → `http://bff-service:6668/v1/static/icon/user.png`

## Troubleshooting

### Still getting 404 after deployment?

1. **Check if image was actually updated:**
   ```bash
   docker image inspect safvr/wanwu-backend:latest | grep Created
   ```

2. **Verify Dokploy pulled new image:**
   ```bash
   docker exec bff-service ls -la /app/configs/
   # Should show both microservice and middleware directories
   ```

3. **Check bff-service logs:**
   ```bash
   docker logs bff-service | grep -i static
   ```

4. **Test from inside nginx container:**
   ```bash
   docker exec nginx-wanwu curl -I http://bff-service:6668/v1/static/icon/user-default-icon.png
   # Should return: HTTP/1.1 200 OK
   ```

5. **Restart services in order:**
   ```bash
   docker restart bff-service
   docker restart nginx-wanwu
   ```

### Images exist but still 404?

Check the exact URL being requested:
- Look for double slashes: `/user/api//v1/static/...`
- Frontend might be constructing URLs incorrectly
- Check browser Network tab for the exact failing request

### Volume mount still overriding?

In production Dokploy deployment, you may want to remove the volume mount:
```yaml
# Comment out in docker-compose.yaml for production:
# - ./configs:/app/configs:ro
```

But this is optional since the image now includes everything.

## Success Indicators

✅ All static assets load without 404 errors
✅ User default icon displays
✅ Title logo displays
✅ No console errors related to static assets
✅ Backend container has all required files
✅ No need for manual file copying to server

