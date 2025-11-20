#!/bin/bash

echo "=========================================="
echo "Restarting Services for Category Updates"
echo "=========================================="
echo ""

# Navigate to project directory
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

echo "1. Restarting BFF Service (reads workflow config)..."
docker restart bff-service
sleep 3

echo ""
echo "2. Restarting Frontend (nginx) for UI updates..."
docker restart nginx-wanwu
sleep 3

echo ""
echo "3. Checking service status..."
docker ps --filter "name=bff-service" --format "table {{.Names}}\t{{.Status}}"
docker ps --filter "name=nginx-wanwu" --format "table {{.Names}}\t{{.Status}}"

echo ""
echo "=========================================="
echo "✅ Services restarted successfully!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Wait 30 seconds for services to fully start"
echo "2. Clear browser cache (Ctrl+F5 or Cmd+Shift+R)"
echo "3. Navigate to http://localhost:8081/templateSquare"
echo "4. You should see the new EHS categories!"
echo ""
echo "Categories should be:"
echo "  [All] [Incident Investigation] [Compliance]"
echo "  [Corrective Actions] [Training] [Audit]"
echo "  [Reporting] [Safety Inspection]"
echo ""
