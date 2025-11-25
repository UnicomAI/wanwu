#!/bin/bash

# Fix nginx 403 error by deploying updated config
# This script updates the nginx configuration to proxy /workflow/ to the workflow service

set -e

echo "=========================================="
echo "Fixing nginx 403 Error"
echo "=========================================="
echo ""

# Check if we're on AWS or local
if [ -f "/home/ubuntu/.aws-deployment" ]; then
    echo "✓ Running on AWS deployment"
    IS_AWS=true
else
    echo "✓ Running locally"
    IS_AWS=false
fi

echo ""
echo "Step 1: Checking current nginx configuration..."
echo "----------------------------------------------"

# Check current config
CURRENT_CONFIG=$(docker exec nginx-wanwu cat /etc/nginx/conf.d/aibase.conf 2>/dev/null | grep -A 3 "location.*workflow" | head -5 || echo "ERROR")

if echo "$CURRENT_CONFIG" | grep -q "proxy_pass.*workflow-wanwu:8999"; then
    echo "✓ nginx config is already correct (proxying to workflow service)"
    echo ""
    echo "The 403 error might be from a different cause."
    echo "Please check:"
    echo "  1. Is the workflow-wanwu container running?"
    echo "  2. Can you access http://workflow-wanwu:8999 from inside nginx container?"
    echo ""
    exit 0
elif echo "$CURRENT_CONFIG" | grep -q "try_files"; then
    echo "✗ nginx config is OLD (trying to serve static files)"
    echo ""
    echo "Current config:"
    echo "$CURRENT_CONFIG"
    echo ""
elif echo "$CURRENT_CONFIG" | grep -q "ERROR"; then
    echo "✗ Could not read nginx config (container might not be running)"
    exit 1
fi

echo "Step 2: Deploying updated nginx configuration..."
echo "----------------------------------------------"

# Pull latest changes
echo "Pulling latest code..."
git pull origin main

# Restart nginx with new config
echo "Restarting nginx container with updated config..."
docker compose up -d nginx

echo ""
echo "Step 3: Verifying the fix..."
echo "----------------------------------------------"

# Wait for nginx to start
sleep 3

# Check new config
NEW_CONFIG=$(docker exec nginx-wanwu cat /etc/nginx/conf.d/aibase.conf | grep -A 3 "location.*workflow" | head -5)

if echo "$NEW_CONFIG" | grep -q "proxy_pass.*workflow-wanwu:8999"; then
    echo "✓ nginx config updated successfully!"
    echo ""
    echo "New config:"
    echo "$NEW_CONFIG"
    echo ""
    echo "=========================================="
    echo "✓ Fix Applied Successfully!"
    echo "=========================================="
    echo ""
    echo "Next steps:"
    echo "  1. Clear your browser cache (Ctrl+Shift+R or Cmd+Shift+R)"
    echo "  2. Try accessing the workflow again:"
    echo "     https://app.safvr.com/aibase/workflow?id=7576662999398612992"
    echo ""
else
    echo "✗ nginx config still not correct"
    echo ""
    echo "New config:"
    echo "$NEW_CONFIG"
    echo ""
    echo "Manual fix required. Please check docker-compose.yaml volumes section."
    exit 1
fi

