#!/bin/bash

# Script to add volume mounts for translated JavaScript files to docker-compose.yaml

set -e

echo "===================================="
echo "Docker Compose Volume Mount Fix"
echo "===================================="

# Check if workflow-static-overrides directory has any files
if [ ! "$(ls -A project/nginx/workflow-static-overrides/js 2>/dev/null)" ]; then
    echo "ERROR: No translated files found in project/nginx/workflow-static-overrides/js"
    echo "Please run ./fix_workflow_nodes.sh first"
    exit 1
fi

echo -e "\n[1/3] Backing up docker-compose.yaml..."
cp docker-compose.yaml docker-compose.yaml.backup.$(date +%Y%m%d_%H%M%S)
echo "  ✓ Backup created"

echo -e "\n[2/3] Finding JavaScript files to mount..."
MOUNT_LINES=""
for file in project/nginx/workflow-static-overrides/js/*.js; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        # Remove the project/ prefix for docker-compose volume mount
        mount_line="      - ./project/nginx/workflow-static-overrides/js/${filename}:/usr/share/nginx/html/workflow/static/js/${filename}:ro"
        echo "  Will mount: $filename"
        MOUNT_LINES="${MOUNT_LINES}\n${mount_line}"
    fi
done

if [ -z "$MOUNT_LINES" ]; then
    echo "ERROR: No JavaScript files found to mount!"
    exit 1
fi

echo -e "\n[3/3] Updating docker-compose.yaml nginx service volumes..."

# Find the nginx service and add volume mounts
# We'll add them after the existing volumes using awk

python3 << 'EOF'
import sys

# Read the docker-compose.yaml file
with open('docker-compose.yaml', 'r') as f:
    lines = f.readlines()

# Find the nginx service volumes section
in_nginx_service = False
found_volumes = False
volumes_indent = None
output_lines = []
i = 0

while i < len(lines):
    line = lines[i]
    output_lines.append(line)

    # Check if we're entering nginx service
    if line.strip().startswith('nginx:'):
        in_nginx_service = True

    # Check if we found volumes section in nginx service
    if in_nginx_service and line.strip().startswith('volumes:'):
        found_volumes = True
        volumes_indent = len(line) - len(line.lstrip())

        # Add existing volumes
        i += 1
        while i < len(lines):
            next_line = lines[i]
            next_indent = len(next_line) - len(next_line.lstrip())

            # If we hit a line with same or less indentation and it's not a volume, we're done
            if next_line.strip() and next_indent <= volumes_indent:
                break

            output_lines.append(next_line)
            i += 1

        # Add comment for workflow static overrides
        output_lines.append(' ' * (volumes_indent + 2) + '# Workflow static file overrides for English translation\n')

        # Add new volume mounts for JavaScript files
        import os
        js_dir = 'project/nginx/workflow-static-overrides/js'
        if os.path.exists(js_dir):
            for filename in os.listdir(js_dir):
                if filename.endswith('.js'):
                    mount_line = ' ' * (volumes_indent + 2) + f'- ./project/nginx/workflow-static-overrides/js/{filename}:/usr/share/nginx/html/workflow/static/js/{filename}:ro\n'
                    output_lines.append(mount_line)
                    print(f'  Added mount for: {filename}')

        in_nginx_service = False
        continue

    # If we're past nginx service, stop looking
    if in_nginx_service and line.strip() and not line.strip().startswith('#') and line[0] not in ' \t' and line.strip() != 'nginx:':
        in_nginx_service = False

    i += 1

# Write the updated file
with open('docker-compose.yaml', 'w') as f:
    f.writelines(output_lines)

print('  ✓ docker-compose.yaml updated successfully')
EOF

echo -e "\n===================================="
echo "Docker Compose Update Complete! ✓"
echo "===================================="
echo ""
echo "Next steps:"
echo "  1. Review the changes: git diff docker-compose.yaml"
echo "  2. Restart nginx: docker-compose restart nginx"
echo "  3. Test by creating a new workflow"
echo ""
