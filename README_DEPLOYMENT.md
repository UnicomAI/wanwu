# WanWu Deployment Guide

## 🎯 Quick Answer: How to Deploy Changes?

### ⭐ RECOMMENDED: Automated (2 minutes)
```bash
git add .
git commit -m "your changes"
git push
# ✅ Done! Everything else is automatic
```

### Alternative: Manual (30 minutes)
```bash
./rebuild-and-push.sh backend
# Then redeploy in Dokploy
```

---

## 📚 Documentation Index

### Getting Started (Pick One)

1. **`QUICK_START_CICD.md`** ⭐ **START HERE**
   - 5-minute setup for automated deployments
   - Most efficient workflow
   - Recommended for everyone

2. **`WORKFLOW_COMPARISON.md`**
   - Visual comparison: Manual vs Automated
   - Time and cost savings
   - Why automated is better

### Detailed Guides

3. **`CI_CD_SETUP.md`**
   - Complete CI/CD documentation
   - Advanced features
   - Troubleshooting

4. **`DOKPLOY_UPDATE_WORKFLOW.md`**
   - Manual deployment workflow
   - Backup method
   - Step-by-step instructions

5. **`REBUILD_INSTRUCTIONS.md`**
   - Local build instructions
   - Verification steps
   - Testing guide

6. **`DEPLOYMENT_SUMMARY.md`**
   - Current setup overview
   - Quick reference
   - Files created

---

## 🚀 Setup (One-Time, 5 Minutes)

### Step 1: Add Docker Hub Secret
1. Go to GitHub repo → Settings → Secrets → Actions
2. Add secret: `DOCKER_PASSWORD` = your Docker Hub password

### Step 2: Push CI/CD Workflow
```bash
git add .github/workflows/build-and-push.yml
git commit -m "ci: add automated deployment"
git push
```

### Step 3: Configure Dokploy
- Enable "Auto Deploy" in your Dokploy project
- Set image pull policy to "Always"

✅ **Done! Now just push code to deploy**

---

## 📖 Common Tasks

### Deploy Backend Changes
```bash
# Edit code
vim internal/bff-service/...

# Push
git add .
git commit -m "fix: update backend"
git push

# ✅ Automatically builds and deploys backend
```

### Deploy Frontend Changes
```bash
# Edit code
vim web/src/...

# Push
git add .
git commit -m "feat: update UI"
git push

# ✅ Automatically builds and deploys frontend
```

### Deploy Multiple Services
```bash
# Edit multiple services
vim internal/...
vim web/...

# Push once
git add .
git commit -m "feat: update backend and frontend"
git push

# ✅ Builds both in parallel and deploys
```

### Rollback Deployment
```bash
# Revert last commit
git revert HEAD
git push

# ✅ Previous version rebuilds and redeploys
```

### Manual Build (Backup)
```bash
# If CI/CD is down or you need local build
./rebuild-and-push.sh backend
./rebuild-and-push.sh frontend
./rebuild-and-push.sh all
```

---

## 🔍 Monitor Deployment

### GitHub Actions
- URL: `https://github.com/YOUR_REPO/actions`
- See build progress
- View logs

### Dokploy Dashboard
- Check deployment status
- View service logs
- Monitor health

### Production
```bash
# SSH to AWS
ssh user@aws-instance

# Check status
docker compose ps

# View logs
docker compose logs -f bff-service
```

---

## 📊 What You Get

| Feature | Value |
|---------|-------|
| **Your time per deployment** | 2 minutes (vs 30-45 min) |
| **Total deployment time** | 8-12 minutes (vs 30-45 min) |
| **Error rate** | Very low (automated) |
| **Deploy from** | Anywhere (even phone) |
| **Version tracking** | Automatic (git sha) |
| **Rollback** | Easy (git revert) |
| **Cost** | Free (GitHub Actions + Docker Hub) |
| **Setup time** | 5 minutes (one-time) |

---

## 🛠️ Files Overview

### CI/CD (Automated)
- `.github/workflows/build-and-push.yml` - GitHub Actions workflow
- `QUICK_START_CICD.md` - Quick setup guide
- `CI_CD_SETUP.md` - Detailed documentation
- `WORKFLOW_COMPARISON.md` - Manual vs Automated comparison

### Manual Deployment (Backup)
- `rebuild-and-push.sh` - Local build script
- `DOKPLOY_UPDATE_WORKFLOW.md` - Manual workflow guide
- `REBUILD_INSTRUCTIONS.md` - Build and verification steps

### Configuration
- `.dockerignore.backend` - Backend build context
- `.dockerignore.frontend` - Frontend build context
- `.dockerignore.rag` - RAG build context
- `.dockerignore.agent-base` - Agent build context
- `docker-compose.yaml` - Service definitions
- `.env` - Environment variables

### Documentation
- `README_DEPLOYMENT.md` - This file
- `DEPLOYMENT_SUMMARY.md` - Current setup summary
- `ORGANIZATION_HIERARCHY_EXPLAINED.md` - Org structure
- `ORGANIZATION_WORKFLOW.md` - Org workflows

---

## 🎓 Learning Path

### Day 1: Get Started (5 minutes)
1. Read `QUICK_START_CICD.md`
2. Add Docker Hub secret to GitHub
3. Push workflow file
4. Make a test change and push

### Day 2: Understand the System (15 minutes)
1. Read `WORKFLOW_COMPARISON.md`
2. Read `CI_CD_SETUP.md`
3. Explore GitHub Actions logs
4. Check Dokploy dashboard

### Day 3: Master Deployment (30 minutes)
1. Read `DOKPLOY_UPDATE_WORKFLOW.md`
2. Practice rollback
3. Try manual build (backup method)
4. Customize workflow for your needs

---

## ❓ FAQ

### Q: Do I need Docker running locally?
**A**: No! With CI/CD, GitHub Actions builds images in the cloud.

### Q: How long does deployment take?
**A**: 8-12 minutes total. Your active time: 2 minutes.

### Q: What if CI/CD fails?
**A**: Use manual build script: `./rebuild-and-push.sh backend`

### Q: Can I deploy from my phone?
**A**: Yes! Use GitHub mobile app to commit and push.

### Q: How do I rollback?
**A**: `git revert HEAD && git push` - Previous version redeploys automatically.

### Q: What does it cost?
**A**: Free! GitHub Actions (2,000 min/month) + Docker Hub (unlimited public images).

### Q: Can I test before deploying?
**A**: Yes! Use feature branches, test, then merge to main.

### Q: What if Dokploy doesn't auto-deploy?
**A**: SSH and run: `docker compose pull && docker compose up -d`

---

## 🆘 Troubleshooting

### Build fails in GitHub Actions
→ Check Actions logs for errors
→ Verify `DOCKER_PASSWORD` secret is set
→ Check Dockerfile syntax

### Dokploy not pulling new images
→ SSH: `docker compose pull && docker compose up -d --force-recreate`
→ Check Dokploy auto-deploy settings
→ Verify image was pushed to Docker Hub

### Changes not reflecting
→ Clear browser cache (Ctrl+Shift+R)
→ Check service logs: `docker compose logs -f service-name`
→ Verify correct image version is running

### Need to build locally
→ Start Docker Desktop
→ Run: `./rebuild-and-push.sh backend`
→ Redeploy in Dokploy

---

## 🎯 Best Practices

1. **Use descriptive commit messages**
   ```bash
   git commit -m "feat: add new feature"
   git commit -m "fix: resolve bug in backend"
   git commit -m "docs: update README"
   ```

2. **Test in feature branches**
   ```bash
   git checkout -b feature/new-feature
   # Make changes and test
   git push origin feature/new-feature
   # Create PR, review, then merge to main
   ```

3. **Monitor deployments**
   - Check GitHub Actions for build status
   - Check Dokploy for deployment status
   - Verify in production

4. **Keep documentation updated**
   - Update README when adding features
   - Document breaking changes
   - Keep .env in sync

5. **Use version tags for production**
   ```bash
   # Tag releases
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

---

## 📞 Support

### Documentation
- Start with `QUICK_START_CICD.md`
- Check `WORKFLOW_COMPARISON.md` for comparisons
- Read `CI_CD_SETUP.md` for details

### Logs
- **GitHub Actions**: `https://github.com/YOUR_REPO/actions`
- **Dokploy**: Check dashboard logs
- **Production**: `docker compose logs -f service-name`

### Manual Fallback
- Use `rebuild-and-push.sh` script
- Follow `DOKPLOY_UPDATE_WORKFLOW.md`
- SSH to AWS if needed

---

## 🎉 Summary

**Automated CI/CD is the most efficient way to deploy:**

✅ **2 minutes** of your time per deployment
✅ **Automatic** builds and deployments  
✅ **Smart** change detection (only builds what changed)
✅ **Fast** builds with caching
✅ **Safe** with version tracking and easy rollback
✅ **Free** with GitHub Actions and Docker Hub
✅ **5 minutes** one-time setup

**Just push code and everything else happens automatically!** 🚀

---

## 🚦 Next Steps

1. ⭐ **Read**: `QUICK_START_CICD.md` (2 min)
2. ⚙️ **Setup**: Add Docker Hub secret (2 min)
3. 🧪 **Test**: Push a change (1 min)
4. 🎊 **Enjoy**: Automated deployments forever!

**Start here**: `QUICK_START_CICD.md`
