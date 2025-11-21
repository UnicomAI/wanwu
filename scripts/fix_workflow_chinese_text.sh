#!/bin/bash

# Script to fix Chinese workflow node text in coze-studio and rebuild the workflow service

set -e

COZE_STUDIO_PATH="../coze-studio"
NODE_META_FILE="$COZE_STUDIO_PATH/backend/domain/workflow/entity/node_meta.go"

echo "=========================================="
echo "Fixing Chinese Workflow Node Text"
echo "=========================================="
echo ""

# Check if the file exists
if [ ! -f "$NODE_META_FILE" ]; then
    echo "Error: File not found: $NODE_META_FILE"
    exit 1
fi

echo "1. Backing up original file..."
cp "$NODE_META_FILE" "$NODE_META_FILE.backup"
echo "   ✓ Backup created: $NODE_META_FILE.backup"
echo ""

echo "2. Replacing Chinese text with English..."
# Replace "开始" with "Start" on line 264
sed -i '' '264s/Name:         "开始",/Name:         "Start",/' "$NODE_META_FILE"
echo "   ✓ Changed line 264: Name: \"开始\" → Name: \"Start\""

# Replace "结束" with "End" on line 280
sed -i '' '280s/Name:         "结束",/Name:         "End",/' "$NODE_META_FILE"
echo "   ✓ Changed line 280: Name: \"结束\" → Name: \"End\""

# Also replace the descriptions
sed -i '' '266s/Desc:         "工作流的起始节点，用于设定启动工作流需要的信息",/Desc:         "The starting node of the workflow, used to set the information needed to initiate the workflow.",/' "$NODE_META_FILE"
echo "   ✓ Changed line 266: Description to English"

sed -i '' '282s/Desc:         "工作流的最终节点，用于返回工作流运行后的结果信息",/Desc:         "The final node of the workflow, used to return the result information after the workflow runs.",/' "$NODE_META_FILE"
echo "   ✓ Changed line 282: Description to English"
echo ""

echo "3. Verifying changes..."
echo "   Start node (line 264):"
sed -n '264p' "$NODE_META_FILE"
echo "   End node (line 280):"
sed -n '280p' "$NODE_META_FILE"
echo ""

echo "=========================================="
echo "✓ Chinese text successfully replaced!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. cd $COZE_STUDIO_PATH/backend"
echo "2. Build the workflow service: make build-workflow"
echo "3. Replace the Docker image with the new binary"
echo ""

