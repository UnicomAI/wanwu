#!/bin/bash

# Deploy nginx 403 fix via GitHub Actions + Dokploy
# This script commits the changes and pushes to trigger the CI/CD pipeline

set -e

echo "=========================================="
echo "Deploy nginx 403 Fix"
echo "=========================================="
echo ""

# Check if there are changes to commit
if git diff --quiet && git diff --cached --quiet; then
    echo "✓ No uncommitted changes detected"
    echo ""
    echo "Checking if we're ahead of origin..."
    
    if git rev-parse @{u} > /dev/null 2>&1; then
        LOCAL=$(git rev-parse @)
        REMOTE=$(git rev-parse @{u})
        
        if [ "$LOCAL" = "$REMOTE" ]; then
            echo "✓ Already up to date with origin"
            echo ""
            echo "The fix has already been pushed. Checking deployment status..."
            echo ""
            echo "Next steps:"
            echo "  1. Check GitHub Actions: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/actions"
            echo "  2. Wait for the workflow to complete"
            echo "  3. Dokploy will automatically redeploy"
            echo "  4. Test: https://app.safvr.com/aibase/workflow?id=7576662999398612992"
            exit 0
        else
            echo "✓ Local commits ready to push"
        fi
    fi
else
    echo "✓ Uncommitted changes detected"
    echo ""
    echo "Files changed:"
    git status --short
    echo ""
    
    # Ask for confirmation
    read -p "Commit and push these changes? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi
    
    # Stage all changes
    git add .
    
    # Commit
    git commit -m "Fix nginx 403 and 404 errors for workflow access

- Updated docker-compose.yaml to mount nginx config as volume
- Updated GitHub Actions to rebuild frontend when nginx config changes
- Fixed nginx proxy_pass: removed trailing slash to preserve full path
- nginx now proxies /workflow/ to workflow-wanwu:8999 (without trailing slash)

This fixes both the 403 Forbidden and 404 Not Found errors when accessing workflows."
    
    echo ""
    echo "✓ Changes committed"
fi

echo ""
echo "Pushing to GitHub..."
git push origin main

echo ""
echo "=========================================="
echo "✓ Deployment Triggered!"
echo "=========================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Monitor GitHub Actions:"
echo "   https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/actions"
echo ""
echo "2. The workflow will:"
echo "   - Rebuild safvr/wanwu-frontend:latest (includes updated nginx config)"
echo "   - Push to Docker Hub"
echo "   - Trigger Dokploy deployment"
echo ""
echo "3. Wait ~5 minutes for deployment to complete"
echo ""
echo "4. Test the fix:"
echo "   - Clear browser cache (Ctrl+Shift+R or Cmd+Shift+R)"
echo "   - Access: https://app.safvr.com/aibase/workflow?id=7576662999398612992"
echo ""
echo "5. Verify nginx config on server:"
echo "   ssh ubuntu@34.238.142.181"
echo "   docker exec nginx-wanwu cat /etc/nginx/conf.d/aibase.conf | grep -A 5 'location.*workflow'"
echo ""

