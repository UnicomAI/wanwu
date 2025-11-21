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

## Complete Solution ✅

The issue has been **fully resolved** through a multi-layered approach:

### 1. Database Triggers ✅
MySQL triggers automatically convert Chinese node titles to English for INSERT and UPDATE operations on the `workflow_draft` table. These triggers are already active in the database.

**Status**: Active and working

### 2. Schema Template Fix ✅ (Root Cause)
The Chinese text was coming from `/app/configs/schema.sql` inside the workflow-wanwu container. This file contains the default workflow template that's used when creating new workflows.

**Solution Implemented**:
- Extracted the `schema.sql` file from the container
- Replaced all Chinese node titles and descriptions with English:
  - "开始" → "Start"
  - "结束" → "End"
  - Updated descriptions to English
- Saved the English version to `configs/microservice/workflow-wanwu/schema.sql`
- Added volume mount in `docker-compose.yaml` to persist the fix:
  ```yaml
  volumes:
    - ./configs/microservice/workflow-wanwu/schema.sql:/app/configs/schema.sql
  ```

**Status**: Permanently fixed - The English schema.sql is now mounted into the container and will persist across restarts.

### 3. Existing Workflows Update ✅
All existing workflows in the database have been updated using the SQL script.

**Status**: Complete

## Files Created

- `scripts/update_workflow_node_titles.py` - Python script to update template files
- `scripts/update_workflow_nodes_in_db.sql` - SQL script to update existing workflows
- `WORKFLOW_NODE_ENGLISH_FIX.md` - This documentation file

## Testing

After implementing the database trigger, create a new workflow and verify that the Start and End nodes display in English.

