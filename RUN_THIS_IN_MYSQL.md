# IMMEDIATE FIX FOR 403 ERROR

## The Problem
The workflow service (Coze) needs users in **TWO tables**:
1. `opencoze.user` - The user must exist here
2. `opencoze.space_user` - The user must be linked to the space here

Your migration only synced to `space_user`, but NOT to `user` table!

## Run This in Your MySQL Terminal NOW

Copy and paste this entire block:

```sql
-- STEP 1: Sync all users from iam_service.users to opencoze.user
INSERT INTO opencoze.user (id, name, unique_name, email, created_at, updated_at)
SELECT 
    u.id,
    u.name,
    u.name AS unique_name,
    COALESCE(u.email, CONCAT('user', u.id, '@local.com')) AS email,
    UNIX_TIMESTAMP() * 1000,
    UNIX_TIMESTAMP() * 1000
FROM iam_service.users u
WHERE u.status = 1
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    updated_at = UNIX_TIMESTAMP() * 1000;

-- STEP 2: Verify
SELECT 'Users in opencoze.user:' AS check_name, COUNT(*) AS count FROM opencoze.user;
SELECT 'Users in opencoze.space_user:' AS check_name, COUNT(*) AS count FROM opencoze.space_user;

-- STEP 3: Show all users in space 1
SELECT 
    su.space_id,
    su.user_id,
    u.name AS user_name,
    u.email,
    CASE su.role_type
        WHEN 1 THEN 'Owner'
        WHEN 2 THEN 'Admin'
        WHEN 3 THEN 'Member'
        ELSE 'Unknown'
    END AS role
FROM opencoze.space_user su
LEFT JOIN opencoze.user u ON su.user_id = u.id
WHERE su.space_id = 1
ORDER BY su.role_type, su.user_id;
```

## After Running This

1. **Test immediately**: Go to https://app.safvr.com/aibase/workflow?id=7576662999398612992
2. **Expected result**: Workflow should load without 403 error
3. **If still 403**: Check the browser console (F12) for the exact error

## Why This Happened

The workflow service is a third-party service (Coze from ByteDance) that has its own user management:
- It has its own `user` table to store user information
- It has a `space_user` table to link users to spaces (organizations)
- **BOTH tables must have the user** for access to work

Our previous migration only synced to `space_user`, missing the `user` table!

## For Future Deployments

The updated code in `internal/iam-service/client/orm/workflow_sync.go` now syncs to BOTH tables automatically, so new users will work correctly.

## Files Updated

1. `internal/iam-service/client/orm/workflow_sync.go` - Now syncs to both `user` and `space_user` tables
2. `scripts/sync_users_to_workflow.sql` - Updated migration script
3. `FINAL_FIX_403.sql` - Complete fix script

## Quick Verification Commands

```sql
-- Check if user 1 exists in both tables
SELECT 'User 1 in opencoze.user:' AS check_name, COUNT(*) FROM opencoze.user WHERE id = 1;
SELECT 'User 1 in opencoze.space_user:' AS check_name, COUNT(*) FROM opencoze.space_user WHERE user_id = 1 AND space_id = 1;
```

Both should return count = 1.

