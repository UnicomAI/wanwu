#!/bin/bash

echo "=========================================="
echo "  Updating Frontend with Latest Changes"
echo "=========================================="
echo ""

# Navigate to web directory
cd /Users/mohankumarv/Desktop/SAFVR/wanwu/web

echo "1. Building frontend (this may take 1-2 minutes)..."
npm run build

if [ $? -ne 0 ]; then
    echo "❌ Build failed! Check errors above."
    exit 1
fi

echo ""
echo "2. Copying new build to nginx container..."
docker cp dist/. nginx-wanwu:/usr/share/nginx/html/aibase/

if [ $? -ne 0 ]; then
    echo "❌ Copy failed! Is nginx container running?"
    exit 1
fi

echo ""
echo "3. Restarting nginx..."
docker restart nginx-wanwu

echo ""
echo "4. Waiting for nginx to start..."
sleep 10

echo ""
echo "=========================================="
echo "  ✅ Frontend Updated Successfully!"
echo "=========================================="
echo ""
echo "IMPORTANT: Clear your browser cache!"
echo "  - Windows/Linux: Ctrl + F5"
echo "  - Mac: Cmd + Shift + R"
echo ""
echo "Then navigate to:"
echo "  http://localhost:8081/templateSquare"
echo ""
echo "You should see the updated categories!"
echo ""
