# Workflow Node Default Language Fix

## Problem
The workflow node-based editor shows Chinese text ("开始" and "结束") for the default Start and End nodes when creating a new workflow.

## Root Cause
The workflow service (`workflow-wanwu` container) is a pre-built binary from `github.com/coze-dev/coze-studio/backend` that has hardcoded Chinese text for default node titles.

## Solutions Implemented

### 1. Updated Workflow Templates ✅
All 30 workflow template JSON files have been updated to use English node titles:
- "开始" → "Start"
- "结束" → "End"

**Script**: `scripts/update_workflow_node_titles.py`

**Result**: All workflow templates now have English node titles.

### 2. Database Update Script ✅
Created SQL script to update existing workflows in the database.

**Script**: `scripts/update_workflow_nodes_in_db.sql`

**Usage**:
```bash
docker exec -i mysql-wanwu mysql -u root -p'Wanwu123456' < scripts/update_workflow_nodes_in_db.sql
```

## Remaining Issue

The workflow service still generates new workflows with Chinese node titles because this is hardcoded in the binary. There are three possible solutions:

### Option 1: Modify Workflow Service Source Code (Recommended)
If you have access to the workflow service source code:

1. Find the node creation code in the workflow service
2. Change the default titles from Chinese to English
3. Rebuild the Docker image

### Option 2: Use a Database Trigger
Create a MySQL trigger that automatically replaces Chinese node titles with English when a new workflow is created:

```sql
USE opencoze;

DELIMITER $$

CREATE TRIGGER workflow_draft_english_nodes
BEFORE INSERT ON workflow_draft
FOR EACH ROW
BEGIN
    SET NEW.canvas = REPLACE(
        REPLACE(NEW.canvas, 
            '"title":"开始"', 
            '"title":"Start"'
        ),
        '"description":"工作流的起始节点，用于设定启动工作流需要的信息"',
        '"description":"The starting node of the workflow, used to set the information needed to start the workflow"'
    );
    
    SET NEW.canvas = REPLACE(
        REPLACE(NEW.canvas, 
            '"title":"结束"', 
            '"title":"End"'
        ),
        '"description":"工作流的最终节点，用于返回工作流运行后的结果信息"',
        '"description":"The final node of the workflow, used to return the result information after the workflow runs"'
    );
END$$

DELIMITER ;
```

### Option 3: Frontend Patch
Modify the frontend code to replace Chinese text with English when displaying workflows.

## Recommended Next Steps

1. **Immediate Fix**: Implement the database trigger (Option 2)
2. **Long-term Fix**: Contact the workflow service maintainers or modify the source code (Option 1)

## Files Created

- `scripts/update_workflow_node_titles.py` - Python script to update template files
- `scripts/update_workflow_nodes_in_db.sql` - SQL script to update existing workflows
- `WORKFLOW_NODE_ENGLISH_FIX.md` - This documentation file

## Testing

After implementing the database trigger, create a new workflow and verify that the Start and End nodes display in English.

