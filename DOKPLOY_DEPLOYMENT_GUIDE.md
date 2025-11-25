# Dokploy Deployment Guide - Workflow 403 Fix

## Quick Summary
You need to deploy these changes to fix the 403 Forbidden error on app.safvr.com because your production database has existing users that were created before this fix.

## Step-by-Step Deployment

### Step 1: Push Code to Repository

```bash
# Make sure all changes are committed
git status

# If there are uncommitted changes:
git add .
git commit -m "Fix: Add automatic workflow space_user synchronization for 403 error"

# Push to your repository
git push origin main  # or your branch name
```

### Step 2: Deploy via Dokploy

#### Option A: Automatic Deployment (if configured)
If Dokploy is set up with auto-deploy from your git repository:
1. Go to Dokploy dashboard
2. Navigate to your `iam-service` deployment
3. Wait for automatic deployment to trigger
4. Monitor the build logs

#### Option B: Manual Deployment
If you need to manually trigger deployment:
1. Go to Dokploy dashboard
2. Navigate to your `iam-service` deployment
3. Click "Redeploy" or "Rebuild" button
4. Wait for the build to complete
5. Check deployment logs for success

**Look for this log message:**
```
Successfully connected to workflow database (opencoze)
```

### Step 3: Run Migration Script

You need to sync existing users to the workflow database. Choose one of these methods:

#### Method A: Using Dokploy Terminal
1. In Dokploy, go to your MySQL service
2. Open the terminal/console
3. Run:
```bash
mysql -uroot -pWanwu123456 <<'EOF'
INSERT INTO opencoze.space_user (space_id, user_id, role_type, created_at, updated_at)
SELECT 
    ou.org_id, ou.user_id,
    CASE 
        WHEN EXISTS (SELECT 1 FROM iam_service.user_roles ur WHERE ur.org_id = ou.org_id AND ur.user_id = ou.user_id AND ur.is_admin = 1) THEN 2
        WHEN EXISTS (SELECT 1 FROM iam_service.orgs o WHERE o.id = ou.org_id AND o.creator_id = ou.user_id) THEN 1
        ELSE 3
    END,
    UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000
FROM iam_service.org_users ou
WHERE ou.status IS NULL OR ou.status != 'deleted'
ON DUPLICATE KEY UPDATE role_type = VALUES(role_type), updated_at = UNIX_TIMESTAMP() * 1000;
EOF
```

#### Method B: Using SSH
1. SSH into your Dokploy server:
```bash
ssh user@your-server.com
```

2. Run the migration:
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 <<'EOF'
INSERT INTO opencoze.space_user (space_id, user_id, role_type, created_at, updated_at)
SELECT 
    ou.org_id, ou.user_id,
    CASE 
        WHEN EXISTS (SELECT 1 FROM iam_service.user_roles ur WHERE ur.org_id = ou.org_id AND ur.user_id = ou.user_id AND ur.is_admin = 1) THEN 2
        WHEN EXISTS (SELECT 1 FROM iam_service.orgs o WHERE o.id = ou.org_id AND o.creator_id = ou.user_id) THEN 1
        ELSE 3
    END,
    UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000
FROM iam_service.org_users ou
WHERE ou.status IS NULL OR ou.status != 'deleted'
ON DUPLICATE KEY UPDATE role_type = VALUES(role_type), updated_at = UNIX_TIMESTAMP() * 1000;
EOF
```

#### Method C: Upload and Run SQL File
1. Upload `scripts/sync_users_to_workflow.sql` to your server
2. Run:
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 < /path/to/sync_users_to_workflow.sql
```

### Step 4: Verify the Fix

#### 4.1 Check Migration Results
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
SELECT 
    COUNT(*) as total_users,
    COUNT(DISTINCT space_id) as total_spaces,
    SUM(CASE WHEN role_type = 1 THEN 1 ELSE 0 END) as owners,
    SUM(CASE WHEN role_type = 2 THEN 1 ELSE 0 END) as admins,
    SUM(CASE WHEN role_type = 3 THEN 1 ELSE 0 END) as members
FROM opencoze.space_user;
"
```

Expected output:
```
+-------------+--------------+--------+--------+---------+
| total_users | total_spaces | owners | admins | members |
+-------------+--------------+--------+--------+---------+
|          XX |           XX |     XX |     XX |      XX |
+-------------+--------------+--------+--------+---------+
```

#### 4.2 Test Workflow Access
1. Open browser and go to: https://app.safvr.com
2. Log in with your account
3. Navigate to: https://app.safvr.com/aibase/workflow?id=7576662999398612992
4. **Expected:** Workflow loads without 403 error
5. **If still 403:** Check the troubleshooting section below

#### 4.3 Check IAM Service Logs
```bash
docker logs iam-service | grep -i "workflow\|space_user" | tail -20
```

Look for:
- ✅ `Successfully connected to workflow database (opencoze)`
- ✅ `Successfully synced user X to workflow space Y with role Z`

### Step 5: Test New User Creation

Create a test user to verify auto-sync works:

1. Register a new user or create via admin panel
2. Check if they appear in space_user:
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
SELECT * FROM opencoze.space_user ORDER BY created_at DESC LIMIT 5;
"
```
3. Log in as the new user and try accessing a workflow

## Troubleshooting

### Issue: IAM service won't start after deployment

**Check logs:**
```bash
docker logs iam-service --tail 100
```

**Common causes:**
- Database connection issues
- Configuration errors

**Solution:**
- Verify MySQL is running: `docker ps | grep mysql`
- Check environment variables in Dokploy

### Issue: Migration script fails

**Error: "Table doesn't exist"**
```bash
# Check if tables exist:
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
SHOW TABLES FROM iam_service LIKE 'org_users';
SHOW TABLES FROM opencoze LIKE 'space_user';
"
```

**Error: "Access denied"**
- Verify MySQL password is correct
- Check if you're using the right container name (`mysql-wanwu`)

### Issue: Still getting 403 errors after deployment

1. **Verify user is in space_user table:**
```bash
# Replace USER_ID with your actual user ID
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
SELECT * FROM opencoze.space_user WHERE user_id = USER_ID;
"
```

2. **Check IAM service logs for sync errors:**
```bash
docker logs iam-service | grep -i error | tail -20
```

3. **Manually add your user:**
```bash
# Replace USER_ID and SPACE_ID with actual values
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
INSERT INTO opencoze.space_user (space_id, user_id, role_type, created_at, updated_at)
VALUES (SPACE_ID, USER_ID, 3, UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000)
ON DUPLICATE KEY UPDATE updated_at = UNIX_TIMESTAMP() * 1000;
"
```

## Rollback Plan

If something goes wrong:

1. **Revert code in git:**
```bash
git revert HEAD
git push
```

2. **Redeploy in Dokploy:**
- Trigger new deployment with reverted code

3. **Note:** The `space_user` data is safe to keep even after rollback

## Post-Deployment Checklist

- [ ] Code pushed to repository
- [ ] IAM service redeployed successfully
- [ ] Migration script executed
- [ ] Verified user count in space_user table
- [ ] Tested workflow access with existing user
- [ ] Tested new user creation and auto-sync
- [ ] No errors in IAM service logs
- [ ] 403 errors resolved

## Need Help?

Check these files for more details:
- `WHY_IT_WORKED_LOCALLY.md` - Explanation of the issue
- `WORKFLOW_403_FIX_SUMMARY.md` - Technical summary
- `docs/fix_403_workflow_error.md` - Detailed troubleshooting
- `DEPLOYMENT_CHECKLIST.md` - Complete deployment checklist

