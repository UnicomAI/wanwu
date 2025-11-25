# Deployment Checklist - Workflow 403 Fix

## Pre-Deployment

- [ ] Review all code changes
- [ ] Verify build succeeds: `go build ./cmd/iam-service/`
- [ ] Backup current deployment (optional but recommended)
- [ ] Note current user count for verification

## Deployment Steps

### Step 1: Build and Deploy IAM Service
```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu
docker-compose up -d --build iam-service
```

**Verification:**
- [ ] IAM service container is running: `docker ps | grep iam-service`
- [ ] Check logs for successful startup: `docker logs iam-service | tail -20`
- [ ] Look for: "Successfully connected to workflow database (opencoze)"

### Step 2: Run Migration Script
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 <<'EOF'
INSERT INTO opencoze.space_user (space_id, user_id, role_type, created_at, updated_at)
SELECT 
    ou.org_id AS space_id,
    ou.user_id AS user_id,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM iam_service.user_roles ur 
            WHERE ur.org_id = ou.org_id 
            AND ur.user_id = ou.user_id 
            AND ur.is_admin = 1
        ) THEN 2
        WHEN EXISTS (
            SELECT 1 FROM iam_service.orgs o 
            WHERE o.id = ou.org_id 
            AND o.creator_id = ou.user_id
        ) THEN 1
        ELSE 3
    END AS role_type,
    UNIX_TIMESTAMP() * 1000 AS created_at,
    UNIX_TIMESTAMP() * 1000 AS updated_at
FROM iam_service.org_users ou
WHERE ou.status IS NULL OR ou.status != 'deleted'
ON DUPLICATE KEY UPDATE 
    role_type = VALUES(role_type),
    updated_at = UNIX_TIMESTAMP() * 1000;
EOF
```

**Verification:**
- [ ] Migration completed without errors
- [ ] Check synced user count:
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

### Step 3: Test Workflow Access

**Test with existing user:**
- [ ] Log in to https://app.safvr.com
- [ ] Navigate to: https://app.safvr.com/aibase/workflow?id=7576662999398612992
- [ ] Workflow loads without 403 error
- [ ] Can interact with workflow editor

**Test with new user (if possible):**
- [ ] Create a new user account
- [ ] Log in as new user
- [ ] Access a workflow
- [ ] Verify no 403 errors

### Step 4: Verify Automatic Sync

**Check that new users are auto-synced:**
- [ ] Create a test user via admin panel or registration
- [ ] Check they appear in space_user table:
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
SELECT * FROM opencoze.space_user WHERE user_id = (SELECT MAX(id) FROM iam_service.users);
"
```
- [ ] Verify the new user can access workflows

## Post-Deployment Verification

### Check Logs
```bash
docker logs iam-service | grep -i "workflow\|space_user" | tail -50
```

**Expected messages:**
- [ ] "Successfully connected to workflow database (opencoze)"
- [ ] "Successfully synced user X to workflow space Y with role Z" (when users are created)

### Database Verification
```bash
# Check total users synced
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
SELECT COUNT(*) as synced_users FROM opencoze.space_user;
"

# Check org_users count for comparison
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
SELECT COUNT(*) as org_users FROM iam_service.org_users;
"
```

- [ ] Numbers match or are close (accounting for deleted users)

### Functional Testing
- [ ] Existing users can access workflows
- [ ] New users can access workflows
- [ ] No 403 errors in browser console
- [ ] Workflow editor loads and functions correctly
- [ ] User permissions are respected (owner/admin/member)

## Troubleshooting

### If IAM service fails to start:
```bash
docker logs iam-service
```
Look for database connection errors. Verify MySQL is running.

### If migration fails:
Check if tables exist:
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
SHOW TABLES FROM iam_service LIKE 'org_users';
SHOW TABLES FROM opencoze LIKE 'space_user';
"
```

### If still getting 403 errors:
1. Check if user exists in space_user table
2. Check IAM service logs for sync errors
3. Manually add user if needed (see docs/fix_403_workflow_error.md)

## Rollback Plan

If critical issues occur:

1. **Stop IAM service:**
```bash
docker-compose stop iam-service
```

2. **Revert code changes:**
```bash
git checkout HEAD~1 -- internal/iam-service/ cmd/iam-service/
```

3. **Rebuild and restart:**
```bash
docker-compose up -d --build iam-service
```

4. **Note:** The space_user data is safe to keep even after rollback

## Success Criteria

- [x] IAM service builds successfully
- [ ] IAM service starts without errors
- [ ] Workflow database connection established
- [ ] Existing users synced to space_user table
- [ ] Users can access workflows without 403 errors
- [ ] New users are automatically synced
- [ ] No errors in application logs

## Documentation

Created documentation files:
- [x] `WORKFLOW_403_FIX_SUMMARY.md` - Overview and quick deploy
- [x] `docs/workflow_space_user_sync.md` - Technical details
- [x] `docs/fix_403_workflow_error.md` - Troubleshooting guide
- [x] `scripts/sync_users_to_workflow.sql` - Migration script
- [x] `DEPLOYMENT_CHECKLIST.md` - This file

## Sign-off

- [ ] Deployment completed successfully
- [ ] All tests passed
- [ ] No critical errors in logs
- [ ] Users confirmed workflow access works

**Deployed by:** _______________
**Date:** _______________
**Time:** _______________

