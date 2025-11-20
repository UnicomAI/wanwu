-- Update all workflow-related tables to convert Chinese node titles to English
USE opencoze;

-- Update workflow_draft table
UPDATE workflow_draft 
SET canvas = REPLACE(
    REPLACE(canvas, 
        '"title":"开始"', 
        '"title":"Start"'
    ),
    '"description":"工作流的起始节点，用于设定启动工作流需要的信息"',
    '"description":"The starting node of the workflow, used to set the information needed to start the workflow"'
)
WHERE canvas LIKE '%"title":"开始"%';

UPDATE workflow_draft 
SET canvas = REPLACE(
    REPLACE(canvas, 
        '"title":"结束"', 
        '"title":"End"'
    ),
    '"description":"工作流的最终节点，用于返回工作流运行后的结果信息"',
    '"description":"The final node of the workflow, used to return the result information after the workflow runs"'
)
WHERE canvas LIKE '%"title":"结束"%';

-- Check if workflow_snapshot has a canvas column
SELECT 'Updating workflow_snapshot table...' as status;

-- Update workflow_snapshot table (if it has canvas column)
UPDATE workflow_snapshot 
SET canvas = REPLACE(
    REPLACE(canvas, 
        '"title":"开始"', 
        '"title":"Start"'
    ),
    '"description":"工作流的起始节点，用于设定启动工作流需要的信息"',
    '"description":"The starting node of the workflow, used to set the information needed to start the workflow"'
)
WHERE canvas LIKE '%"title":"开始"%';

UPDATE workflow_snapshot 
SET canvas = REPLACE(
    REPLACE(canvas, 
        '"title":"结束"', 
        '"title":"End"'
    ),
    '"description":"工作流的最终节点，用于返回工作流运行后的结果信息"',
    '"description":"The final node of the workflow, used to return the result information after the workflow runs"'
)
WHERE canvas LIKE '%"title":"结束"%';

-- Show results
SELECT 'workflow_draft' as table_name, COUNT(*) as rows_with_start 
FROM workflow_draft 
WHERE canvas LIKE '%"title":"Start"%'
UNION ALL
SELECT 'workflow_draft' as table_name, COUNT(*) as rows_with_end 
FROM workflow_draft 
WHERE canvas LIKE '%"title":"End"%'
UNION ALL
SELECT 'workflow_snapshot' as table_name, COUNT(*) as rows_with_start 
FROM workflow_snapshot 
WHERE canvas LIKE '%"title":"Start"%'
UNION ALL
SELECT 'workflow_snapshot' as table_name, COUNT(*) as rows_with_end 
FROM workflow_snapshot 
WHERE canvas LIKE '%"title":"End"%';

