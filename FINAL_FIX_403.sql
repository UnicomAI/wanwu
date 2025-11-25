-- ============================================
-- FINAL FIX FOR 403 FORBIDDEN ERROR
-- The workflow service needs users in TWO tables:
-- 1. opencoze.user - User must exist
-- 2. opencoze.space_user - User must be linked to space
-- ============================================

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

-- STEP 2: Sync all org_users to space_user
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
WHERE (ou.status IS NULL OR ou.status != 'deleted')
AND EXISTS (SELECT 1 FROM iam_service.users u WHERE u.id = ou.user_id AND u.status = 1)
ON DUPLICATE KEY UPDATE 
    role_type = VALUES(role_type),
    updated_at = UNIX_TIMESTAMP() * 1000;

-- STEP 3: Verify the fix
SELECT 'Users in opencoze.user:' AS check_name, COUNT(*) AS count FROM opencoze.user;
SELECT 'Users in opencoze.space_user:' AS check_name, COUNT(*) AS count FROM opencoze.space_user;
SELECT 'Users in space 1:' AS check_name, COUNT(*) AS count FROM opencoze.space_user WHERE space_id = 1;

-- STEP 4: Show all users in space 1
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

