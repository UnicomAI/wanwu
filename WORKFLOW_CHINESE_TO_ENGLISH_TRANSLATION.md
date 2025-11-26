# Workflow Frontend Chinese to English Translation

## Problem
After redeploying the frontend, the workflow UI displayed Chinese text in several areas:
- **Start node**: "开始" instead of "Start"
- **End node**: "结束" instead of "End"  
- **Publish dialog**: "发布类型", "私密发布为工具：仅自己可见", etc.
- **UI elements**: "取消", "确定", "变量名", "变量值", etc.

## Root Cause
The Chinese text was embedded in the workflow frontend JavaScript bundle file:
- File: `workflow-frontend/static/js/index~0.1aa44a47.js`
- This file is copied into the Docker image during the frontend build
- The file contains hardcoded Chinese strings for UI labels and messages

## Solution Applied

### Comprehensive Translation of All JavaScript Files
The issue was that Chinese text was embedded in **multiple JavaScript files**, not just the main bundle. Webpack splits code into multiple chunks (main bundle + async chunks), and several of these chunks contained Chinese text.

**Translation Script:**
Created `scripts/translate_all_workflow_js.py` to automatically find and translate all JS files containing Chinese text.

**Translation mappings:**
```python
TRANSLATIONS = [
    ('私密发布为工具：仅自己可见', 'Private publish as tool: Only visible to yourself'),
    ('公开发布为工具：组织内可见', 'Public publish as tool: Visible within organization'),
    ('公开发布为工具：所有人可见', 'Public publish as tool: Visible to everyone'),
    ('工作流的起始节点，用于设定启动工作流需要的信息', 'The starting node of the workflow, used to set the information needed to start the workflow'),
    ('工作流的最终节点，用于返回工作流运行后的结果信息', 'The final node of the workflow, used to return the result information after the workflow runs'),
    ('发布类型', 'Publish Type'),
    ('返回变量', 'Return variables'),
    ('返回文本', 'Return text'),
    ('输出变量', 'Output variable'),
    ('变量名', 'Variable name'),
    ('变量值', 'Variable value'),
    ('测试运行', 'Test run'),
    ('结束', 'End'),
    ('开始', 'Start'),
    ('取消', 'Cancel'),
    ('确定', 'OK'),
]
```

### Results
✅ **Translated 8 JavaScript files** with **40+ occurrences** of Chinese text:
- `5313.ac1bfd72.js`: 17 translations (Start, End, Cancel, OK)
- `7335.90144004.js`: 1 translation (Cancel)
- `index~0.1aa44a47.js`: 2 translations (node descriptions)
- `async/676.da07da6a.js`: 1 translation (Start)
- `async/2161.31ec5170.js`: 2 translations (Start, Cancel)
- `async/3215.3bdf929d.js`: 2 translations (Start, Cancel)
- `async/2346.259b852a.js`: 13 translations (Variable name/value, Start, Cancel, OK)
- `async/7242.39f44ed2.js`: 2 translations (Start)

✅ **All critical node labels translated** (开始 → Start, 结束 → End)
✅ **No Chinese text remaining** in node UI elements

## Files Modified
1. `workflow-frontend/static/js/index~0.1aa44a47.js` - Main bundle (node descriptions)
2. `workflow-frontend/static/js/5313.ac1bfd72.js` - Node labels (Start/End)
3. `workflow-frontend/static/js/7335.90144004.js` - UI buttons
4. `workflow-frontend/static/js/async/*.js` - Multiple async chunks with node UI
5. `scripts/translate_all_workflow_js.py` - Translation script for future use
6. `Dockerfile.frontend` - Updated comment to indicate files are pre-translated

## Deployment
The translated file is now committed to the repository. When the frontend Docker image is built:
1. The `workflow-frontend` directory (with translated files) is copied into the image
2. No runtime translation is needed
3. The English text will be displayed immediately

## GitHub Actions Integration
The changes will be automatically deployed via the existing GitHub Actions workflow:
- File: `.github/workflows/build-and-push.yml`
- When changes to `workflow-frontend/**` are pushed, the frontend image is rebuilt
- The new image with English translations is pushed to Docker Hub
- Dokploy automatically pulls and deploys the updated image

## Verification
To verify the translation worked:
```bash
# Run the translation script to check all files
python3 scripts/translate_all_workflow_js.py

# Check for Chinese node labels (should return 0)
grep -r "开始\|结束" workflow-frontend/static/js/*.js | wc -l
# Should return: 0

# Check for English node labels
grep -r "Start\|End" workflow-frontend/static/js/5313.ac1bfd72.js | head -5
# Should show: Start, End
```

## Important Notes
1. **Browser Cache**: After deployment, users may need to clear browser cache or do a hard refresh (Ctrl+Shift+R / Cmd+Shift+R) to see the changes
2. **Docker Rebuild Required**: The frontend Docker image must be rebuilt to include the translated files
3. **Database Data**: If node labels still show Chinese, check if the workflow canvas JSON in the database contains Chinese text (see `scripts/update_workflow_nodes_in_db.sql`)

## Troubleshooting

If you still see Chinese text after deployment:

### 1. Check if Docker image was rebuilt
The frontend Docker image must be rebuilt to include the translated files:
```bash
# Check if workflow-frontend files are in the image
docker run --rm <frontend-image> ls -la /usr/share/nginx/html/workflow/static/js/ | grep "5313\|index"
```

### 2. Clear browser cache
Users need to do a hard refresh:
- **Chrome/Edge**: Ctrl+Shift+R (Windows) or Cmd+Shift+R (Mac)
- **Firefox**: Ctrl+F5 (Windows) or Cmd+Shift+R (Mac)
- Or clear cache: Settings → Privacy → Clear browsing data

### 3. Check database for Chinese text
The workflow canvas JSON in the database might contain Chinese node titles:
```sql
-- Check for workflows with Chinese node titles
SELECT id, name, canvas 
FROM workflow_draft 
WHERE canvas LIKE '%"title":"开始"%' 
   OR canvas LIKE '%"title":"结束"%'
LIMIT 10;

-- Run the update script if needed
-- See: scripts/update_workflow_nodes_in_db.sql
```

### 4. Verify files are translated
```bash
# Should return 0 (no Chinese node labels)
grep -r "开始\|结束" workflow-frontend/static/js/*.js workflow-frontend/static/js/async/*.js 2>/dev/null | wc -l

# Should show English text
grep -r "Start\|End" workflow-frontend/static/js/5313.ac1bfd72.js | head -3
```

## Notes
- This is a **permanent fix** - the translated files are committed to the repository
- The translation script (`scripts/translate_all_workflow_js.py`) can be run again if new files are added
- Similar approach to the database workflow canvas translation we did previously
- If the workflow frontend source is updated and rebuilt, translations may need to be reapplied
- **Important**: After rebuilding the Docker image, ensure Dokploy pulls the new image

