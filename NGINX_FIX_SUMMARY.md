# nginx 403 Fix - Summary

## 🎯 Root Cause (Confirmed)

The 403 error is from **nginx**, not the workflow service or database:

- nginx was trying to serve static files from `/usr/share/nginx/html/workflow/`
- These files don't exist in the container
- nginx returns 403 Forbidden
- Request never reaches the workflow service

**The database sync we did earlier was necessary but not the cause of the 403 error.**

## ✅ Changes Made

### 1. Updated nginx Configuration (Already Correct)

File: `configs/middleware/nginx/conf.d/aibase.conf`

```nginx
location ^~ /workflow/ {
    proxy_pass       http://workflow-wanwu:8999/;
    proxy_buffering  off;
    proxy_cache      off;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

This config was already correct locally, but not deployed to AWS.

### 2. Updated docker-compose.yaml

Added volume mount to ensure nginx always uses the latest config:

```yaml
volumes:
  - ${WANWU_PROJECT_DIR}/nginx/log:/var/log/nginx
  # Mount nginx config to use the updated proxy configuration
  - ./configs/middleware/nginx/conf.d/aibase.conf:/etc/nginx/conf.d/aibase.conf:ro
```

### 3. Updated GitHub Actions Workflow

File: `.github/workflows/build-and-push.yml`

Added nginx config path to frontend build triggers:

```yaml
frontend:
  - 'web/**'
  - 'Dockerfile.frontend'
  - 'configs/middleware/nginx/conf.d/**'  # NEW
```

Now when you change nginx config, it will automatically rebuild the frontend image.

## 🚀 How to Deploy

### Recommended: Automated Deployment

Run this script:

```bash
./deploy-nginx-fix.sh
```

This will:
1. Commit the changes
2. Push to GitHub
3. Trigger GitHub Actions to rebuild the frontend image
4. Automatically trigger Dokploy deployment
5. Deploy the new image to AWS

### Alternative: Manual Deployment via Dokploy

Since the volume mount is already configured, you can also just redeploy via Dokploy UI:

1. Log into Dokploy
2. Find your compose deployment
3. Click "Redeploy"
4. nginx will mount the updated config

## 🔍 Verification

After deployment, verify the fix:

### 1. Check nginx config on server

```bash
ssh ubuntu@34.238.142.181
docker exec nginx-wanwu cat /etc/nginx/conf.d/aibase.conf | grep -A 5 "location.*workflow"
```

Should show `proxy_pass http://workflow-wanwu:8999/`

### 2. Test in browser

1. Clear browser cache (Ctrl+Shift+R or Cmd+Shift+R)
2. Access: https://app.safvr.com/aibase/workflow?id=7576662999398612992
3. Should load the workflow editor! 🎉

### 3. Check nginx logs (if still issues)

```bash
docker logs nginx-wanwu --tail 50
```

## 📋 What Happens After Deployment

1. **GitHub Actions** builds new `safvr/wanwu-frontend:latest` image
2. **Dokploy** pulls the new image
3. **nginx container** restarts with:
   - New config baked into the image (from Dockerfile)
   - Volume mount overlaying the config (from docker-compose.yaml)
4. **Requests to `/workflow/`** are now proxied to `workflow-wanwu:8999`
5. **Workflow service** serves both its API and frontend
6. **403 error is gone!** ✅

## 🎓 Lessons Learned

1. **403 errors can come from different sources**:
   - nginx (file not found, permission denied)
   - Application (authentication/authorization)
   - Always check the response body to identify the source

2. **Volume mounts override image contents**:
   - Even if config is baked into the image, volume mounts take precedence
   - This is useful for quick fixes without rebuilding images

3. **CI/CD path filters matter**:
   - If config changes don't trigger rebuilds, they won't be deployed
   - Always include relevant paths in the workflow triggers

## 🔗 Related Files

- `configs/middleware/nginx/conf.d/aibase.conf` - nginx configuration
- `docker-compose.yaml` - Volume mounts and service definitions
- `.github/workflows/build-and-push.yml` - CI/CD pipeline
- `Dockerfile.frontend` - Frontend image build (includes nginx)
- `deploy-nginx-fix.sh` - Automated deployment script
- `DEPLOY_FIX_TO_AWS.md` - Detailed deployment guide

## 🚨 If It Still Doesn't Work

If you still get errors after deployment:

1. **Check if it's a different error** (not 403 from nginx)
2. **Check workflow service logs**: `docker logs workflow-wanwu --tail 100`
3. **Verify workflow service is running**: `docker ps | grep workflow-wanwu`
4. **Test workflow service directly**: `curl -I http://workflow-wanwu:8999/` (from inside nginx container)
5. **Then** we can investigate session authentication issues

But first, let's fix the nginx issue! 🚀

