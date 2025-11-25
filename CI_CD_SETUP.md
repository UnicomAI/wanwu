# CI/CD Setup for Automated Deployment

## Overview

This setup provides **fully automated deployment** to Dokploy:
1. You push code changes to GitHub
2. GitHub Actions automatically builds Docker images
3. Images are pushed to Docker Hub (`safvr`)
4. Dokploy automatically pulls and redeploys updated images

## Architecture

```
Local Changes → Git Push → GitHub Actions → Docker Hub → Dokploy → Production
```

## Setup Steps

### 1. Add Secrets to GitHub

1. Go to your GitHub repository
2. Navigate to **Settings → Secrets and variables → Actions**
3. Click **New repository secret**
4. Add the following secrets:

| Secret Name | Value |
|-------------|-------|
| `DOCKER_PASSWORD` | Your Docker Hub password or access token |
| `DOKPLOY_URL` | `http://34.238.142.181:3000` |
| `DOKPLOY_TOKEN` | Your Dokploy API token |
| `DOKPLOY_COMPOSE_ID` | Your compose project ID (e.g., `jtNpnkZEn-HXrGuR2PKEE`) |

**Note**: After images are pushed, the workflow calls `POST /api/compose.deploy` to trigger Dokploy redeployment.

### 2. Configure Dokploy for Auto-Deploy

In your Dokploy project settings:

1. **Enable Auto-Deploy**:
   - Go to your Dokploy project
   - Enable "Auto Deploy" or "Watch for Image Updates"
   - Set image pull policy to "Always"

2. **Configure Webhook (Optional)**:
   - If Dokploy supports webhooks, add the webhook URL to GitHub Actions
   - This triggers immediate redeployment after image push

### 3. Push to GitHub

```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Add all changes
git add .

# Commit with descriptive message
git commit -m "feat: translate permissions and roles to English"

# Push to trigger CI/CD
git push origin main
```

## How It Works

### Intelligent Change Detection

The workflow automatically detects which services changed:

| Changes in | Builds |
|------------|--------|
| `internal/`, `cmd/`, `pkg/` | Backend only |
| `web/` | Frontend only |
| `rag/` | RAG + RAG-base |
| `agent/` | Agent + Agent-base |

### Build Process

1. **Parallel Builds**: Services build in parallel for speed
2. **Layer Caching**: Uses Docker BuildKit cache for faster builds
3. **Multi-tagging**: Tags images with both `:latest` and `:commit-sha`
4. **Platform**: Builds for `linux/amd64` (your AWS instance)

### Deployment Flow

```
Code Push
    ↓
GitHub Actions (5-15 min)
    ↓
Docker Hub (safvr/*)
    ↓
Dokploy Auto-Pull (1-2 min)
    ↓
Production Updated ✅
```

## Workflow Features

### 1. Smart Triggers

Only builds when relevant files change:
```yaml
paths:
  - 'internal/**'    # Backend code
  - 'web/**'         # Frontend code
  - 'rag/**'         # RAG code
  - 'agent/**'       # Agent code
  - 'Dockerfile.*'   # Dockerfiles
```

### 2. Manual Trigger

You can manually trigger builds from GitHub:
1. Go to **Actions** tab
2. Select **Build and Push Docker Images**
3. Click **Run workflow**
4. Choose branch and run

### 3. Build Caching

Uses Docker layer caching to speed up builds:
- First build: ~10-15 minutes
- Subsequent builds: ~3-5 minutes (with cache)

### 4. Version Tracking

Each build is tagged with:
- `:latest` - Always points to the newest version
- `:commit-sha` - Specific version for rollback

## Efficient Workflow

### For Small Changes (e.g., translations, bug fixes)

```bash
# 1. Make changes locally
vim internal/bff-service/server/http/middleware/init.go

# 2. Test locally (optional)
# docker compose up -d

# 3. Commit and push
git add .
git commit -m "fix: update permission labels"
git push

# ✅ Done! CI/CD handles the rest
```

### For Major Changes (e.g., new features)

```bash
# 1. Create feature branch
git checkout -b feature/new-feature

# 2. Make changes and test locally
# ... make changes ...
docker compose up -d
# ... test ...

# 3. Push to feature branch
git add .
git commit -m "feat: add new feature"
git push origin feature/new-feature

# 4. Create Pull Request
# Review and merge to main

# 5. Merge triggers production deployment
# ✅ Automatically deployed!
```

### For Hotfixes

```bash
# 1. Make quick fix
vim internal/bff-service/...

# 2. Push directly to main
git add .
git commit -m "hotfix: critical bug fix"
git push

# ✅ Deployed in ~10 minutes
```

## Monitoring Deployment

### 1. GitHub Actions

Monitor build progress:
1. Go to **Actions** tab in GitHub
2. Click on the running workflow
3. Watch real-time logs

### 2. Dokploy Dashboard

Monitor deployment:
1. Login to Dokploy
2. Check service status
3. View deployment logs

### 3. Production Verification

After deployment completes:
```bash
# SSH to AWS instance
ssh user@your-aws-instance

# Check running containers
docker ps | grep safvr

# Check logs
docker compose logs -f bff-service
docker compose logs -f nginx
```

## Rollback Strategy

### Option 1: Revert Git Commit

```bash
# Revert to previous commit
git revert HEAD
git push

# CI/CD will rebuild and deploy previous version
```

### Option 2: Use Specific Image Tag

```bash
# SSH to AWS instance
ssh user@your-aws-instance

# Update docker-compose.yaml or .env
WANWU_BACKEND_IMAGE=safvr/wanwu-backend:<previous-commit-sha>

# Restart
docker compose up -d
```

### Option 3: Manual Rollback in Dokploy

1. Go to Dokploy dashboard
2. Navigate to deployment history
3. Click "Rollback" on previous successful deployment

## Optimization Tips

### 1. Use Branch Protection

Protect `main` branch:
- Require pull request reviews
- Require status checks to pass
- Prevent direct pushes to main

### 2. Add Testing Stage

Add tests before building images:
```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run tests
        run: |
          go test ./...
          cd web && npm test
  
  build-backend:
    needs: test  # Only build if tests pass
    # ...
```

### 3. Use Staging Environment

Set up staging branch:
```yaml
on:
  push:
    branches:
      - main      # Production
      - staging   # Staging environment
```

### 4. Notification Setup

Add Slack/Discord notifications:
```yaml
- name: Notify deployment
  if: success()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    text: 'Deployment successful! 🚀'
```

## Cost Optimization

### GitHub Actions

- **Free tier**: 2,000 minutes/month for private repos
- **Estimated usage**: ~10-15 min per deployment
- **~130 deployments/month** on free tier

### Docker Hub

- **Free tier**: Unlimited public repositories
- **Rate limits**: 200 pulls/6 hours (anonymous), unlimited (authenticated)
- **Storage**: Unlimited for public images

## Troubleshooting

### Issue: Build fails in GitHub Actions

**Check**:
1. View workflow logs in GitHub Actions
2. Verify `DOCKER_PASSWORD` secret is set
3. Check Dockerfile syntax

### Issue: Dokploy not pulling new images

**Solution**:
```bash
# SSH to AWS instance
docker compose pull
docker compose up -d --force-recreate
```

### Issue: Build is slow

**Solution**:
- Builds use layer caching
- First build is slow (~10-15 min)
- Subsequent builds are faster (~3-5 min)
- Consider using GitHub Actions cache

### Issue: Wrong architecture

**Verify** workflow uses:
```yaml
platforms: linux/amd64
```

## Current Status

✅ **CI/CD Workflow Created**: `.github/workflows/build-and-push.yml`
⏳ **Pending**: Add `DOCKER_PASSWORD` secret to GitHub
⏳ **Pending**: Configure Dokploy auto-deploy
⏳ **Pending**: Push to GitHub to test

## Next Steps

1. **Add Docker Hub secret to GitHub**
2. **Configure Dokploy for auto-deploy**
3. **Test the workflow**:
   ```bash
   git add .
   git commit -m "test: CI/CD setup"
   git push
   ```
4. **Monitor deployment** in GitHub Actions and Dokploy
5. **Verify** changes in production

## Files Overview

- `.github/workflows/build-and-push.yml` - Main CI/CD workflow
- `DOKPLOY_UPDATE_WORKFLOW.md` - Manual deployment guide (backup)
- `rebuild-and-push.sh` - Local build script (backup)
- `CI_CD_SETUP.md` - This file

## Summary

With this setup:
- ✅ **No manual Docker builds** needed
- ✅ **No manual image pushes** needed
- ✅ **Automatic deployment** to production
- ✅ **Smart change detection** (only builds what changed)
- ✅ **Fast builds** with caching
- ✅ **Version tracking** for rollbacks
- ✅ **Parallel builds** for speed

**Just push your code and everything else is automated!** 🚀
