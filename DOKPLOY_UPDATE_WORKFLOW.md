# Dokploy Deployment Update Workflow

## Overview
Your application is deployed on Dokploy using Docker images from `safvr` Docker Hub. To update the deployment with code changes, follow this workflow.

## Current Setup
- **Docker Registry**: `docker.io/safvr`
- **Images**:
  - `safvr/wanwu-backend:latest` (BFF, IAM, App services)
  - `safvr/wanwu-frontend:latest` (Vue.js frontend)
  - `safvr/wanwu-rag:latest` (RAG service)
  - `safvr/agent:latest` (Agent service)
  - `safvr/agent-base:latest` (Agent base)
  - `safvr/rag-base:latest` (RAG base)

## Update Workflow

### Step 1: Make Code Changes Locally
Edit your code as needed (e.g., translations, bug fixes, new features).

### Step 2: Rebuild Docker Images Locally

#### For Backend Changes (Go services):
```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Backup and use backend-specific dockerignore
cp .dockerignore .dockerignore.backup
cp .dockerignore.backend .dockerignore

# Build backend image for x86/amd64
docker build --platform linux/amd64 \
  --build-arg WANWU_ARCH=amd64 \
  -f Dockerfile.backend \
  -t safvr/wanwu-backend:latest .

# Restore original dockerignore
mv .dockerignore.backup .dockerignore
```

#### For Frontend Changes (Vue.js):
```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Backup and use frontend-specific dockerignore
cp .dockerignore .dockerignore.backup
cp .dockerignore.frontend .dockerignore

# Build frontend image for x86/amd64
docker build --platform linux/amd64 \
  --build-arg WANWU_ARCH=amd64 \
  -f Dockerfile.frontend \
  -t safvr/wanwu-frontend:latest .

# Restore original dockerignore
mv .dockerignore.backup .dockerignore
```

#### For RAG Changes:
```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Backup and use rag-specific dockerignore
cp .dockerignore .dockerignore.backup
cp .dockerignore.rag .dockerignore

# Build rag image for x86/amd64
docker build --platform linux/amd64 \
  -f Dockerfile.rag \
  -t safvr/wanwu-rag:latest .

# Restore original dockerignore
mv .dockerignore.backup .dockerignore
```

#### For Agent Changes:
```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Backup and use agent-specific dockerignore
cp .dockerignore .dockerignore.backup
cp .dockerignore.agent-base .dockerignore

# Build agent image for x86/amd64
docker build --platform linux/amd64 \
  -f Dockerfile.agent \
  -t safvr/agent:latest .

# Restore original dockerignore
mv .dockerignore.backup .dockerignore
```

### Step 3: Push Images to Docker Hub
```bash
# Login to Docker Hub (if not already logged in)
docker login

# Push the updated image(s)
docker push safvr/wanwu-backend:latest
docker push safvr/wanwu-frontend:latest
docker push safvr/wanwu-rag:latest
docker push safvr/agent:latest
```

### Step 4: Update Dokploy Deployment

#### Option A: Via Dokploy UI (Recommended)
1. Login to your Dokploy dashboard
2. Navigate to your WanWu application
3. Go to the **Deployments** or **Services** section
4. For each service that changed:
   - Click **Redeploy** or **Restart**
   - Dokploy will pull the latest `:latest` tag from Docker Hub
   - Wait for the deployment to complete

#### Option B: Via SSH to AWS Instance
```bash
# SSH into your AWS instance
ssh your-user@your-aws-instance

# Navigate to your deployment directory
cd /path/to/wanwu

# Pull latest images
docker compose pull

# Restart services
docker compose up -d

# Or restart specific services
docker compose up -d bff-service iam-service app-service
docker compose up -d nginx
```

#### Option C: Force Pull and Restart
```bash
# On your AWS instance
docker compose down
docker compose pull
docker compose up -d
```

### Step 5: Verify Deployment

#### Check Container Status
```bash
# On AWS instance
docker compose ps

# Check logs
docker compose logs -f bff-service
docker compose logs -f nginx
```

#### Test in Browser
1. Clear browser cache (Ctrl/Cmd + Shift + R)
2. Access your application URL
3. Verify changes are reflected
4. Check browser console for errors

## Quick Reference Commands

### Build All Images Locally
```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Backend
cp .dockerignore.backend .dockerignore
docker build --platform linux/amd64 --build-arg WANWU_ARCH=amd64 -f Dockerfile.backend -t safvr/wanwu-backend:latest .

# Frontend
cp .dockerignore.frontend .dockerignore
docker build --platform linux/amd64 --build-arg WANWU_ARCH=amd64 -f Dockerfile.frontend -t safvr/wanwu-frontend:latest .

# RAG
cp .dockerignore.rag .dockerignore
docker build --platform linux/amd64 -f Dockerfile.rag -t safvr/wanwu-rag:latest .

# Agent
cp .dockerignore.agent-base .dockerignore
docker build --platform linux/amd64 -f Dockerfile.agent -t safvr/agent:latest .

# Restore
mv .dockerignore.backup .dockerignore
```

### Push All Images
```bash
docker push safvr/wanwu-backend:latest
docker push safvr/wanwu-frontend:latest
docker push safvr/wanwu-rag:latest
docker push safvr/agent:latest
```

### Update Dokploy (SSH)
```bash
ssh your-user@your-aws-instance
cd /path/to/wanwu
docker compose pull
docker compose up -d
```

## Troubleshooting

### Issue: Changes not reflecting after redeploy

**Cause**: Docker is using cached images

**Solution**:
```bash
# On AWS instance
docker compose down
docker image rm safvr/wanwu-backend:latest
docker image rm safvr/wanwu-frontend:latest
docker compose pull
docker compose up -d
```

### Issue: Build fails locally

**Cause**: Wrong dockerignore or missing dependencies

**Solution**:
- Ensure you're using the correct `.dockerignore.*` file
- Check Docker build logs for specific errors
- Verify all dependencies are available

### Issue: Network errors in Dokploy

**Cause**: Network configuration mismatch

**Solution**:
- Ensure `docker-compose.yaml` has `driver: bridge` for networks
- Run `docker network prune -f` on AWS instance
- Restart deployment

## Best Practices

1. **Always test locally first** before pushing to production
2. **Use version tags** for production (e.g., `v1.0.0`) instead of `:latest`
3. **Keep a backup** of working images
4. **Monitor logs** during deployment
5. **Test thoroughly** after each update

## Version Tags (Recommended for Production)

Instead of using `:latest`, use semantic versioning:

```bash
# Tag with version
docker tag safvr/wanwu-backend:latest safvr/wanwu-backend:v1.0.0
docker push safvr/wanwu-backend:v1.0.0

# Update .env in Dokploy
WANWU_BACKEND_IMAGE=safvr/wanwu-backend:v1.0.0
```

This allows you to:
- Rollback to previous versions easily
- Track which version is deployed
- Test new versions without affecting production

## Current Changes to Deploy

Based on `REBUILD_INSTRUCTIONS.md`, you have:
- ✅ Permission labels translated to English (Backend)
- ✅ Role names translated to English (Backend)

**To deploy these changes:**
1. Build backend image: `docker build ... -f Dockerfile.backend ...`
2. Push to Docker Hub: `docker push safvr/wanwu-backend:latest`
3. Redeploy in Dokploy or run `docker compose up -d` on AWS

## Notes

- Dokploy automatically pulls `:latest` tag when you redeploy
- Make sure your AWS instance has enough disk space for new images
- Consider setting up CI/CD for automated builds and deployments
- Keep your local `.env` and Dokploy `.env` in sync
