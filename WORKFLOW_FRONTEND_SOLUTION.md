# Workflow Frontend 404 Solution

## 🎯 Root Cause

The workflow service (`workflow-wanwu:8999`) is **backend-only** - it only serves APIs at `/api/` and `/v1/`.

The workflow **frontend** (React SPA) needs to be served as **static files** by nginx at `/workflow/`.

Currently, nginx has:
- `/usr/share/nginx/html/workflow/static/` - Only has one JS file
- **Missing**: `/usr/share/nginx/html/workflow/index.html` and other frontend files

## 📋 The Workflow Frontend Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Browser                              │
│  https://app.safvr.com/aibase/workflow?id=123               │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                      nginx (8081)                            │
│                                                              │
│  /workflow/          → Serve static files (index.html, JS)  │
│  /api/*              → Proxy to workflow-wanwu:8999         │
│  /v1/*               → Proxy to workflow-wanwu:8999         │
└─────────────────────────────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              workflow-wanwu:8999 (Backend Only)             │
│                                                              │
│  /api/workflow_api/*  - Workflow CRUD APIs                  │
│  /v1/workflow/*       - Workflow execution APIs             │
└─────────────────────────────────────────────────────────────┘
```

## ✅ Solution Options

### Option 1: Extract from Original Aliyun Image (Quickest)

The original `wanwulite/frontend:v0.2.7-e3a5f7d9` image likely has the workflow frontend built-in.

**Steps**:
1. Pull the original image
2. Extract `/usr/share/nginx/html/workflow/` directory
3. Copy into your nginx container or add to your frontend build

**Commands** (run on AWS server):
```bash
# Pull original image
docker pull crpi-6pj79y7ddzdpexs8.cn-hangzhou.personal.cr.aliyuncs.com/wanwulite/frontend:v0.2.7-e3a5f7d9

# Create temp container
docker create --name temp-frontend crpi-6pj79y7ddzdpexs8.cn-hangzhou.personal.cr.aliyuncs.com/wanwulite/frontend:v0.2.7-e3a5f7d9

# Extract workflow frontend
docker cp temp-frontend:/usr/share/nginx/html/workflow /tmp/workflow-frontend

# Copy to running nginx
docker cp /tmp/workflow-frontend/. nginx-wanwu:/usr/share/nginx/html/workflow/

# Cleanup
docker rm temp-frontend

# Verify
docker exec nginx-wanwu ls -la /usr/share/nginx/html/workflow/
docker exec nginx-wanwu ls -la /usr/share/nginx/html/workflow/ | grep index.html
```

### Option 2: Build from Coze Studio Source (Proper Long-term Solution)

Build the workflow frontend from the official Coze Studio repository.

**Steps**:
1. Clone coze-studio repository
2. Build the frontend
3. Add to your Dockerfile.frontend

**Commands**:
```bash
# Clone coze-studio
git clone https://github.com/coze-dev/coze-studio.git /tmp/coze-studio
cd /tmp/coze-studio

# Build frontend (requires Node.js 18+)
cd frontend
npm install
npm run build

# The build output will be in frontend/dist/
# Copy this to your wanwu project
cp -r dist /path/to/wanwu/workflow-frontend-dist
```

Then update `Dockerfile.frontend`:
```dockerfile
# Add after line 22
COPY ./workflow-frontend-dist /usr/share/nginx/html/workflow
```

### Option 3: Update nginx Config to Serve from Workflow Service (Not Recommended)

Some workflow services can serve their own frontend. Check if `workflow-wanwu:8999` has frontend files.

**Test**:
```bash
docker exec workflow-wanwu ls -la /app/static/ 2>/dev/null || echo "No static files"
docker exec workflow-wanwu ls -la /app/public/ 2>/dev/null || echo "No public files"
```

If files exist, update nginx config to proxy everything to the workflow service.

## 🚀 Recommended Approach

**For immediate fix**: Use Option 1 (extract from original image)

**For long-term**: Use Option 2 (build from source) and commit to your repository

## 📝 Update Dockerfile.frontend (After Getting Files)

Once you have the workflow frontend files (from Option 1 or 2):

```dockerfile
ARG WANWU_ARCH

# --- Phase 1: Build stage ---
FROM --platform=linux/$WANWU_ARCH node:14 AS builder
ARG WANWU_ARCH
WORKDIR /app
COPY web .

ENV npm_config_registry=https://registry.npmmirror.com
ENV npm_config_unsafe_perm=true

RUN set -euo && npm install
RUN set -euo && npm rebuild node-sass
RUN set -euo && npm run build

# --- Phase 2: Runtime stage ---
FROM --platform=linux/$WANWU_ARCH nginx:1.27
ARG WANWU_ARCH

COPY ./configs/middleware/nginx/conf.d /etc/nginx/conf.d

COPY --from=builder /app/dist /usr/share/nginx/html/aibase

# Add workflow frontend (from extracted files or build)
COPY ./workflow-frontend /usr/share/nginx/html/workflow

CMD ["nginx", "-g", "daemon off;"]
```

## 🔍 Verification

After deploying:

1. **Check files exist**:
   ```bash
   docker exec nginx-wanwu ls -la /usr/share/nginx/html/workflow/
   docker exec nginx-wanwu cat /usr/share/nginx/html/workflow/index.html | head -5
   ```

2. **Test in browser**:
   - Access: https://app.safvr.com/aibase/workflow
   - Should load the workflow editor UI

3. **Check nginx logs**:
   ```bash
   docker logs nginx-wanwu --tail 50
   ```

## 🎓 Why This Happened

The original `wanwulite/frontend` image had:
- Your custom `web/` frontend → `/usr/share/nginx/html/aibase`
- Coze workflow frontend → `/usr/share/nginx/html/workflow`

Your `Dockerfile.frontend` only builds and copies the `web/` directory, missing the workflow frontend.

The workflow service is backend-only and doesn't serve its own frontend (unlike some monolithic apps).

