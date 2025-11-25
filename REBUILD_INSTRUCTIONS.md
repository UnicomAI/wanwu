# Rebuild Instructions

## Changes Made

### 1. Permission Labels Translated to English
**File**: `/internal/bff-service/server/http/middleware/init.go`

All Chinese permission labels have been translated to English:
- Model Management
- Knowledge Base
- MCP Platform
- Resource Library
- Safety Guardrails
- Text Q&A
- Workflow
- Agent
- Application Plaza
- Organization Management
- User
- Organization
- Role
- Platform Configuration
- Statistics
- OAuth Key Management

### 2. Role Names Translated to English
**File**: `/internal/iam-service/client/orm/org.go`

Auto-created role names now use English:
- "Organization Administrator" (for sub-organizations)
- "Super Administrator" (for top-level system organization)

## How to Rebuild and Deploy

### Option 1: For Dokploy Deployment (Production - RECOMMENDED)

If you're deploying to Dokploy on AWS:

#### Quick Method (Using Script):
```bash
# Navigate to project root
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Rebuild and push backend only (for permission/role changes)
./rebuild-and-push.sh backend

# Or rebuild and push all services
./rebuild-and-push.sh all
```

#### Manual Method:
```bash
# 1. Build backend image locally
cd /Users/mohankumarv/Desktop/SAFVR/wanwu
cp .dockerignore.backend .dockerignore
docker build --platform linux/amd64 --build-arg WANWU_ARCH=amd64 \
  -f Dockerfile.backend -t safvr/wanwu-backend:latest .
mv .dockerignore.backup .dockerignore

# 2. Push to Docker Hub
docker push safvr/wanwu-backend:latest

# 3. Update Dokploy deployment
# Option A: Via Dokploy UI - Click "Redeploy" on your service
# Option B: Via SSH to AWS instance:
ssh your-user@your-aws-instance
cd /path/to/wanwu
docker compose pull
docker compose up -d
```

See `DOKPLOY_UPDATE_WORKFLOW.md` for complete details.

### Option 2: Using Docker Compose (Local Development)

If you're using the Docker setup locally:

```bash
# Navigate to project root
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Rebuild and restart the services
docker-compose down
docker-compose build
docker-compose up -d

# Or rebuild specific services
docker-compose build bff-service iam-service
docker-compose up -d bff-service iam-service
```

### Option 2: Manual Build

If you're building manually:

```bash
# Navigate to project root
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Build BFF service
cd internal/bff-service
go build -o ../../bin/bff-service ./cmd

# Build IAM service
cd ../iam-service
go build -o ../../bin/iam-service ./cmd

# Restart the services
# (Use your specific process manager or restart commands)
```

### Option 3: Using Make (if available)

Check if there's a Makefile:

```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Common make commands
make build
make restart

# Or
make clean && make build && make deploy
```

## Verification Steps

### 1. Check Backend is Running

```bash
# Check if services are running
docker-compose ps

# Or check processes
ps aux | grep -E "bff-service|iam-service"
```

### 2. Verify Changes in UI

1. **Clear browser cache** (Ctrl/Cmd + Shift + R)
2. Login to the application
3. Navigate to: **Settings → Role**
4. Click **"Add Role"** or **"Edit Role"**
5. Look at the **Menu Permissions** dropdown
6. **All labels should now be in English!**

### 3. Test Organization Creation

1. **As System Admin**: Create a top-level organization
2. **Switch to that organization**: Check the header dropdown
3. **Create another organization**: It should create a sub-organization
4. Verify the new role created is named "Organization Administrator" (in English)

## Troubleshooting

### Issue: Changes not reflecting in UI

**Solution**:
```bash
# Hard refresh the browser
Ctrl/Cmd + Shift + R

# Or clear browser cache completely
# Chrome: Settings → Privacy → Clear browsing data
# Firefox: Settings → Privacy → Clear Data
```

### Issue: Backend not starting

**Solution**:
```bash
# Check logs
docker-compose logs -f bff-service
docker-compose logs -f iam-service

# Look for build errors or runtime errors
```

### Issue: Database migration needed

If you see database-related errors, you might need to update existing roles:

**Note**: This will only affect newly created organizations. Existing organizations will keep their Chinese role names unless manually updated in the database.

## Expected Results

After rebuilding:

✅ All permission labels in role management are in English
✅ Newly created organizations have English role names
✅ Organization hierarchy works as documented
✅ Sub-organizations are created correctly based on current context

## Next Steps

1. Test the role creation with English labels
2. Create a test organization and verify the admin role is in English
3. Test the sub-organization creation workflow
4. Update the documentation if needed

## Notes

- Existing organizations with Chinese role names will remain unchanged
- Only new organizations created after this update will have English role names
- If you need to update existing role names, you'll need to do it manually via the database or through the UI's edit function
