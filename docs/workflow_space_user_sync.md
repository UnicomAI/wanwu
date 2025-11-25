# Workflow Space User Synchronization

## Problem

Users were experiencing 403 Forbidden errors when accessing workflows in the deployed application. The error occurred because:

1. The IAM service manages users in the `iam_service` database with the `org_users` table
2. The workflow microservice (Coze) uses a separate `opencoze` database with a `space_user` table
3. There was no synchronization between these two tables
4. When users tried to access workflows, the workflow service checked the `space_user` table and returned 403 Forbidden because the user was not found

## Solution

Implemented automatic synchronization from `org_users` to `space_user` table:

### 1. New Workflow Sync Module

Created `internal/iam-service/client/orm/workflow_sync.go` with:
- `InitWorkflowDB()` - Initializes connection to the workflow database (opencoze)
- `SyncUserToWorkflowSpace()` - Syncs a user to the space_user table
- `RemoveUserFromWorkflowSpace()` - Removes a user from the space_user table

### 2. Updated User Creation Flow

Modified the following functions to automatically sync users:

- `CreateUser()` in `internal/iam-service/client/orm/user.go`
  - Syncs user to workflow database after successful creation
  - Determines role type based on assigned roles (admin vs member)

- `RegisterByEmail()` in `internal/iam-service/client/orm/register.go`
  - Syncs new registered users to both TopOrgID and their new personal space
  - Sets user as owner (role_type=1) of their own space

- `AddOrgUser()` in `internal/iam-service/client/orm/org.go`
  - Syncs user when added to an organization
  - Preserves admin/member role type

- `RemoveOrgUser()` in `internal/iam-service/client/orm/org.go`
  - Removes user from workflow space when removed from organization

### 3. IAM Service Initialization

Updated `cmd/iam-service/main.go` to:
- Initialize workflow database connection on startup
- Gracefully handle connection failures (logs warning but doesn't fail startup)

## Role Type Mapping

The `space_user` table uses the following role types:
- `1` = Owner (creator of the space)
- `2` = Admin (has admin role in the organization)
- `3` = Member (regular user)

## Migration for Existing Users

For existing deployments with users already in the system, run the migration script:

```bash
mysql -u root -p < scripts/sync_users_to_workflow.sql
```

This script will:
1. Sync all existing users from `org_users` to `space_user`
2. Automatically determine role types based on:
   - Organization creator → Owner (1)
   - Users with admin roles → Admin (2)
   - All other users → Member (3)
3. Use `ON DUPLICATE KEY UPDATE` to safely handle re-runs

## Deployment Steps

1. **Build and deploy the updated IAM service**
   ```bash
   docker-compose up -d --build iam-service
   ```

2. **Run the migration script** (for existing users)
   ```bash
   docker exec -i mysql-wanwu mysql -uroot -pWanwu123456 < scripts/sync_users_to_workflow.sql
   ```

3. **Verify the sync**
   ```sql
   SELECT COUNT(*) FROM opencoze.space_user;
   ```

## Testing

After deployment:
1. Log in as an existing user
2. Navigate to a workflow: https://app.safvr.com/aibase/workflow?id=WORKFLOW_ID
3. The workflow should load without 403 Forbidden errors

## Error Handling

The synchronization is designed to be non-blocking:
- If the workflow database connection fails, the IAM service will log a warning but continue to operate
- If individual sync operations fail, they are logged but don't fail the main user operation
- This ensures that IAM operations (user creation, org management) continue to work even if workflow sync is temporarily unavailable

## Monitoring

Check IAM service logs for sync status:
```bash
docker logs iam-service | grep -i "workflow\|space_user"
```

Successful sync messages:
```
Successfully connected to workflow database (opencoze)
Successfully synced user X to workflow space Y with role Z
```

Warning messages (non-critical):
```
Workflow database not initialized, skipping space_user sync
Failed to sync user X to workflow space Y: [error details]
```

