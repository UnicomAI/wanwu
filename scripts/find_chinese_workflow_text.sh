#!/bin/bash

# Script to find Chinese workflow node text in the coze-studio repository
# This will help us locate where "开始" and "结束" are defined

REPO_PATH="${1:-../coze-studio}"

echo "Searching for Chinese workflow node text in: $REPO_PATH"
echo "=============================================="
echo ""

echo "1. Searching for '开始' (Start):"
echo "-----------------------------------"
grep -r "开始" "$REPO_PATH" --include="*.go" --include="*.ts" --include="*.tsx" --include="*.js" --include="*.jsx" --include="*.json" -n | grep -v "node_modules" | grep -v ".git" | head -20

echo ""
echo "2. Searching for '结束' (End):"
echo "-----------------------------------"
grep -r "结束" "$REPO_PATH" --include="*.go" --include="*.ts" --include="*.tsx" --include="*.js" --include="*.jsx" --include="*.json" -n | grep -v "node_modules" | grep -v ".git" | head -20

echo ""
echo "3. Searching for 'nodeMeta' with 'title':"
echo "-----------------------------------"
grep -r "nodeMeta" "$REPO_PATH" --include="*.go" --include="*.ts" --include="*.tsx" -n -A 3 | grep -E "(title|开始|结束)" | head -30

echo ""
echo "4. Searching for workflow node type definitions:"
echo "-----------------------------------"
grep -r "type.*Node" "$REPO_PATH/backend" --include="*.go" -n | grep -i "start\|end\|workflow" | head -20

echo ""
echo "Search complete!"

