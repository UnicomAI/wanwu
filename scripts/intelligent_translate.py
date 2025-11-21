#!/usr/bin/env python3
"""
Intelligent translation script that handles mixed Chinese-English text patterns.
Uses regex to find and replace complex patterns.
"""

import os
import re

def translate_file(file_path):
    """Translate a file using regex patterns"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_content = content
        
        # Pattern-based translations for mixed Chinese-English
        patterns = [
            # Model descriptions
            (r'增加温度将使Model of 回答更具创造性', 'Increasing temperature will make the model\'s responses more creative'),
            (r'增加Temperature将使Model of 回答更具创造性', 'Increasing temperature will make the model\'s responses more creative'),
            (r'Generate 程中核采样方法概率阈Value。取Value越大，Generate of 随机性越高；取Value越小，Generate of Confirm性越高',
             'Probability threshold for nucleus sampling during generation. Higher values increase randomness; lower values increase determinism'),
            (r'Used for控制Model已使用字词 of 重复率。提高此项可以降低Model在输出中重复相同字词 of 重复度。',
             'Used to control the repetition rate of words already used by the model. Increasing this reduces repetition of the same words in output.'),
            (r'Used for控制ModelGenerate when  of 重复度，提高此项可以降低ModelGenerate of 重复度',
             'Used to control repetition during model generation. Increasing this reduces repetition.'),
            (r'Model回答 of tokens of Max长度', 'Maximum length of tokens in model response'),
            
            # API and authentication
            (r'API身份认证', 'API Authentication'),
            (r'选择样例', 'Select Example'),
            (r'JSON样例Import', 'Import JSON Example'),
            (r'YAML样例Import', 'Import YAML Example'),
            (r'模板样例Import', 'Import Template Example'),
            (r'Please enter对应API of openapi3\.0结构，可以选择示例了解Details。',
             'Please enter the OpenAPI 3.0 structure for the corresponding API. You can select an example to learn more details.'),
            (r'隐私政策', 'Privacy Policy'),
            (r'填写API对应 of 隐私政策urlLink', 'Fill in the privacy policy URL link for the API'),
            (r'认证弹窗', 'Authentication Dialog'),
            (r'认证Type', 'Authentication Type'),
            
            # Session component
            (r'会话列表', 'Session List'),
            (r'新建会话', 'New Session'),
            (r'历史会话', 'History Sessions'),
            (r'当前会话', 'Current Session'),
            (r'会话标题', 'Session Title'),
            (r'重命名会话', 'Rename Session'),
            (r'删除会话', 'Delete Session'),
            
            # Knowledge and metadata
            (r'元数据', 'Metadata'),
            (r'元Data', 'Metadata'),
            (r'元数据规则', 'Metadata Rules'),
            (r'元Data规Then', 'Metadata Rules'),
            (r'批量添加分段状态', 'Batch Add Segment Status'),
            (r'BatchAdd分段Status', 'Batch Add Segment Status'),
            (r'添加分段', 'Add Segment'),
            (r'Add分段', 'Add Segment'),
            (r'父子分段', 'Parent-child Segment'),
            (r'通用分段', 'General Segment'),
            (r'保存并重新解析子分段', 'Save and Reparse Child Segments'),
            (r'Save并重新Parse子分段', 'Save and Reparse Child Segments'),
            (r'添加子分段', 'Add Child Segment'),
            (r'Add子分段', 'Add Child Segment'),
            (r'确认要删除这个子分段吗？', 'Confirm deletion of this child segment?'),
            (r'Confirm要Delete这个子分段吗？', 'Confirm deletion of this child segment?'),
            (r'无修改', 'No changes'),
            (r'无Modify', 'No changes'),
            (r'无数据', 'No data'),
            (r'无Data', 'No data'),
            (r'个子分段', ' child segments'),
            
            # Safety
            (r'敏感词', 'Sensitive Word'),
            (r'敏感词库', 'Sensitive Word Library'),
            (r'安全护栏', 'Safety Guardrail'),
            (r'拦截规则', 'Interception Rule'),
            (r'拦截规Then', 'Interception Rule'),
            
            # Workflow
            (r'工作流', 'Workflow'),
            (r'节点', 'Node'),
            (r'运行历史', 'Run History'),
            (r'运行详情', 'Run Details'),
            (r'运行日志', 'Run Logs'),
            
            # Common UI
            (r'确认', 'Confirm'),
            (r'取消', 'Cancel'),
            (r'保存', 'Save'),
            (r'删除', 'Delete'),
            (r'编辑', 'Edit'),
            (r'新增', 'Add'),
            (r'添加', 'Add'),
            (r'搜索', 'Search'),
            (r'筛选', 'Filter'),
            (r'导出', 'Export'),
            (r'导入', 'Import'),
            (r'上传', 'Upload'),
            (r'下载', 'Download'),
            (r'刷新', 'Refresh'),
            (r'关闭', 'Close'),
            (r'提交', 'Submit'),
            (r'重置', 'Reset'),
            (r'返回', 'Back'),
            (r'操作', 'Actions'),
            (r'状态', 'Status'),
            (r'类型', 'Type'),
            (r'名称', 'Name'),
            (r'描述', 'Description'),
            (r'详情', 'Details'),
            (r'提示', 'Tip'),
            (r'警告', 'Warning'),
            (r'错误', 'Error'),
            (r'成功', 'Success'),
            (r'失败', 'Failed'),
            (r'加载中', 'Loading'),
            (r'暂无数据', 'No data'),
            (r'请输入', 'Please enter'),
            (r'请选择', 'Please select'),
            (r'必填', 'Required'),
            (r'选填', 'Optional'),
        ]
        
        # Apply all patterns
        for pattern, replacement in patterns:
            content = re.sub(pattern, replacement, content)
        
        # Write back if changed
        if content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(content)
            return True
        return False
        
    except Exception as e:
        print(f"Error processing {file_path}: {e}")
        return False

def main():
    """Process all Vue and JS files"""
    import sys
    
    if len(sys.argv) > 1:
        # Single file mode
        file_path = sys.argv[1]
        if translate_file(file_path):
            print(f"✓ Translated: {file_path}")
        else:
            print(f"- No changes: {file_path}")
    else:
        # Batch mode
        files_changed = 0
        for root, dirs, files in os.walk('web/src'):
            if 'lang' in dirs:
                dirs.remove('lang')
            if 'node_modules' in dirs:
                dirs.remove('node_modules')
            
            for file in files:
                if file.endswith(('.vue', '.js')):
                    file_path = os.path.join(root, file)
                    if translate_file(file_path):
                        files_changed += 1
                        print(f"✓ {file_path}")
        
        print(f"\nTranslated {files_changed} files")

if __name__ == '__main__':
    main()

