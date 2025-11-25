# Deployment Summary

## Current Setup

### Docker Images on Docker Hub (`safvr`)
All custom images are built for `linux/amd64` and pushed to `docker.io/safvr`:

| Service | Image | Status |
|---------|-------|--------|
| Agent Base | `safvr/agent-base:latest` | ✅ Built & Pushed |
| Agent | `safvr/agent:latest` | ✅ Built & Pushed |
| Backend | `safvr/wanwu-backend:latest` | ✅ Built & Pushed |
| Frontend | `safvr/wanwu-frontend:latest` | ✅ Built & Pushed |
| RAG Base | `safvr/rag-base:latest` | ✅ Built & Pushed |
| RAG | `safvr/wanwu-rag:latest` | ✅ Built & Pushed |

### Deployment Platform
- **Platform**: Dokploy (on AWS)
- **Deployment Method**: Docker Compose
- **Network**: `wanwu-net` (bridge driver)

## Quick Update Workflow

### ⭐ RECOMMENDED: Automated CI/CD (Most Efficient)

**Setup once, then just push code!**

1. **Make changes locally**
2. **Push to GitHub**:
   ```bash
   git add .
   git commit -m "feat: your changes"
   git push
   ```
3. **Done!** GitHub Actions automatically:
   - Builds changed services
   - Pushes to Docker Hub
   - Dokploy auto-deploys

See `QUICK_START_CICD.md` for setup (5 minutes).

### Alternative: Manual Build (Backup Method)

If you need to build locally:

```bash
# Backend only
./rebuild-and-push.sh backend

# Frontend only
./rebuild-and-push.sh frontend

# All services
./rebuild-and-push.sh all
```

Then redeploy in Dokploy or SSH: `docker compose pull && docker compose up -d`

## Files Created

### CI/CD (Automated - Recommended)
1. **`.github/workflows/build-and-push.yml`** - GitHub Actions CI/CD workflow
2. **`QUICK_START_CICD.md`** - Quick setup guide (5 minutes)
3. **`CI_CD_SETUP.md`** - Detailed CI/CD documentation

### Manual Deployment (Backup)
4. **`DOKPLOY_UPDATE_WORKFLOW.md`** - Manual deployment workflow guide
5. **`rebuild-and-push.sh`** - Local build and push script
6. **`REBUILD_INSTRUCTIONS.md`** - Rebuild and verification steps

### Configuration
7. **`.dockerignore.*`** - Service-specific dockerignore files:
   - `.dockerignore.backend`
   - `.dockerignore.frontend`
   - `.dockerignore.rag`
   - `.dockerignore.agent-base`

## Environment Configuration

### Local `.env` (Development)
Located at: `/Users/mohankumarv/Desktop/SAFVR/wanwu/.env`

### Dokploy `.env` (Production)
Should match local `.env` with these key images:
```bash
WANWU_BACKEND_IMAGE=safvr/wanwu-backend:latest
WANWU_FRONTEND_IMAGE=safvr/wanwu-frontend:latest
WANWU_RAG_IMAGE=safvr/wanwu-rag:latest
WANWU_AGENT_IMAGE=safvr/agent:latest
WANWU_AGENT_BASE_IMAGE=safvr/agent-base:latest
WANWU_KAFKA_IMAGE=bitnami/kafka:3.9
```

## Network Configuration

Updated `docker-compose.yaml`:
```yaml
networks:
  wanwu-net:
    driver: bridge
```

This ensures Docker Compose creates the network automatically.

## Troubleshooting

### Issue: Stale containers on AWS
```bash
# SSH to AWS instance
docker compose down
docker container prune -f
docker network prune -f
docker compose up -d
```

### Issue: Changes not reflecting
```bash
# Force pull and restart
docker compose down
docker image rm safvr/wanwu-backend:latest
docker compose pull
docker compose up -d
```

### Issue: Network not found
Ensure `docker-compose.yaml` has:
```yaml
networks:
  wanwu-net:
    driver: bridge
```

## Current Pending Changes

Based on `REBUILD_INSTRUCTIONS.md`:
- ✅ Permission labels translated to English (Backend)
- ✅ Role names translated to English (Backend)

**To deploy**: Run `./rebuild-and-push.sh backend` then redeploy in Dokploy.

## Best Practices

1. **Always test locally** before pushing to production
2. **Use version tags** for production (e.g., `v1.0.0`) instead of `:latest`
3. **Monitor logs** during deployment
4. **Keep backups** of working images
5. **Document changes** in commit messages

## Next Steps

1. ✅ Build and push backend image with translation changes
2. ⏳ Redeploy in Dokploy
3. ⏳ Verify changes in production
4. ⏳ Test role creation and permissions

## Support Files

- `DOKPLOY_UPDATE_WORKFLOW.md` - Detailed deployment guide
- `REBUILD_INSTRUCTIONS.md` - Rebuild and verification steps
- `rebuild-and-push.sh` - Automated build script
- `ORGANIZATION_HIERARCHY_EXPLAINED.md` - Organization structure docs
- `ORGANIZATION_WORKFLOW.md` - Organization workflow docs

## Commands Reference

```bash
# Build and push backend
./rebuild-and-push.sh backend

# Build and push frontend
./rebuild-and-push.sh frontend

# Build and push all
./rebuild-and-push.sh all

# Update Dokploy (SSH to AWS)
ssh user@aws-instance
cd /path/to/wanwu
docker compose pull
docker compose up -d

# Check logs
docker compose logs -f bff-service
docker compose logs -f nginx

# Check status
docker compose ps
```
