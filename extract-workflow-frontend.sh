#!/bin/bash

# Extract workflow frontend from original Aliyun image
# This script should be run on the AWS server

set -e

echo "=========================================="
echo "Extract Workflow Frontend"
echo "=========================================="
echo ""

ORIGINAL_IMAGE="crpi-6pj79y7ddzdpexs8.cn-hangzhou.personal.cr.aliyuncs.com/wanwulite/frontend:v0.2.7-e3a5f7d9"
TEMP_CONTAINER="temp-frontend-extract"
EXTRACT_DIR="/tmp/workflow-frontend"

echo "Step 1: Pull original frontend image..."
docker pull "$ORIGINAL_IMAGE"

echo ""
echo "Step 2: Create temporary container..."
docker create --name "$TEMP_CONTAINER" "$ORIGINAL_IMAGE"

echo ""
echo "Step 3: Extract workflow frontend files..."
rm -rf "$EXTRACT_DIR"
docker cp "$TEMP_CONTAINER":/usr/share/nginx/html/workflow "$EXTRACT_DIR"

echo ""
echo "Step 4: Verify extracted files..."
echo "Files in $EXTRACT_DIR:"
ls -la "$EXTRACT_DIR"
echo ""
echo "Checking for index.html:"
if [ -f "$EXTRACT_DIR/index.html" ]; then
    echo "✓ index.html found!"
    head -5 "$EXTRACT_DIR/index.html"
else
    echo "✗ index.html NOT found - this might be a problem"
fi

echo ""
echo "Step 5: Copy to running nginx container..."
docker cp "$EXTRACT_DIR"/. nginx-wanwu:/usr/share/nginx/html/workflow/

echo ""
echo "Step 6: Cleanup temporary container..."
docker rm "$TEMP_CONTAINER"

echo ""
echo "Step 7: Verify files in nginx container..."
echo "Files in nginx container:"
docker exec nginx-wanwu ls -la /usr/share/nginx/html/workflow/

echo ""
echo "Checking for index.html in nginx:"
if docker exec nginx-wanwu test -f /usr/share/nginx/html/workflow/index.html; then
    echo "✓ index.html exists in nginx container!"
else
    echo "✗ index.html NOT found in nginx container"
fi

echo ""
echo "=========================================="
echo "✓ Extraction Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Test in browser:"
echo "   https://app.safvr.com/aibase/workflow"
echo ""
echo "2. If it works, save the files to your repository:"
echo "   mkdir -p workflow-frontend"
echo "   docker cp nginx-wanwu:/usr/share/nginx/html/workflow/. workflow-frontend/"
echo ""
echo "3. Update Dockerfile.frontend to include these files:"
echo "   COPY ./workflow-frontend /usr/share/nginx/html/workflow"
echo ""
echo "4. Commit and push to make it permanent"
echo ""

