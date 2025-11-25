#!/bin/bash
# Script to translate Chinese text to English in workflow frontend JavaScript files
# This script is run during the Docker build process to ensure all UI text is in English

set -e

WORKFLOW_JS_DIR="${1:-workflow-frontend/static/js}"

echo "🌐 Translating workflow frontend JavaScript files from Chinese to English..."
echo "📁 Target directory: $WORKFLOW_JS_DIR"

# Find the main index file that contains most of the Chinese text
INDEX_FILE=$(find "$WORKFLOW_JS_DIR" -name "index~0*.js" | head -1)

if [ -z "$INDEX_FILE" ]; then
    echo "⚠️  Warning: Could not find index~0*.js file"
    exit 0
fi

echo "📝 Translating file: $INDEX_FILE"

# Create a backup
cp "$INDEX_FILE" "$INDEX_FILE.bak"

# Apply translations using sed
# Note: We use a temporary file to avoid issues with in-place editing
TEMP_FILE=$(mktemp)

cat "$INDEX_FILE" | \
# Publish dialog translations
sed 's/发布类型/Publish Type/g' | \
sed 's/私密发布为工具：仅自己可见/Private publish as tool: Only visible to yourself/g' | \
sed 's/公开发布为工具：组织内可见/Public publish as tool: Visible within organization/g' | \
sed 's/公开发布为工具：所有人可见/Public publish as tool: Visible to everyone/g' | \
sed 's/私密发布为应用：仅自己可见/Private publish as application: Only visible to yourself/g' | \
sed 's/公开发布为应用：组织内可见/Public publish as application: Visible within organization/g' | \
sed 's/公开发布为应用：所有人可见/Public publish as application: Visible to everyone/g' | \
# Node name translations
sed 's/"title":"开始"/"title":"Start"/g' | \
sed 's/"title":"结束"/"title":"End"/g' | \
# Node description translations  
sed 's/工作流的起始节点，用于设置启动工作流所需的信息/The starting node of the workflow, used to set the information needed to start the workflow/g' | \
sed 's/工作流的最终节点，用于返回工作流运行后的结果信息/The final node of the workflow, used to return the result information after the workflow runs/g' | \
# Common UI text translations
sed 's/取消/Cancel/g' | \
sed 's/确认/Confirm/g' | \
sed 's/确定/OK/g' | \
sed 's/删除/Delete/g' | \
sed 's/编辑/Edit/g' | \
sed 's/保存/Save/g' | \
sed 's/复制/Copy/g' | \
sed 's/添加/Add/g' | \
sed 's/创建/Create/g' | \
sed 's/关闭/Close/g' | \
sed 's/输入/Input/g' | \
sed 's/输出/Output/g' | \
sed 's/配置/Configuration/g' | \
sed 's/操作/Operation/g' | \
sed 's/类型/Type/g' | \
sed 's/变量名/Variable name/g' | \
sed 's/请输入/Please enter/g' | \
sed 's/请稍后重试/Please try again later/g' | \
sed 's/工作流/Workflow/g' | \
sed 's/模型/Model/g' | \
sed 's/智能体/Agent/g' | \
sed 's/应用/Application/g' | \
sed 's/插件/Plugin/g' | \
sed 's/知识库/Knowledge Base/g' | \
sed 's/自定义/Custom/g' | \
sed 's/全部/All/g' | \
sed 's/使用/Use/g' | \
sed 's/扣子/Coze/g' | \
sed 's/个/items/g' | \
sed 's/天/days/g' \
> "$TEMP_FILE"

# Replace the original file with the translated version
mv "$TEMP_FILE" "$INDEX_FILE"

echo "✅ Translation complete!"
echo "📊 Backup saved to: $INDEX_FILE.bak"

# Verify the translation worked
if grep -q "Publish Type" "$INDEX_FILE"; then
    echo "✅ Verification passed: English text found in translated file"
else
    echo "⚠️  Warning: Could not verify translation"
fi

