-- ============================================
-- VERIFICATION SCRIPT FOR 403 FIX
-- Run these commands to verify the fix was applied correctly
-- ============================================

-- 1. Check if users exist in opencoze.user table
SELECT 'Step 1: Users in opencoze.user table' AS step;
SELECT id, name, email FROM opencoze.user ORDER BY id LIMIT 10;

-- 2. Check if users exist in opencoze.space_user table
SELECT 'Step 2: Users in opencoze.space_user table' AS step;
SELECT space_id, user_id, role_type FROM opencoze.space_user ORDER BY space_id, user_id LIMIT 10;

-- 3. Check specific user (user_id=1) in both tables
SELECT 'Step 3: Check user_id=1 in opencoze.user' AS step;
SELECT * FROM opencoze.user WHERE id = 1;

SELECT 'Step 4: Check user_id=1 in opencoze.space_user' AS step;
SELECT * FROM opencoze.space_user WHERE user_id = 1;

-- 5. Check if there are any users in IAM that are NOT in workflow
SELECT 'Step 5: Users in IAM but NOT in workflow user table' AS step;
SELECT u.id, u.name, u.email 
FROM iam_service.users u
WHERE u.status = 1
AND NOT EXISTS (SELECT 1 FROM opencoze.user wu WHERE wu.id = u.id)
LIMIT 10;

-- 6. Check if there are any org_users that are NOT in space_user
SELECT 'Step 6: Org users NOT in space_user table' AS step;
SELECT ou.org_id, ou.user_id
FROM iam_service.org_users ou
WHERE (ou.status IS NULL OR ou.status != 'deleted')
AND EXISTS (SELECT 1 FROM iam_service.users u WHERE u.id = ou.user_id AND u.status = 1)
AND NOT EXISTS (SELECT 1 FROM opencoze.space_user su WHERE su.space_id = ou.org_id AND su.user_id = ou.user_id)
LIMIT 10;

-- 7. Summary counts
SELECT 'Step 7: Summary' AS step;
SELECT 
    (SELECT COUNT(*) FROM iam_service.users WHERE status = 1) AS iam_users_count,
    (SELECT COUNT(*) FROM opencoze.user) AS workflow_users_count,
    (SELECT COUNT(*) FROM iam_service.org_users WHERE status IS NULL OR status != 'deleted') AS iam_org_users_count,
    (SELECT COUNT(*) FROM opencoze.space_user) AS workflow_space_users_count;

