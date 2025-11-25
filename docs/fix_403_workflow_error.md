# Fix 403 Forbidden Error for Workflows

## Quick Fix Guide

This guide will help you fix the 403 Forbidden error when accessing workflows at https://app.safvr.com/aibase/workflow.

## Problem Summary

Users get a 403 Forbidden error when trying to access workflows because:
- The IAM service stores users in `iam_service.org_users` table
- The workflow service checks `opencoze.space_user` table for access control
- These tables were not synchronized

## Solution Steps

### Step 1: Rebuild and Deploy IAM Service

The code changes have been made to automatically sync users. Rebuild the IAM service:

```bash
# Navigate to your deployment directory
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Rebuild and restart the IAM service
docker-compose up -d --build iam-service

# Check logs to verify it started successfully
docker logs -f iam-service
```

Look for this message in the logs:
```
Successfully connected to workflow database (opencoze)
```

### Step 2: Sync Existing Users (One-time Migration)

Run the migration script to sync all existing users:

```bash
# Copy the migration script to the MySQL container
docker cp scripts/sync_users_to_workflow.sql mysql-wanwu:/tmp/

# Execute the migration
docker exec -i mysql-wanwu mysql -uroot -p${WANWU_MYSQL_PASSWORD} < /tmp/sync_users_to_workflow.sql
```

Or run it directly:
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 <<EOF
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

### Step 3: Verify the Fix

1. **Check the sync results:**
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

2. **Test the workflow access:**
   - Log in to https://app.safvr.com
   - Navigate to a workflow: https://app.safvr.com/aibase/workflow?id=7576662999398612992
   - The workflow should now load without 403 errors

### Step 4: Verify Automatic Sync for New Users

Create a new user and verify they can access workflows:

1. Register a new user or create one via the admin panel
2. Log in as the new user
3. Try to access a workflow
4. Should work without any 403 errors

## Troubleshooting

### IAM Service Won't Start

Check logs:
```bash
docker logs iam-service
```

If you see workflow database connection errors, verify:
- MySQL is running: `docker ps | grep mysql`
- Database credentials are correct in the config

### Migration Script Fails

Check if the tables exist:
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
SHOW TABLES FROM iam_service LIKE 'org_users';
SHOW TABLES FROM opencoze LIKE 'space_user';
"
```

### Still Getting 403 Errors

1. Check if the user exists in space_user:
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
SELECT * FROM opencoze.space_user WHERE user_id = YOUR_USER_ID;
"
```

2. Check IAM service logs for sync errors:
```bash
docker logs iam-service | grep -i "space_user\|workflow"
```

3. Manually add the user:
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "
INSERT INTO opencoze.space_user (space_id, user_id, role_type, created_at, updated_at)
VALUES (YOUR_SPACE_ID, YOUR_USER_ID, 3, UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000)
ON DUPLICATE KEY UPDATE updated_at = UNIX_TIMESTAMP() * 1000;
"
```

## What Changed

The following files were modified to implement automatic user synchronization:

1. **internal/iam-service/client/orm/workflow_sync.go** (NEW)
   - Handles connection to workflow database
   - Provides sync functions

2. **internal/iam-service/client/orm/user.go**
   - Auto-syncs users when created

3. **internal/iam-service/client/orm/org.go**
   - Auto-syncs when users are added/removed from orgs

4. **internal/iam-service/client/orm/register.go**
   - Auto-syncs new registered users

5. **cmd/iam-service/main.go**
   - Initializes workflow database connection

## Future Users

All new users created after this fix will be automatically synchronized to the workflow database. No manual intervention needed.

