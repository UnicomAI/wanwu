-- Create MySQL trigger to automatically convert Chinese node titles to English
-- This trigger runs before inserting or updating workflow_draft records

USE opencoze;

-- Drop existing triggers if they exist
DROP TRIGGER IF EXISTS workflow_draft_english_nodes_insert;
DROP TRIGGER IF EXISTS workflow_draft_english_nodes_update;

-- Create trigger for INSERT operations
DELIMITER $$

CREATE TRIGGER workflow_draft_english_nodes_insert
BEFORE INSERT ON workflow_draft
FOR EACH ROW
BEGIN
    -- Replace Start node Chinese text with English
    SET NEW.canvas = REPLACE(
        REPLACE(NEW.canvas, 
            '"title":"开始"', 
            '"title":"Start"'
        ),
        '"description":"工作流的起始节点，用于设定启动工作流需要的信息"',
        '"description":"The starting node of the workflow, used to set the information needed to start the workflow"'
    );
    
    -- Replace End node Chinese text with English
    SET NEW.canvas = REPLACE(
        REPLACE(NEW.canvas, 
            '"title":"结束"', 
            '"title":"End"'
        ),
        '"description":"工作流的最终节点，用于返回工作流运行后的结果信息"',
        '"description":"The final node of the workflow, used to return the result information after the workflow runs"'
    );
END$$

-- Create trigger for UPDATE operations
CREATE TRIGGER workflow_draft_english_nodes_update
BEFORE UPDATE ON workflow_draft
FOR EACH ROW
BEGIN
    -- Replace Start node Chinese text with English
    SET NEW.canvas = REPLACE(
        REPLACE(NEW.canvas, 
            '"title":"开始"', 
            '"title":"Start"'
        ),
        '"description":"工作流的起始节点，用于设定启动工作流需要的信息"',
        '"description":"The starting node of the workflow, used to set the information needed to start the workflow"'
    );
    
    -- Replace End node Chinese text with English
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

-- Verify triggers were created
SHOW TRIGGERS FROM opencoze WHERE `Table` = 'workflow_draft';

