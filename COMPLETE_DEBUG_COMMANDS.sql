-- ============================================
-- COMPLETE DEBUG SCRIPT FOR 403 ERROR
-- Run these commands one by one in MySQL
-- ============================================

-- 1. CHECK SPACE_USER TABLE
SELECT 'Checking space_user table...' AS step;
SELECT * FROM opencoze.space_user;

-- 2. CHECK IAM USERS
SELECT 'Checking IAM users...' AS step;
SELECT id, name, nick, email, is_admin, status FROM iam_service.users ORDER BY id;

-- 3. CHECK ORG_USERS
SELECT 'Checking org_users...' AS step;
SELECT * FROM iam_service.org_users ORDER BY user_id, org_id;

-- 4. CHECK WORKFLOW META (to see which space the workflow belongs to)
SELECT 'Checking workflow metadata...' AS step;
SELECT id, space_id, name, creator_id, created_at FROM opencoze.workflow_meta 
WHERE id = 7576662999398612992;

-- 5. CHECK IF USER 1 IS IN SPACE_USER FOR SPACE 1
SELECT 'Checking if user 1 is in space 1...' AS step;
SELECT * FROM opencoze.space_user WHERE user_id = 1 AND space_id = 1;

-- 6. CHECK OPENCOZE USER TABLE (workflow service might have its own user table)
SELECT 'Checking opencoze.user table...' AS step;
SELECT * FROM opencoze.user;

-- 7. CHECK SPACE TABLE
SELECT 'Checking space table...' AS step;
SELECT * FROM opencoze.space;

-- ============================================
-- FIXES TO TRY
-- ============================================

-- FIX 1: Add user 1 to space_user for space 1 as owner
SELECT 'FIX 1: Adding user 1 to space_user...' AS step;
INSERT INTO opencoze.space_user (space_id, user_id, role_type, created_at, updated_at)
VALUES (1, 1, 1, UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000)
ON DUPLICATE KEY UPDATE 
    role_type = 1, 
    updated_at = UNIX_TIMESTAMP() * 1000;

-- FIX 2: Check if space 1 exists, if not create it
SELECT 'FIX 2: Checking/Creating space 1...' AS step;
INSERT INTO opencoze.space (id, name, created_at, updated_at)
VALUES (1, 'Default Space', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000)
ON DUPLICATE KEY UPDATE updated_at = UNIX_TIMESTAMP() * 1000;

-- FIX 3: Check if user 1 exists in opencoze.user table, if not create it
SELECT 'FIX 3: Checking/Creating user 1 in opencoze.user...' AS step;
INSERT INTO opencoze.user (id, name, created_at, updated_at)
VALUES (1, 'admin', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000)
ON DUPLICATE KEY UPDATE updated_at = UNIX_TIMESTAMP() * 1000;

-- ============================================
-- VERIFICATION
-- ============================================

-- Verify all fixes
SELECT 'VERIFICATION: space_user' AS check_name;
SELECT * FROM opencoze.space_user WHERE space_id = 1;

SELECT 'VERIFICATION: space' AS check_name;
SELECT * FROM opencoze.space WHERE id = 1;

SELECT 'VERIFICATION: user' AS check_name;
SELECT * FROM opencoze.user WHERE id = 1;

SELECT 'VERIFICATION: workflow_meta' AS check_name;
SELECT id, space_id, name, creator_id FROM opencoze.workflow_meta WHERE id = 7576662999398612992;

-- ============================================
-- COMPREHENSIVE FIX: Sync ALL users
-- ============================================

SELECT 'COMPREHENSIVE FIX: Syncing all users from IAM to workflow...' AS step;

-- First, ensure all users exist in opencoze.user
INSERT INTO opencoze.user (id, name, created_at, updated_at)
SELECT 
    u.id,
    u.name,
    UNIX_TIMESTAMP() * 1000,
    UNIX_TIMESTAMP() * 1000
FROM iam_service.users u
WHERE u.status = 1
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    updated_at = UNIX_TIMESTAMP() * 1000;

-- Then, sync all org_users to space_user
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

-- Final verification
SELECT 'FINAL COUNT: space_user entries' AS metric;
SELECT COUNT(*) as total_entries FROM opencoze.space_user;

SELECT 'FINAL COUNT: users in space 1' AS metric;
SELECT COUNT(*) as users_in_space_1 FROM opencoze.space_user WHERE space_id = 1;

SELECT 'FINAL LIST: All users in space 1' AS metric;
SELECT su.*, u.name as user_name 
FROM opencoze.space_user su
LEFT JOIN opencoze.user u ON su.user_id = u.id
WHERE su.space_id = 1;

