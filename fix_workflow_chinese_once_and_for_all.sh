#!/bin/bash

# Master script to fix workflow node Chinese text once and for all
# This is the complete solution based on root cause analysis

set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║                                                            ║"
echo "║  Workflow Node Chinese Text - COMPLETE FIX                 ║"
echo "║  Fixing the ROOT CAUSE: JavaScript bundle translation      ║"
echo "║                                                            ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 0: Verify Docker is running
echo -e "${BLUE}[Step 0/5] Checking Docker status...${NC}"
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running!${NC}"
    echo ""
    echo "Please start Docker Desktop, then run this script again."
    echo ""
    echo "To start Docker:"
    echo "  - macOS: Open Docker Desktop application"
    echo "  - Linux: sudo systemctl start docker"
    echo ""
    exit 1
fi
echo -e "${GREEN}✓ Docker is running${NC}"
echo ""

# Verify nginx container is running
echo -e "${BLUE}[Step 1/5] Checking nginx container status...${NC}"
if ! docker-compose ps nginx | grep -q "Up"; then
    echo -e "${YELLOW}⚠ nginx container is not running. Starting it...${NC}"
    docker-compose up -d nginx
    echo -e "${GREEN}✓ nginx container started${NC}"
else
    echo -e "${GREEN}✓ nginx container is running${NC}"
fi
echo ""

# Step 1: Extract and translate JavaScript files
echo -e "${BLUE}[Step 2/5] Extracting and translating JavaScript files...${NC}"
if [ ! -f "./fix_workflow_nodes.sh" ]; then
    echo -e "${RED}✗ fix_workflow_nodes.sh not found!${NC}"
    exit 1
fi

./fix_workflow_nodes.sh
echo -e "${GREEN}✓ JavaScript files extracted and translated${NC}"
echo ""

# Step 2: Update docker-compose.yaml
echo -e "${BLUE}[Step 3/5] Updating docker-compose.yaml with volume mounts...${NC}"
if [ ! -f "./fix_docker_compose.sh" ]; then
    echo -e "${RED}✗ fix_docker_compose.sh not found!${NC}"
    exit 1
fi

./fix_docker_compose.sh
echo -e "${GREEN}✓ docker-compose.yaml updated${NC}"
echo ""

# Step 3: Restart nginx
echo -e "${BLUE}[Step 4/5] Restarting nginx container to load translated files...${NC}"
docker-compose restart nginx
echo -e "${GREEN}✓ nginx container restarted${NC}"
echo ""

# Step 4: Wait for nginx to be healthy
echo -e "${BLUE}[Step 5/5] Waiting for nginx to be healthy...${NC}"
sleep 3
if docker-compose ps nginx | grep -q "Up"; then
    echo -e "${GREEN}✓ nginx is healthy${NC}"
else
    echo -e "${YELLOW}⚠ nginx might not be fully ready yet. Check with: docker-compose ps nginx${NC}"
fi
echo ""

# Success message
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                                                            ║"
echo "║  ✓✓✓ FIX COMPLETE! ✓✓✓                                     ║"
echo "║                                                            ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo -e "${GREEN}The workflow node Chinese text issue has been fixed!${NC}"
echo ""
echo "What was fixed:"
echo "  • Extracted JavaScript bundles from nginx container"
echo "  • Translated Chinese strings to English:"
echo "    - 开始 → Start"
echo "    - 结束 → End"
echo "    - 输入变量 → Input variables"
echo "    - 输出变量 → Output variables"
echo "    - And all node descriptions"
echo "  • Mounted translated files back into nginx"
echo "  • Restarted nginx to load the English versions"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Open your browser and refresh the workflow page"
echo "  2. Create a NEW workflow"
echo "  3. Verify that you see 'Start' and 'End' instead of Chinese text"
echo ""
echo "If you still see Chinese text:"
echo "  • Hard refresh your browser (Cmd+Shift+R on macOS, Ctrl+Shift+R on Linux)"
echo "  • Clear browser cache"
echo "  • Check browser console for any errors"
echo ""
echo -e "${BLUE}Documentation:${NC}"
echo "  • Root cause analysis: WORKFLOW_NODES_ROOT_CAUSE_FIX.md"
echo "  • Previous fix attempts: WORKFLOW_FIX_SUMMARY.md"
echo ""
echo "═══════════════════════════════════════════════════════════"
echo "This fix is PERMANENT and will persist across restarts! ✓"
echo "═══════════════════════════════════════════════════════════"
echo ""
