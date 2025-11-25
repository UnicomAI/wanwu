# Deploy nginx 403 Fix via Dokploy

## 🎯 Problem Summary

The 403 error is caused by **nginx configuration mismatch**:

- **Local config** (in `configs/middleware/nginx/conf.d/aibase.conf`): ✅ Correctly proxies `/workflow/` to `workflow-wanwu:8999`
- **AWS nginx container**: ❌ Still using OLD config that tries to serve static files from `/usr/share/nginx/html/workflow/`
- **Result**: nginx can't find static files → returns 403 Forbidden

This is **NOT** a database sync issue or session authentication issue. The request never reaches the workflow service because nginx blocks it first.

## 🏗️ Your Deployment Setup

- **CI/CD**: GitHub Actions (`.github/workflows/build-and-push.yml`)
- **Deployment**: Dokploy (automatically triggered after successful builds)
- **Images**: `safvr/wanwu-frontend:latest` (includes nginx + Vue.js app)
- **Config**: nginx config is baked into the frontend image at build time

## 🚀 Deployment Options

### ✅ Option 1: Push to GitHub (Recommended - Automated)

This will trigger the full CI/CD pipeline:

**Step 1: Commit and push the changes**

```bash
git add .
git commit -m "Fix nginx 403 error: proxy /workflow/ to workflow service"
git push origin main
```

**Step 2: Monitor GitHub Actions**

1. Go to: https://github.com/YOUR_USERNAME/wanwu/actions
2. Watch the "Build and Push Docker Images" workflow
3. It will:
   - Detect that `configs/middleware/nginx/conf.d/aibase.conf` changed
   - Rebuild `safvr/wanwu-frontend:latest` with the new config
   - Push to Docker Hub
   - Trigger Dokploy deployment

**Step 3: Wait for Dokploy to redeploy**

- Dokploy will automatically pull the new image and restart the nginx container
- This takes ~2-5 minutes

**Step 4: Test in browser**

1. Clear browser cache (Ctrl+Shift+R or Cmd+Shift+R)
2. Access: https://app.safvr.com/aibase/workflow?id=7576662999398612992
3. Should now work! 🎉

---

### ⚡ Option 2: Quick Fix via Dokploy (Immediate)

If you need an immediate fix without waiting for the full rebuild:

**The volume mount is already configured in `docker-compose.yaml`**, so you just need to redeploy via Dokploy.

**Using Dokploy UI:**
1. Log into Dokploy
2. Find your compose deployment
3. Click "Redeploy" or "Restart"
4. The nginx container will mount the updated config from the repository

**Using Dokploy API (if you have MCP access):**
```bash
curl -X POST "$DOKPLOY_URL" \
  -H "x-api-key: $DOKPLOY_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"composeId": "$DOKPLOY_COMPOSE_ID"}'
```

---

### 🔧 Option 3: Manual SSH (Emergency Only)

Only use this if both above options fail:

```bash
ssh ubuntu@34.238.142.181
cd /path/to/wanwu
git pull origin main
docker compose up -d nginx
```

## 📋 What Changed

### Before (OLD config - causes 403):
```nginx
location ^~ /workflow {
    try_files $uri $uri/ /workflow/index.html;
    root   /usr/share/nginx/html/;
}
```
- nginx tries to serve static files from `/usr/share/nginx/html/workflow/`
- Files don't exist → 403 Forbidden

### After (NEW config - works):
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
- nginx proxies all `/workflow/` requests to the workflow service
- Workflow service serves both its API and frontend on port 8999

## 🔍 Why This Happened

The `docker-compose.yaml` was missing the nginx config volume mount, so the container was using the config baked into the `safvr/wanwu-frontend` image (which was built with the old config).

**Fix applied**: Added this line to `docker-compose.yaml`:
```yaml
volumes:
  - ./configs/middleware/nginx/conf.d/aibase.conf:/etc/nginx/conf.d/aibase.conf:ro
```

Now nginx will always use the latest config from your repository.

## ✅ Verification

After deploying, verify:

1. **nginx config is correct**:
   ```bash
   docker exec nginx-wanwu cat /etc/nginx/conf.d/aibase.conf | grep -A 5 "location.*workflow"
   ```

2. **workflow service is running**:
   ```bash
   docker ps | grep workflow-wanwu
   ```

3. **workflow service is accessible from nginx**:
   ```bash
   docker exec nginx-wanwu curl -I http://workflow-wanwu:8999/
   ```
   Should return `HTTP/1.1 200 OK` or similar (not 403)

4. **Browser test**:
   - Clear cache
   - Access https://app.safvr.com/aibase/workflow?id=7576662999398612992
   - Should load the workflow editor!

## 🚨 If It Still Doesn't Work

If you still get 403 after this fix, check:

1. **Is workflow-wanwu container running?**
   ```bash
   docker ps | grep workflow-wanwu
   ```

2. **Check workflow service logs**:
   ```bash
   docker logs workflow-wanwu --tail 100
   ```

3. **Test workflow service directly**:
   ```bash
   curl -I http://34.238.142.181:8999/
   ```
   (Only works if port 8999 is exposed)

4. **Check if it's now a session authentication issue**:
   - If you get a different error (not 403 from nginx)
   - Then we can investigate session authentication
   - But first, let's fix the nginx issue!

## 📝 Summary

| Issue | Cause | Fix |
|-------|-------|-----|
| 403 Forbidden | nginx using old config (tries to serve static files) | Mount updated config that proxies to workflow service |
| Database sync didn't help | Database sync fixes workflow service auth, not nginx config | Deploy nginx config fix |
| Worked locally | Local docker-compose might have different config or volume mounts | Ensure AWS uses same config |

## 🎯 Next Steps

1. ✅ Deploy this fix to AWS (run `./fix-nginx-403.sh`)
2. ✅ Test in browser
3. ⏳ If still issues, check workflow service logs
4. ⏳ If needed, investigate session authentication (but only after nginx fix is confirmed)

