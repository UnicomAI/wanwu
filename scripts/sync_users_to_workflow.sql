-- Script to sync existing users from iam_service.org_users to opencoze.space_user
-- This is a one-time migration script to fix the 403 Forbidden error for existing users

-- Insert all org_users into space_user table
-- role_type: 1=owner, 2=admin, 3=member
-- We'll set all users as members (3) by default, except for the creator of the org who becomes owner (1)

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
        ) THEN 2  -- Admin
        WHEN EXISTS (
            SELECT 1 FROM iam_service.orgs o 
            WHERE o.id = ou.org_id 
            AND o.creator_id = ou.user_id
        ) THEN 1  -- Owner (creator of the org)
        ELSE 3    -- Member
    END AS role_type,
    UNIX_TIMESTAMP() * 1000 AS created_at,
    UNIX_TIMESTAMP() * 1000 AS updated_at
FROM iam_service.org_users ou
WHERE ou.status IS NULL OR ou.status != 'deleted'
ON DUPLICATE KEY UPDATE 
    role_type = VALUES(role_type),
    updated_at = UNIX_TIMESTAMP() * 1000;

-- Verify the sync
SELECT 
    COUNT(*) as total_synced_users,
    COUNT(DISTINCT space_id) as total_spaces,
    SUM(CASE WHEN role_type = 1 THEN 1 ELSE 0 END) as owners,
    SUM(CASE WHEN role_type = 2 THEN 1 ELSE 0 END) as admins,
    SUM(CASE WHEN role_type = 3 THEN 1 ELSE 0 END) as members
FROM opencoze.space_user;

