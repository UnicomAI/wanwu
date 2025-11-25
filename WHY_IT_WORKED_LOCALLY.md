# Why Workflows Worked Locally But Not in Production

## Your Questions Answered

### Q1: Why did workflows work locally without these changes?

**Answer:** When you run the application locally with `docker-compose up`, the IAM service automatically creates a **default admin user** on first startup:

```go
// From internal/iam-service/server/grpc/iam/service.go
func (s *Service) InitData() error {
    // Creates admin user with ID, username "admin", password "Wanwu123456"
    if adminUserID, err = s.cli.GetAdminUser(ctx); err == gorm.ErrRecordNotFound {
        if adminUserID, status = s.cli.CreateUser(ctx, &model.User{
            IsAdmin:  true,
            Status:   true,
            Name:     "admin",
            Nick:     "admin",
            Password: "Wanwu123456",
        }, topOrgID, []uint32{adminRoleID}); err != nil {
            // ...
        }
    }
}
```

**The key difference:**

1. **Local Development (Fresh Database):**
   - You start with an empty database
   - IAM service creates the admin user using `CreateUser()` function
   - **WITH our new changes**, `CreateUser()` now automatically syncs the user to `space_user` table
   - So the admin user gets synced to workflow database automatically
   - You can access workflows without 403 errors

2. **Production (Existing Database):**
   - Your production database already has users created BEFORE these code changes
   - Those users were created with the OLD code that didn't sync to `space_user`
   - The `space_user` table is empty for existing users
   - When you try to access workflows, you get 403 Forbidden

### Q2: Do I need to deploy to Dokploy with these changes?

**Answer:** **YES, you need to deploy these changes to production**, but you have two scenarios:

#### Scenario A: Fresh Deployment (New Database)
If you're deploying to a **new server with a fresh database**:
- ✅ Just deploy the new code
- ✅ The admin user will be auto-created and auto-synced
- ✅ All future users will be auto-synced
- ❌ **NO need to run the migration script**

#### Scenario B: Existing Deployment (Existing Users)
If you're deploying to **your existing production server** (app.safvr.com):
- ✅ Deploy the new code
- ✅ **MUST run the migration script** to sync existing users
- ✅ All future users will be auto-synced

**For your case (app.safvr.com with existing users), you MUST do BOTH:**

1. **Deploy the new code:**
   ```bash
   # In Dokploy, rebuild and redeploy the iam-service
   # This ensures new users will be auto-synced going forward
   ```

2. **Run the migration script:**
   ```bash
   # SSH into your server or use Dokploy terminal
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

## Summary

| Environment | Database State | Need Code Deploy? | Need Migration? | Why? |
|-------------|---------------|-------------------|-----------------|------|
| **Local Dev** | Fresh/Empty | ✅ Yes | ❌ No | Admin user auto-created with new code, auto-synced |
| **Production (New)** | Fresh/Empty | ✅ Yes | ❌ No | Same as local - auto-sync works |
| **Production (Existing)** | Has Users | ✅ Yes | ✅ **YES** | Old users need manual sync |

## What Happens After Deployment

### For Existing Users (After Migration):
- ✅ Can access workflows immediately
- ✅ No more 403 errors
- ✅ Proper role-based access (owner/admin/member)

### For New Users (After Code Deploy):
- ✅ Automatically synced when created
- ✅ Automatically synced when added to organizations
- ✅ Automatically synced when registering via email
- ✅ No manual intervention needed

## Deployment Steps for app.safvr.com

1. **Push code to your repository:**
   ```bash
   git add .
   git commit -m "Fix: Add workflow space_user synchronization"
   git push
   ```

2. **In Dokploy:**
   - Go to your iam-service deployment
   - Click "Rebuild" or trigger a new deployment
   - Wait for the build to complete

3. **Run migration (via Dokploy terminal or SSH):**
   ```bash
   docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 < /path/to/scripts/sync_users_to_workflow.sql
   ```
   
   Or use the inline version from above.

4. **Verify:**
   - Log in to https://app.safvr.com
   - Try accessing a workflow
   - Should work without 403 errors

## Why This Architecture?

The workflow microservice (Coze) is a **third-party service from ByteDance** that:
- Uses its own database (`opencoze`)
- Has its own access control (`space_user` table)
- Cannot directly access the IAM service database

So we need to **synchronize** user data between the two systems. This is a common pattern in microservices architecture called **data synchronization** or **eventual consistency**.

