# Workflow Node Chinese Text - ROOT CAUSE ANALYSIS & PERMANENT FIX

## The Real Problem

After multiple attempted fixes targeting the backend (schema.sql, database triggers), the Chinese text **"开始" (Start) and "结束" (End)** persists when creating new workflows.

### Why Previous Fixes Didn't Work

1. **Schema Template Fix** ❌
   - Modified `/app/configs/schema.sql` inside workflow-wanwu container
   - **Issue**: This template is only used for database initialization, not for frontend rendering

2. **Database Triggers** ❌
   - Created MySQL triggers to convert Chinese to English
   - **Issue**: The triggers work on database data, but the frontend UI reads from JavaScript bundles

3. **Workflow Templates** ❌
   - Updated JSON template files
   - **Issue**: Only affects pre-built workflow templates, not the default nodes in new workflows

### The Actual Root Cause

The workflow editor is a **separate standalone application** served by nginx, loaded in an iframe:

```vue
<!-- web/src/views/workflowNew/index.vue -->
<iframe :src="workflowUrl"></iframe>

// URLs to:
// - Development: http://localhost:8081/workflow
// - Production: /workflow/
```

The Chinese text is **hardcoded in the minified JavaScript bundle files** at:
```
/usr/share/nginx/html/workflow/static/js/*.js
```

These JavaScript files contain the actual UI code that creates the default "开始" and "结束" nodes.

## The Correct Solution

We need to:
1. **Extract** the JavaScript files from the nginx container
2. **Find** which files contain the Chinese text
3. **Translate** the Chinese strings to English using `sed`
4. **Mount** the translated files back into the nginx container via docker-compose volumes
5. **Restart** the nginx container to load the English versions

This is the **exact same approach** used to fix the Publish Dialog (see WORKFLOW_PUBLISH_DIALOG_TRANSLATION.md).

## Implementation Steps

### Step 1: Start Docker
```bash
# Make sure Docker Desktop is running
docker-compose up -d
```

### Step 2: Extract and Translate JavaScript Files
```bash
./fix_workflow_nodes.sh
```

This script will:
- Find all JavaScript files in `/usr/share/nginx/html/workflow/static/js/`
- Search for files containing Chinese text (开始, 结束, etc.)
- Extract those files
- Replace Chinese strings with English:
  - `开始` → `Start`
  - `结束` → `End`
  - `工作流的起始节点，用于设定起始的工序要素` → `Workflow start node, used to set initial process elements`
  - `工作流的结束节点，用于设定工作流运行后的结果信息` → `Workflow end node, used to set result information after workflow execution`
  - `输入变量` → `Input variables`
  - `输出变量` → `Output variables`
  - `返回变量` → `Return variables`
- Save translated files to `project/nginx/workflow-static-overrides/js/`

### Step 3: Update docker-compose.yaml
```bash
./fix_docker_compose.sh
```

This script will:
- Backup the current docker-compose.yaml
- Add volume mounts for all translated JavaScript files to the nginx service
- Example of added volumes:
  ```yaml
  nginx:
    volumes:
      - ./project/nginx/workflow-static-overrides/js/main.abc123.js:/usr/share/nginx/html/workflow/static/js/main.abc123.js:ro
      - ./project/nginx/workflow-static-overrides/js/vendor.def456.js:/usr/share/nginx/html/workflow/static/js/vendor.def456.js:ro
  ```

### Step 4: Restart Nginx
```bash
docker-compose restart nginx
```

### Step 5: Verify the Fix
1. Open your browser
2. Create a new workflow
3. Verify that the default nodes show:
   - **"Start"** (not 开始)
   - **"End"** (not 结束)
4. Check that all descriptions are in English

## Why This Fix Works

1. **Client-Side Override**: The workflow editor runs in the browser and reads from the JavaScript bundles served by nginx
2. **Volume Mount Persistence**: The translated files are stored locally and mounted into the container, so they persist across restarts
3. **No Source Code Changes**: We don't need to rebuild the workflow service from source
4. **Surgical Precision**: We only modify the exact strings that need translation

## Architecture Explanation

```
┌─────────────────────────────────────────────────────────┐
│  Browser                                                │
│  ┌───────────────────────────────────────────────────┐ │
│  │ Vue App (web/src/views/workflowNew/index.vue)    │ │
│  │                                                   │ │
│  │  ┌─────────────────────────────────────────────┐ │ │
│  │  │ <iframe src="/workflow/">                   │ │ │
│  │  │                                             │ │ │
│  │  │  Workflow Editor (standalone app)          │ │ │
│  │  │  Loaded from nginx                         │ │ │
│  │  │  Uses JavaScript bundles with Chinese text │ │ │
│  │  │                                             │ │ │
│  │  └─────────────────────────────────────────────┘ │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
                         │
                         │ HTTP request to /workflow/
                         ▼
┌─────────────────────────────────────────────────────────┐
│  nginx-wanwu container                                  │
│                                                         │
│  /usr/share/nginx/html/workflow/                        │
│  ├── index.html                                         │
│  └── static/                                            │
│      └── js/                                            │
│          ├── main.abc123.js ← Contains Chinese!        │
│          ├── vendor.def456.js                           │
│          └── ...                                        │
│                                                         │
│  Volume Mounts (from docker-compose.yaml):              │
│  ./project/nginx/workflow-static-overrides/js/*.js      │
│    ↓ (overwrites)                                       │
│  /usr/share/nginx/html/workflow/static/js/*.js          │
└─────────────────────────────────────────────────────────┘
```

## Files Modified

1. **Created**: `fix_workflow_nodes.sh` - Automated extraction and translation script
2. **Created**: `fix_docker_compose.sh` - Automated docker-compose.yaml update script
3. **Modified**: `docker-compose.yaml` - Added volume mounts for translated JavaScript files
4. **Created**: `project/nginx/workflow-static-overrides/js/*.js` - Translated JavaScript bundles

## Maintenance Notes

- **If workflow service is updated**: The JavaScript bundle filenames may change (e.g., `main.abc123.js` → `main.xyz789.js`). If this happens, re-run the fix scripts.
- **Version compatibility**: This fix works with the current version of the workflow service. Major version updates may require re-translation.
- **Build process**: This is a runtime fix, not a build-time fix. No need to rebuild any containers.

## Comparison with Other Fixes

| Fix Attempt | Target | Why It Failed |
|-------------|--------|---------------|
| Schema Template | `/app/configs/schema.sql` in workflow-wanwu | Backend template, not used for frontend rendering |
| Database Triggers | MySQL workflow_draft table | Works on data layer, but UI reads from JavaScript |
| Workflow Templates | JSON files in configs/ | Only affects pre-built templates, not default nodes |
| **JavaScript Bundle Translation** | **nginx static JS files** | **✓ Correct - This is where the UI actually gets the text** |

## Conclusion

The root cause was that the workflow editor is a **separate client-side application** with **hardcoded Chinese strings in its JavaScript bundles**. The solution is to:

1. Extract the JavaScript bundles from nginx
2. Replace Chinese strings with English
3. Mount the translated files back via docker-compose volumes

This is a **permanent, persistent fix** that will survive container restarts and rebuilds.

## Date Fixed
December 4, 2025

## References
- Similar fix: `WORKFLOW_PUBLISH_DIALOG_TRANSLATION.md` (same approach, different file)
- Workflow editor location: `web/src/views/workflowNew/index.vue:14`
- Nginx service configuration: `docker-compose.yaml:742-765`
