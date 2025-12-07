# Workflow Publish Dialog Translation Fix

## Problem
The workflow publish dialog displayed Chinese text instead of English:
- "发布类型" (Publish Type)
- "私密发布为工具：仅自己可见" (Private publish as tool: Only visible to yourself)
- "公开发布为工具：组织内可见" (Public publish as tool: Visible within organization)
- "公开发布为工具：所有人可见" (Public publish as tool: Visible to everyone)

## Root Cause
The Chinese text was embedded in the workflow service's JavaScript bundle file (`index~0.1aa44a47.js`) which is served by nginx from `/usr/share/nginx/html/workflow/static/js/`.

## Solution
1. **Downloaded the JavaScript file** containing the Chinese text
2. **Replaced Chinese strings with English** using sed:
   ```bash
   sed 's/发布类型/Publish Type/g' | \
   sed 's/私密发布为工具：仅自己可见/Private publish as tool: Only visible to yourself/g' | \
   sed 's/公开发布为工具：组织内可见/Public publish as tool: Visible within organization/g' | \
   sed 's/公开发布为工具：所有人可见/Public publish as tool: Visible to everyone/g'
   ```
3. **Created a volume mount** in docker-compose.yaml to override the original file with the translated version

## Files Modified
- `docker-compose.yaml` - Added volume mount for translated JavaScript file
- `project/nginx/workflow-static-overrides/js/index~0.1aa44a47.js` - Translated JavaScript file

## Result
✅ The workflow publish dialog now displays all text in English
✅ The fix persists across container restarts via volume mount
✅ No need to rebuild the workflow service from source

## Notes
- This is a client-side fix that only affects the UI text
- The underlying workflow service API and functionality remain unchanged
- If the workflow service is updated to a new version with a different JavaScript bundle filename, this fix will need to be reapplied

