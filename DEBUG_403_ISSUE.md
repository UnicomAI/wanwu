# Debug 403 Issue - Step by Step

## Run these commands in your MySQL terminal to diagnose:

### 1. Check if space_user table has data
```sql
SELECT * FROM opencoze.space_user;
```

### 2. Check what user_id you're logged in as
First, we need to find your actual user ID. Run:
```sql
SELECT id, name, nick, email FROM iam_service.users ORDER BY id;
```

### 3. Check org_users table
```sql
SELECT * FROM iam_service.org_users;
```

### 4. Check if your user is in space_user for space_id=1
```sql
SELECT * FROM opencoze.space_user WHERE space_id = 1;
```

### 5. Check the workflow you're trying to access
```sql
SELECT id, space_id, name, creator_id FROM opencoze.workflow_meta WHERE id = 7576662999398612992;
```

## Based on the logs, I see:

From your logs:
- `"space_id":"1"` - The workflow belongs to space_id 1
- `"creator":{"id":"1"` - The creator is user_id 1
- The workflow service is looking for user permissions

## Possible Issues:

### Issue 1: You're not logged in as user_id 1
The workflow was created by user_id 1, but you might be logged in as a different user.

**Solution:** Check which user you're logged in as by looking at your JWT token or checking the browser console.

### Issue 2: The space_user table is empty or missing your user
Even after migration, your specific user might not be in the table.

**Solution:** Manually add your user:
```sql
-- Replace YOUR_USER_ID with your actual user ID
INSERT INTO opencoze.space_user (space_id, user_id, role_type, created_at, updated_at)
VALUES (1, YOUR_USER_ID, 1, UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000)
ON DUPLICATE KEY UPDATE updated_at = UNIX_TIMESTAMP() * 1000;
```

### Issue 3: The workflow service is checking a different table
The workflow service might be checking permissions in a different way.

**Check the user table in opencoze database:**
```sql
SELECT * FROM opencoze.user;
```

## Quick Fix - Add user_id 1 to space_user

Since the logs show the workflow was created by user_id 1, try this:

```sql
INSERT INTO opencoze.space_user (space_id, user_id, role_type, created_at, updated_at)
VALUES (1, 1, 1, UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000)
ON DUPLICATE KEY UPDATE role_type = 1, updated_at = UNIX_TIMESTAMP() * 1000;
```

Then try accessing the workflow again.

## If still not working, check the actual 403 response

Look at the browser console (F12) and check:
1. What's the exact URL that's returning 403?
2. What are the request headers (especially Authorization)?
3. What's the response body?

The 403 might be coming from:
- Nginx (before reaching the workflow service)
- The workflow service itself
- A different authentication layer

