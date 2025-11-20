-- Update workflow node titles from Chinese to English in the database
-- This script updates the "开始" (Start) and "结束" (End) node titles in existing workflows

USE opencoze;

-- Update Start nodes (开始 -> Start)
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

-- Update End nodes (结束 -> End)
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

-- Show count of updated workflows
SELECT COUNT(*) as updated_workflows
FROM workflow_draft
WHERE canvas LIKE '%"title":"Start"%'
   OR canvas LIKE '%"title":"End"%';

