#!/usr/bin/env python3
"""
Update Chinese text in workflow canvas JSON to English.
This script updates the Start and End node titles and descriptions in the database.
"""

import mysql.connector
import json
import sys

# Database connection settings
DB_CONFIG = {
    'host': 'localhost',
    'port': 3307,  # Mapped port from docker-compose
    'user': 'root',
    'password': 'Wanwu123456',
    'database': 'opencoze',
    'charset': 'utf8mb4'
}

def update_canvas_json(canvas_json):
    """Update Chinese text in canvas JSON to English."""
    try:
        # Parse JSON
        canvas = json.loads(canvas_json)
        
        # Update nodes
        if 'nodes' in canvas:
            for node in canvas['nodes']:
                if 'data' in node and 'nodeMeta' in node['data']:
                    node_meta = node['data']['nodeMeta']
                    
                    # Update Start node
                    if node_meta.get('title') == '开始':
                        node_meta['title'] = 'Start'
                        node_meta['description'] = 'The starting node of the workflow, used to set the information needed to initiate the workflow.'
                    
                    # Update End node
                    elif node_meta.get('title') == '结束':
                        node_meta['title'] = 'End'
                        node_meta['description'] = 'The final node of the workflow, used to return the result information after the workflow runs.'
        
        return json.dumps(canvas, ensure_ascii=False, separators=(',', ': '))
    except Exception as e:
        print(f"Error parsing JSON: {e}", file=sys.stderr)
        return None

def main():
    """Main function to update workflow canvas in database."""
    try:
        # Connect to database
        conn = mysql.connector.connect(**DB_CONFIG)
        cursor = conn.cursor()
        
        # Update workflow_draft table
        print("Updating workflow_draft table...")
        cursor.execute("SELECT id, canvas FROM workflow_draft WHERE canvas LIKE '%开始%' OR canvas LIKE '%结束%'")
        rows = cursor.fetchall()
        
        updated_count = 0
        for row_id, canvas_json in rows:
            updated_canvas = update_canvas_json(canvas_json)
            if updated_canvas:
                cursor.execute("UPDATE workflow_draft SET canvas = %s WHERE id = %s", (updated_canvas, row_id))
                updated_count += 1
                print(f"  Updated workflow_draft id={row_id}")
        
        conn.commit()
        print(f"Updated {updated_count} records in workflow_draft")
        
        # Update workflow_snapshot table
        print("\nUpdating workflow_snapshot table...")
        cursor.execute("SELECT id, canvas FROM workflow_snapshot WHERE canvas LIKE '%开始%' OR canvas LIKE '%结束%'")
        rows = cursor.fetchall()
        
        updated_count = 0
        for row_id, canvas_json in rows:
            updated_canvas = update_canvas_json(canvas_json)
            if updated_canvas:
                cursor.execute("UPDATE workflow_snapshot SET canvas = %s WHERE id = %s", (updated_canvas, row_id))
                updated_count += 1
                print(f"  Updated workflow_snapshot id={row_id}")
        
        conn.commit()
        print(f"Updated {updated_count} records in workflow_snapshot")
        
        cursor.close()
        conn.close()
        
        print("\n✅ Successfully updated all workflow canvas JSON!")
        
    except mysql.connector.Error as e:
        print(f"❌ Database error: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"❌ Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()

