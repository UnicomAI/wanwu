# Fix 404 Not Found Error

## 🎯 Problem

After fixing the 403 error, you're now getting a 404 Not Found error:

```
GET https://app.safvr.com/workflow/?workflow_id=7576737463243112448&space_id=1 404 (Not Found)
```

## 🔍 Root Cause

The nginx proxy configuration had a **trailing slash** in the `proxy_pass` directive:

```nginx
location ^~ /workflow/ {
    proxy_pass http://workflow-wanwu:8999/;  # ← Trailing slash!
}
```

This causes nginx to **strip** the `/workflow/` prefix when proxying:
- Request: `GET /workflow/?workflow_id=123`
- Proxied to workflow service as: `GET /?workflow_id=123`
- Workflow service expects: `GET /workflow/?workflow_id=123`
- Result: **404 Not Found**

## ✅ Fix Applied

Removed the trailing slash from `proxy_pass`:

```nginx
location ^~ /workflow/ {
    proxy_pass http://workflow-wanwu:8999;  # ← No trailing slash!
}
```

Now the full path is preserved:
- Request: `GET /workflow/?workflow_id=123`
- Proxied to workflow service as: `GET /workflow/?workflow_id=123`
- Result: **200 OK** ✅

## 🚀 Deploy the Fix

### Option 1: Push to GitHub (Automated)

```bash
git add configs/middleware/nginx/conf.d/aibase.conf
git commit -m "Fix 404 error: remove trailing slash from workflow proxy_pass"
git push origin main
```

This will:
1. Trigger GitHub Actions to rebuild the frontend image
2. Automatically trigger Dokploy deployment
3. Deploy the fix to AWS

### Option 2: Quick Fix via Dokploy

Since the volume mount is configured, just redeploy:

1. Log into Dokploy
2. Find your compose deployment
3. Click "Redeploy"

The nginx container will mount the updated config immediately.

### Option 3: Manual SSH (Emergency)

```bash
ssh ubuntu@34.238.142.181
cd /path/to/wanwu
git pull origin main
docker compose restart nginx
```

## 🔍 Verification

After deployment:

1. **Check nginx config**:
   ```bash
   docker exec nginx-wanwu cat /etc/nginx/conf.d/aibase.conf | grep -A 2 "location.*workflow"
   ```
   
   Should show:
   ```nginx
   location ^~ /workflow/ {
       proxy_pass       http://workflow-wanwu:8999;
   ```
   (No trailing slash after 8999)

2. **Test in browser**:
   - Clear cache (Ctrl+Shift+R)
   - Access: https://app.safvr.com/aibase/workflow?id=7576737463243112448
   - Should load the workflow editor! 🎉

## 📋 nginx proxy_pass Trailing Slash Behavior

| Configuration | Request | Proxied As | Notes |
|---------------|---------|------------|-------|
| `proxy_pass http://backend/;` | `/api/test` | `/test` | Strips location prefix |
| `proxy_pass http://backend;` | `/api/test` | `/api/test` | Keeps full path |
| `proxy_pass http://backend/v1/;` | `/api/test` | `/v1/test` | Replaces prefix |
| `proxy_pass http://backend/v1;` | `/api/test` | `/v1/api/test` | Appends to path |

**Rule of thumb**: 
- **With trailing slash** → nginx strips the location prefix
- **Without trailing slash** → nginx keeps the full path

## 🎓 Lessons Learned

1. **403 → 404 progression is common**:
   - First fix: nginx config (403 → 404)
   - Second fix: proxy_pass trailing slash (404 → 200)

2. **Trailing slashes matter in nginx**:
   - Small detail, big impact
   - Always test after changing proxy configurations

3. **Volume mounts are great for quick fixes**:
   - No need to rebuild images
   - Just redeploy and the new config is used

## 🚨 If It Still Doesn't Work

If you still get 404 after this fix:

1. **Check workflow service logs**:
   ```bash
   docker logs workflow-wanwu --tail 100
   ```

2. **Verify workflow service is running**:
   ```bash
   docker ps | grep workflow-wanwu
   ```

3. **Test workflow service directly** (from inside nginx container):
   ```bash
   docker exec nginx-wanwu curl -I http://workflow-wanwu:8999/workflow/
   ```

4. **Check if the workflow ID exists**:
   - The workflow ID in the URL might not exist in the database
   - Try accessing the workflow list first: https://app.safvr.com/aibase/workflow

But this trailing slash fix should resolve the 404! 🚀

