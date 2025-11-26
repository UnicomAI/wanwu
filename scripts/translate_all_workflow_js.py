#!/usr/bin/env python3
"""
Translate all Chinese text to English in workflow frontend JavaScript files.
This script processes all JS files in workflow-frontend/static/js that contain Chinese text.
"""

import os
import re
from pathlib import Path

# Define translations - order matters! Do longer phrases first
TRANSLATIONS = [
    ('私密发布为工具：仅自己可见', 'Private publish as tool: Only visible to yourself'),
    ('公开发布为工具：组织内可见', 'Public publish as tool: Visible within organization'),
    ('公开发布为工具：所有人可见', 'Public publish as tool: Visible to everyone'),
    ('工作流的起始节点，用于设定启动工作流需要的信息', 'The starting node of the workflow, used to set the information needed to start the workflow'),
    ('工作流的最终节点，用于返回工作流运行后的结果信息', 'The final node of the workflow, used to return the result information after the workflow runs'),
    ('发布类型', 'Publish Type'),
    ('返回变量', 'Return variables'),
    ('返回文本', 'Return text'),
    ('输出变量', 'Output variable'),
    ('变量名', 'Variable name'),
    ('变量值', 'Variable value'),
    ('测试运行', 'Test run'),
    ('结束', 'End'),
    ('开始', 'Start'),
    ('取消', 'Cancel'),
    ('确定', 'OK'),
]

def translate_file(file_path):
    """Translate Chinese text to English in a single file."""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_content = content
        translations_applied = {}
        
        # Apply all translations
        for chinese, english in TRANSLATIONS:
            if chinese in content:
                count = content.count(chinese)
                content = content.replace(chinese, english)
                translations_applied[chinese] = (english, count)
        
        # Only write if changes were made
        if content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(content)
            return translations_applied
        return None
    except Exception as e:
        print(f"Error processing {file_path}: {e}")
        return None

def find_js_files_with_chinese(base_dir):
    """Find all JS files containing Chinese text."""
    js_files = []
    base_path = Path(base_dir)
    
    # Common Chinese characters pattern
    chinese_pattern = re.compile(r'[\u4e00-\u9fff]+')
    
    for js_file in base_path.rglob('*.js'):
        try:
            with open(js_file, 'r', encoding='utf-8') as f:
                content = f.read()
                if chinese_pattern.search(content):
                    js_files.append(js_file)
        except Exception:
            pass
    
    return js_files

def main():
    workflow_frontend_dir = Path(__file__).parent.parent / 'workflow-frontend' / 'static' / 'js'
    
    if not workflow_frontend_dir.exists():
        print(f"Error: {workflow_frontend_dir} does not exist")
        return
    
    print(f"Searching for JS files with Chinese text in {workflow_frontend_dir}...")
    js_files = find_js_files_with_chinese(workflow_frontend_dir)
    
    if not js_files:
        print("No JS files with Chinese text found.")
        return
    
    print(f"Found {len(js_files)} files to translate:")
    for f in js_files:
        print(f"  - {f.relative_to(workflow_frontend_dir.parent.parent)}")
    
    print("\nTranslating files...")
    total_translations = 0
    files_modified = 0
    
    for js_file in js_files:
        result = translate_file(js_file)
        if result:
            files_modified += 1
            file_translations = sum(count for _, count in result.values())
            total_translations += file_translations
            print(f"  ✓ {js_file.name}: {file_translations} translations")
            for chinese, (english, count) in result.items():
                print(f"    - '{chinese}' → '{english}' ({count}x)")
    
    print(f"\n✅ Translation complete!")
    print(f"   Files modified: {files_modified}")
    print(f"   Total translations: {total_translations}")

if __name__ == '__main__':
    main()

