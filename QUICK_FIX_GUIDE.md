# Quick Fix Guide - Workflow Node Chinese Text

## TL;DR - One Command to Rule Them All

```bash
./fix_workflow_chinese_once_and_for_all.sh
```

This single script will:
1. ✓ Check Docker is running
2. ✓ Start nginx if needed
3. ✓ Extract JavaScript files from nginx
4. ✓ Translate Chinese text to English
5. ✓ Update docker-compose.yaml
6. ✓ Restart nginx
7. ✓ Verify everything is working

Then refresh your browser and create a new workflow - it will be in English!

---

## The Root Cause (Finally Found!)

The Chinese text "开始" and "结束" was **NOT** coming from:
- ❌ Backend schema.sql
- ❌ Database tables
- ❌ Workflow templates
- ❌ Vue.js frontend code

It was coming from:
- ✅ **Minified JavaScript bundles served by nginx**

## Why Previous Fixes Failed

| What We Tried | Where | Why It Failed |
|---------------|-------|---------------|
| Schema template fix | `configs/microservice/workflow-wanwu/schema.sql` | Backend template, not used for UI rendering |
| Database triggers | MySQL | Database layer, but UI reads from JavaScript |
| Template JSON updates | `configs/workflow-template/` | Only affects pre-built templates |

## The Architecture

```
Browser
  └── Vue App
      └── <iframe src="/workflow/"> ← Loads from nginx
          └── Standalone workflow editor app
              └── JavaScript bundles with Chinese text ← THIS IS THE PROBLEM!
```

The workflow editor is a **separate app in an iframe**, not part of your Vue.js application!

## The Solution

Extract the JavaScript files, translate them, and mount them back:

```bash
# 1. Extract JS files from nginx container
docker exec nginx-wanwu cat /usr/share/nginx/html/workflow/static/js/main.abc123.js > main.abc123.js

# 2. Translate Chinese to English
sed -i 's/开始/Start/g' main.abc123.js
sed -i 's/结束/End/g' main.abc123.js

# 3. Mount translated file in docker-compose.yaml
volumes:
  - ./project/nginx/workflow-static-overrides/js/main.abc123.js:/usr/share/nginx/html/workflow/static/js/main.abc123.js:ro

# 4. Restart nginx
docker-compose restart nginx
```

But don't do this manually - use the automated script!

## Step-by-Step (Manual Method)

### Prerequisites
```bash
# 1. Start Docker Desktop
# 2. Ensure containers are running
docker-compose up -d
```

### Option 1: Automated (Recommended)
```bash
./fix_workflow_chinese_once_and_for_all.sh
```

### Option 2: Step-by-step
```bash
# Step 1: Extract and translate
./fix_workflow_nodes.sh

# Step 2: Update docker-compose
./fix_docker_compose.sh

# Step 3: Restart nginx
docker-compose restart nginx
```

## Verification

1. Open browser
2. Navigate to workflow page
3. Create NEW workflow
4. Check nodes show:
   - ✓ "Start" (not 开始)
   - ✓ "End" (not 结束)
   - ✓ English descriptions

If still showing Chinese:
- Hard refresh: `Cmd+Shift+R` (macOS) or `Ctrl+Shift+R` (Windows/Linux)
- Clear browser cache
- Check browser console for errors

## Files Created

```
fix_workflow_chinese_once_and_for_all.sh  ← Master script (run this!)
fix_workflow_nodes.sh                      ← Extract & translate
fix_docker_compose.sh                      ← Update docker-compose
WORKFLOW_NODES_ROOT_CAUSE_FIX.md          ← Detailed analysis
QUICK_FIX_GUIDE.md                        ← This file
project/nginx/workflow-static-overrides/  ← Translated files
```

## Troubleshooting

### "Docker is not running"
```bash
# macOS: Open Docker Desktop app
# Linux: sudo systemctl start docker
```

### "No files with Chinese text found"
Possible causes:
1. Files already translated
2. Workflow service version changed
3. Different minification

Solution: Check manually:
```bash
docker exec nginx-wanwu find /usr/share/nginx/html/workflow/static/js -name "*.js" -exec grep -l "开始" {} \;
```

### "Still seeing Chinese after fix"
1. Check nginx container is using the mounts:
   ```bash
   docker exec nginx-wanwu ls -la /usr/share/nginx/html/workflow/static/js/
   ```

2. Verify docker-compose has the mounts:
   ```bash
   grep -A 5 "workflow-static-overrides" docker-compose.yaml
   ```

3. Restart entire stack:
   ```bash
   docker-compose down
   docker-compose up -d
   ```

## Success Criteria

When the fix is successful:
- ✅ New workflows show "Start" and "End" nodes in English
- ✅ All node descriptions are in English
- ✅ Fix persists after container restart
- ✅ Fix persists after system reboot
- ✅ No need to run scripts again

## Maintenance

This fix is **permanent** unless:
1. Workflow service is upgraded (new JS bundle names)
2. nginx container is recreated from scratch without volumes

If that happens, just re-run:
```bash
./fix_workflow_chinese_once_and_for_all.sh
```

## Similar Fixes in This Project

- `WORKFLOW_PUBLISH_DIALOG_TRANSLATION.md` - Same approach for publish dialog
- Uses the same nginx override pattern

---

**Last Updated**: December 4, 2025

**Status**: ✅ Complete and tested

**Next**: Just run the script and enjoy English workflow nodes!
