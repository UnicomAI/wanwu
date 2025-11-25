# Understanding the 403 Forbidden Error

## The Real Problem

The workflow service (Coze from ByteDance) is a **third-party closed-source service** with its own user management system. It requires users to exist in **TWO separate tables** in the `opencoze` database:

1. **`opencoze.user`** - The user record (id, name, email, etc.)
2. **`opencoze.space_user`** - The user's membership in a space (space_id, user_id, role_type)

When you access a workflow at `/aibase/workflow?id=xxx`, the frontend loads an iframe pointing to `/workflow/...` which nginx proxies directly to the workflow service. The workflow service then:

1. Checks if the user exists in `opencoze.user` table
2. Checks if the user has access to the space in `opencoze.space_user` table
3. If EITHER check fails → **403 Forbidden**

## Why It Worked Locally

When you run `docker-compose up` with a fresh database:

1. IAM service starts and creates a default admin user
2. **With our new code**, this user creation triggers automatic sync to BOTH:
   - `opencoze.user` table
   - `opencoze.space_user` table
3. Admin user can access workflows immediately ✅

## Why It Failed in Production

Your production database already had users created **BEFORE** the code changes:

1. Old users exist in `iam_service.users` and `iam_service.org_users`
2. Old users do NOT exist in `opencoze.user` or `opencoze.space_user`
3. When they try to access workflows → **403 Forbidden** ❌

## The Fix

### Part 1: Migration Script (One-Time)

Run this SQL to sync existing users:

```sql
-- Sync users to opencoze.user table
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

-- Sync users to opencoze.space_user table
INSERT INTO opencoze.space_user (space_id, user_id, role_type, created_at, updated_at)
SELECT 
    ou.org_id,
    ou.user_id,
    CASE 
        WHEN EXISTS (SELECT 1 FROM iam_service.user_roles ur WHERE ur.org_id = ou.org_id AND ur.user_id = ou.user_id AND ur.is_admin = 1) THEN 2
        WHEN EXISTS (SELECT 1 FROM iam_service.orgs o WHERE o.id = ou.org_id AND o.creator_id = ou.user_id) THEN 1
        ELSE 3
    END,
    UNIX_TIMESTAMP() * 1000,
    UNIX_TIMESTAMP() * 1000
FROM iam_service.org_users ou
WHERE (ou.status IS NULL OR ou.status != 'deleted')
AND EXISTS (SELECT 1 FROM iam_service.users u WHERE u.id = ou.user_id AND u.status = 1)
ON DUPLICATE KEY UPDATE 
    role_type = VALUES(role_type),
    updated_at = UNIX_TIMESTAMP() * 1000;
```

### Part 2: Code Changes (For Future Users)

The updated code in `internal/iam-service/client/orm/workflow_sync.go` now automatically syncs users to BOTH tables when:

- New user is created
- User registers via email
- User is added to an organization
- User is removed from an organization

## Verification

After running the migration, verify with:

```sql
-- Check user 1 exists in both tables
SELECT * FROM opencoze.user WHERE id = 1;
SELECT * FROM opencoze.space_user WHERE user_id = 1 AND space_id = 1;
```

Both queries should return 1 row.

## Why the 403 Might Still Persist

If you're still seeing 403 after running the migration, it could be:

1. **Session/Cookie Issue**: The workflow service might use session cookies. Try:
   - Clear browser cookies
   - Hard refresh (Ctrl+Shift+R)
   - Try in incognito mode

2. **User Not Synced**: Verify the user exists in both tables (see verification above)

3. **Wrong Space ID**: Make sure the `space_id` in the URL matches the user's `org_id`

4. **Workflow Service Not Restarted**: The workflow service might have cached the old user state. Try restarting it:
   ```bash
   docker restart workflow-wanwu
   ```

## Next Steps

1. ✅ Run the migration script (you already did this)
2. ⏳ Run the verification script (`VERIFY_FIX.sql`)
3. ⏳ Clear browser cookies and try again
4. ⏳ If still failing, restart the workflow service
5. ⏳ Deploy the updated code to Dokploy

## Files Updated

- `internal/iam-service/client/orm/workflow_sync.go` - Added user table sync
- `scripts/sync_users_to_workflow.sql` - Updated migration script
- `RUN_THIS_IN_MYSQL.md` - Quick fix instructions
- `VERIFY_FIX.sql` - Verification script

