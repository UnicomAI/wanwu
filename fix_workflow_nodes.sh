#!/bin/bash

# Workflow Node Chinese Text Fix Script
# This script extracts JavaScript files from nginx, finds Chinese text, and creates English versions

set -e

echo "===================================="
echo "Workflow Node Translation Fix"
echo "===================================="

# Create override directory if it doesn't exist
mkdir -p project/nginx/workflow-static-overrides/js

# Step 1: Find all JavaScript files in the workflow static directory
echo -e "\n[1/5] Finding all JavaScript files in workflow static directory..."
JS_FILES=$(docker exec nginx-wanwu find /usr/share/nginx/html/workflow/static/js -name "*.js" -type f)
echo "Found $(echo "$JS_FILES" | wc -l) JavaScript files"

# Step 2: Search for files containing Chinese text
echo -e "\n[2/5] Searching for files with Chinese text '开始' (Start) and '结束' (End)..."
FILES_WITH_CHINESE=""
for file in $JS_FILES; do
    if docker exec nginx-wanwu grep -q "开始" "$file" 2>/dev/null; then
        echo "  Found Chinese text in: $file"
        FILES_WITH_CHINESE="$FILES_WITH_CHINESE $file"
    fi
done

if [ -z "$FILES_WITH_CHINESE" ]; then
    echo "  No files with Chinese text found - checking for alternative patterns..."
    # Try searching for other possible Chinese strings
    for file in $JS_FILES; do
        if docker exec nginx-wanwu grep -q "工作流" "$file" 2>/dev/null; then
            echo "  Found Chinese workflow text in: $file"
            FILES_WITH_CHINESE="$FILES_WITH_CHINESE $file"
        fi
    done
fi

if [ -z "$FILES_WITH_CHINESE" ]; then
    echo "  ERROR: No JavaScript files with Chinese text found!"
    echo "  This might mean:"
    echo "  1. The Chinese text is in a different format (minified differently)"
    echo "  2. The workflow service version has changed"
    echo "  3. The files have already been translated"
    exit 1
fi

# Step 3: Extract and translate each file
echo -e "\n[3/5] Extracting and translating JavaScript files..."
for file in $FILES_WITH_CHINESE; do
    filename=$(basename "$file")
    echo "  Processing: $filename"

    # Extract the file
    docker exec nginx-wanwu cat "$file" > "project/nginx/workflow-static-overrides/js/${filename}.tmp"

    # Replace Chinese strings with English
    # Node titles
    sed -i.bak 's/开始/Start/g' "project/nginx/workflow-static-overrides/js/${filename}.tmp"
    sed -i.bak 's/结束/End/g' "project/nginx/workflow-static-overrides/js/${filename}.tmp"

    # Node descriptions (common ones found in workflow systems)
    sed -i.bak 's/工作流的起始节点，用于设定起始的工序要素/Workflow start node, used to set initial process elements/g' "project/nginx/workflow-static-overrides/js/${filename}.tmp"
    sed -i.bak 's/工作流的结束节点，用于设定工作流运行后的结果信息/Workflow end node, used to set result information after workflow execution/g' "project/nginx/workflow-static-overrides/js/${filename}.tmp"

    # Additional common Chinese strings in workflow nodes
    sed -i.bak 's/输入变量/Input variables/g' "project/nginx/workflow-static-overrides/js/${filename}.tmp"
    sed -i.bak 's/输出变量/Output variables/g' "project/nginx/workflow-static-overrides/js/${filename}.tmp"
    sed -i.bak 's/返回变量/Return variables/g' "project/nginx/workflow-static-overrides/js/${filename}.tmp"

    # Move to final location
    mv "project/nginx/workflow-static-overrides/js/${filename}.tmp" "project/nginx/workflow-static-overrides/js/${filename}"

    # Remove backup files
    rm -f "project/nginx/workflow-static-overrides/js/${filename}.tmp.bak"

    echo "  ✓ Created translated version: project/nginx/workflow-static-overrides/js/${filename}"
done

# Step 4: Display files to be mounted
echo -e "\n[4/5] Files ready for mounting:"
ls -lh project/nginx/workflow-static-overrides/js/

# Step 5: Show next steps
echo -e "\n[5/5] Next steps:"
echo "  1. The fix_docker_compose.sh script will update docker-compose.yaml with volume mounts"
echo "  2. Restart the nginx container: docker-compose restart nginx"
echo "  3. Test by creating a new workflow"

echo -e "\n===================================="
echo "Translation complete! ✓"
echo "===================================="
