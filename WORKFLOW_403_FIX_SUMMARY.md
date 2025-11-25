# Workflow 403 Forbidden Error - Fix Summary

## Problem
Users were getting 403 Forbidden errors when accessing workflows at:
```
https://app.safvr.com/aibase/workflow?id=7576662999398612992
```

Error in browser console:
```
GET https://app.safvr.com/workflow/?workflow_id=7576662999398612992&space_id=1 403 (Forbidden)
```

## Root Cause
The application uses two separate databases:
1. **IAM Service Database** (`iam_service`) - stores users in `org_users` table
2. **Workflow Service Database** (`opencoze`) - checks `space_user` table for access control

When users tried to access workflows, the workflow microservice checked the `space_user` table and returned 403 because users were not synchronized to this table.

## Solution Implemented

### 1. Automatic User Synchronization
Created a new module that automatically syncs users from `org_users` to `space_user` table whenever:
- A new user is created
- A user is added to an organization
- A user registers via email
- A user is removed from an organization

### 2. Files Created
- `internal/iam-service/client/orm/workflow_sync.go` - Core sync functionality
- `scripts/sync_users_to_workflow.sql` - One-time migration for existing users
- `docs/workflow_space_user_sync.md` - Technical documentation
- `docs/fix_403_workflow_error.md` - Deployment guide

### 3. Files Modified
- `internal/iam-service/client/orm/user.go` - Added sync on user creation
- `internal/iam-service/client/orm/org.go` - Added sync on org user add/remove
- `internal/iam-service/client/orm/register.go` - Added sync on user registration
- `cmd/iam-service/main.go` - Initialize workflow database connection

## Deployment Instructions

### Quick Deploy (3 Steps)

1. **Rebuild IAM Service:**
```bash
docker-compose up -d --build iam-service
```

2. **Sync Existing Users:**
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

3. **Verify:**
```bash
docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 -e "SELECT COUNT(*) FROM opencoze.space_user;"
```

Then test by accessing: https://app.safvr.com/aibase/workflow?id=7576662999398612992

## Technical Details

### Role Type Mapping
- `1` = Owner (creator of the space/org)
- `2` = Admin (has admin role in the organization)
- `3` = Member (regular user)

### Error Handling
- Sync operations are non-blocking - they log errors but don't fail the main operation
- If workflow database is unavailable, IAM service continues to work normally
- Sync can be retried by re-running the migration script

### Database Schema
The `space_user` table in `opencoze` database:
```sql
CREATE TABLE `space_user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `space_id` bigint unsigned NOT NULL DEFAULT 0,
  `user_id` bigint unsigned NOT NULL DEFAULT 0,
  `role_type` int NOT NULL DEFAULT 3,
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `updated_at` bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_space_user` (`space_id`, `user_id`),
  KEY `idx_user_id` (`user_id`)
)
```

## Testing

After deployment, verify:

1. **Existing users can access workflows:**
   - Log in as an existing user
   - Navigate to any workflow
   - Should load without 403 errors

2. **New users are automatically synced:**
   - Create a new user
   - Check they appear in `opencoze.space_user` table
   - Verify they can access workflows

3. **Check logs:**
```bash
docker logs iam-service | grep -i "workflow\|space_user"
```

Expected log messages:
```
Successfully connected to workflow database (opencoze)
Successfully synced user X to workflow space Y with role Z
```

## Rollback Plan

If issues occur, you can rollback:

1. **Revert IAM service:**
```bash
git checkout HEAD~1 -- internal/iam-service/ cmd/iam-service/
docker-compose up -d --build iam-service
```

2. **Keep the synced data:**
The `space_user` table data is safe to keep even if you rollback the code.

## Future Maintenance

- All new users will be automatically synced going forward
- No manual intervention needed for new users
- The migration script is idempotent and can be re-run safely if needed

## Support

For issues or questions, check:
- `docs/workflow_space_user_sync.md` - Technical details
- `docs/fix_403_workflow_error.md` - Troubleshooting guide
- IAM service logs: `docker logs iam-service`

