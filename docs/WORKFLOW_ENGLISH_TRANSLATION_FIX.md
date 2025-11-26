# Workflow Application English Translation Fix

## Overview

This document describes how to fix Chinese text appearing in the workflow application when it should be displayed in English. The workflow application consists of an iframe embedded in the main Vue.js application, with its own separate i18n (internationalization) system.

## Problem Description

The following items were appearing in Chinese instead of English:

1. **Start and End node titles** in the workflow canvas:
   - "开始" (should be "Start")
   - "结束" (should be "End")
   - Node descriptions were also in Chinese

2. **Publish dialog radio button options**:
   - "发布类型" (should be "Publish Type")
   - "私密发布为工具：仅自己可见" (should be "Private publish as tool: Only visible to yourself")
   - "公开发布为工具：组织内可见" (should be "Public publish as tool: Visible within organization")
   - "公开发布为工具：所有人可见" (should be "Public publish as tool: Visible to everyone")

## Root Cause

1. **Node titles**: The Start/End node titles and descriptions are stored in the database (`opencoze.workflow_draft` table) in the `canvas` JSON field, not from i18n translations.

2. **Publish dialog**: The i18n translation file (`workflow-frontend/static/js/index~0.1aa44a47.js`) contains both English and Chinese translations. The Chinese translations in the `zh-CN` section were being used.

## Solution

### Fix 1: Start/End Node Titles (Database Update)

The node titles are stored in the database. To update them to English, run the following SQL:

```sql
USE opencoze;

UPDATE workflow_draft 
SET canvas = REPLACE(
    REPLACE(
        REPLACE(
            REPLACE(
                REPLACE(
                    REPLACE(canvas, 
                        '开始', 'Start'),
                    '结束', 'End'),
                '工作流的起始节点，用于设定启动工作流需要的信息', 'The starting node of the workflow, used to set the information needed to start the workflow'),
            '工作流的最终节点，用于返回工作流运行后的结果信息', 'The final node of the workflow, used to return the result information after the workflow runs'),
        '输入', 'Input'),
    '输出', 'Output')
WHERE canvas LIKE '%开始%' OR canvas LIKE '%结束%';
```

**To run this on the server:**

```bash
ssh -i /path/to/wanwu-aws.pem ubuntu@34.238.142.181 "docker exec -i mysql-wanwu mysql -uroot -pWanwu123456" < update_script.sql
```

### Fix 2: Publish Dialog Radio Buttons (JavaScript File Update)

Edit the file `workflow-frontend/static/js/index~0.1aa44a47.js` and replace the Chinese strings:

```bash
# Backup the original file first
cp workflow-frontend/static/js/index~0.1aa44a47.js workflow-frontend/static/js/index~0.1aa44a47.js.bak

# Replace Chinese with English
sed -i '' 's/"workflow_publish_type":"发布类型"/"workflow_publish_type":"Publish Type"/g' workflow-frontend/static/js/index~0.1aa44a47.js
sed -i '' 's/"workflow_publish_type_private":"私密发布为工具：仅自己可见"/"workflow_publish_type_private":"Private publish as tool: Only visible to yourself"/g' workflow-frontend/static/js/index~0.1aa44a47.js
sed -i '' 's/"workflow_publish_type_org":"公开发布为工具：组织内可见"/"workflow_publish_type_org":"Public publish as tool: Visible within organization"/g' workflow-frontend/static/js/index~0.1aa44a47.js
sed -i '' 's/"workflow_publish_type_public":"公开发布为工具：所有人可见"/"workflow_publish_type_public":"Public publish as tool: Visible to everyone"/g' workflow-frontend/static/js/index~0.1aa44a47.js
```

## Deployment Steps

### Step 1: Update the JavaScript file on the server

```bash
# Copy the updated file to the server
scp -i /path/to/wanwu-aws.pem workflow-frontend/static/js/index~0.1aa44a47.js ubuntu@34.238.142.181:/tmp/

# Copy into the nginx container and fix permissions
ssh -i /path/to/wanwu-aws.pem ubuntu@34.238.142.181 '
docker cp /tmp/index~0.1aa44a47.js nginx-wanwu:/usr/share/nginx/html/workflow/static/js/index~0.1aa44a47.js
docker exec nginx-wanwu chmod 644 /usr/share/nginx/html/workflow/static/js/index~0.1aa44a47.js
'
```

### Step 2: Update the database (if needed for new workflows)

```bash
ssh -i /path/to/wanwu-aws.pem ubuntu@34.238.142.181 "docker exec -i mysql-wanwu mysql -uroot -pWanwu123456" <<EOF
USE opencoze;
UPDATE workflow_draft 
SET canvas = REPLACE(
    REPLACE(
        REPLACE(
            REPLACE(
                REPLACE(
                    REPLACE(canvas, 
                        '开始', 'Start'),
                    '结束', 'End'),
                '工作流的起始节点，用于设定启动工作流需要的信息', 'The starting node of the workflow, used to set the information needed to start the workflow'),
            '工作流的最终节点，用于返回工作流运行后的结果信息', 'The final node of the workflow, used to return the result information after the workflow runs'),
        '输入', 'Input'),
    '输出', 'Output')
WHERE canvas LIKE '%开始%' OR canvas LIKE '%结束%';
EOF
```

### Step 3: Clear browser cache

After deployment, users need to do a hard refresh (Ctrl+Shift+R or Cmd+Shift+R) to clear the cached JavaScript files.

## Important Notes

1. **File Permissions**: When copying files to the nginx container, always set permissions to 644:
   ```bash
   docker exec nginx-wanwu chmod 644 /usr/share/nginx/html/workflow/static/js/index~0.1aa44a47.js
   ```
   Failure to do so will result in a 403 Forbidden error.

2. **Container Persistence**: Changes made directly to container files will be lost on container restart. Always update the source files in the repository and redeploy.

3. **Database vs i18n**: 
   - Node titles come from the DATABASE (`workflow_draft.canvas` JSON field)
   - UI labels come from the JavaScript i18n files

## Server Details

- **Server IP**: 34.238.142.181
- **SSH Key**: wanwu-aws.pem
- **Nginx Container**: nginx-wanwu
- **MySQL Container**: mysql-wanwu
- **Database**: opencoze
- **MySQL Credentials**: root / Wanwu123456

## Files Modified

- `workflow-frontend/static/js/index~0.1aa44a47.js` - Contains i18n translations for the workflow iframe

## Verification

To verify the fix is working:

1. Navigate to the workflow page: `https://app.safvr.com/aibase/workflow?id=<workflow_id>`
2. Check that Start and End nodes show English titles
3. Click the "Publish" button and verify the radio options are in English

## Troubleshooting

### 403 Forbidden Error
If you see a 403 error for JavaScript files after updating:
```bash
docker exec nginx-wanwu chmod 644 /usr/share/nginx/html/workflow/static/js/index~0.1aa44a47.js
```

### Changes Not Appearing
1. Clear browser cache (Ctrl+Shift+R)
2. Check if the file was copied correctly:
   ```bash
   docker exec nginx-wanwu grep -o "workflow_publish_type_private.\{0,60\}" /usr/share/nginx/html/workflow/static/js/index~0.1aa44a47.js
   ```

### Database Not Updated
Verify the update worked:
```bash
docker exec mysql-wanwu mysql -uroot -pWanwu123456 -e "SELECT id, SUBSTRING(canvas, 1, 200) FROM opencoze.workflow_draft WHERE id = <workflow_id>;"
```

---

*Document created: November 26, 2025*

