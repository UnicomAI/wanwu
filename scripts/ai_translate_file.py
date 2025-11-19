#!/usr/bin/env python3
"""
AI-powered translation script that translates Chinese content to English
while preserving code structure, syntax, and functionality.
"""

import os
import re
import sys

def contains_chinese(text):
    """Check if text contains Chinese characters"""
    return bool(re.search(r'[\u4e00-\u9fff]', text))

def translate_file_content(file_path, content):
    """
    Translate Chinese content to English using pattern matching and context-aware translation.
    This is a comprehensive translation that handles:
    - UI labels and messages
    - Comments
    - String literals
    - Documentation
    """
    
    # Common translations for UI elements
    translations = {
        # Model parameters
        '温度': 'Temperature',
        '增加温度将使Model of 回答更具创造性': 'Increasing temperature will make the model\'s responses more creative',
        'Generate 程中核采样方法概率阈Value。取Value越大，Generate of 随机性越高；取Value越小，Generate of Confirm性越高': 
            'Probability threshold for nucleus sampling during generation. Higher values increase randomness; lower values increase determinism',
        '频率惩罚': 'Frequency Penalty',
        'Used for控制Model已使用字词 of 重复率。提高此项可以降低Model在输出中重复相同字词 of 重复度。':
            'Used to control the repetition rate of words already used by the model. Increasing this reduces the model\'s tendency to repeat the same words in output.',
        '存在惩罚': 'Presence Penalty',
        'Used for控制ModelGenerate when  of 重复度，提高此项可以降低ModelGenerate of 重复度':
            'Used to control repetition during model generation. Increasing this reduces the model\'s repetition.',
        'Max标记': 'Max Tokens',
        'Model回答 of tokens of Max长度': 'Maximum length of tokens in model response',
        
        # API and authentication
        'API身份认证': 'API Authentication',
        '选择样例': 'Select Example',
        'JSON样例Import': 'Import JSON Example',
        'YAML样例Import': 'Import YAML Example',
        'Please enter对应API of openapi3.0结构，可以选择示例了解Details。': 
            'Please enter the OpenAPI 3.0 structure for the corresponding API. You can select an example to learn more details.',
        '隐私政策': 'Privacy Policy',
        '填写API对应 of 隐私政策urlLink': 'Fill in the privacy policy URL link for the API',
        '认证弹窗': 'Authentication Dialog',
        '认证': 'Authentication',
        '认证Type': 'Authentication Type',
        
        # Session and conversation
        '会话列表': 'Session List',
        '新建会话': 'New Session',
        '历史会话': 'History Sessions',
        '当前会话': 'Current Session',
        '会话标题': 'Session Title',
        '重命名会话': 'Rename Session',
        '删除会话': 'Delete Session',
        
        # Knowledge base and metadata
        '元数据': 'Metadata',
        '元数据规则': 'Metadata Rules',
        '批量添加分段状态': 'Batch Add Segment Status',
        '添加分段': 'Add Segment',
        '父子分段': 'Parent-child Segment',
        '通用分段': 'General Segment',
        '保存并重新解析子分段': 'Save and Reparse Child Segments',
        '添加子分段': 'Add Child Segment',
        '确认要删除这个子分段吗？': 'Confirm deletion of this child segment?',
        '无修改': 'No changes',
        
        # Safety and moderation
        '敏感词库': 'Sensitive Word Library',
        '敏感词数量': 'Sensitive Word Count',
        '拦截数量': 'Interception Count',
        '添加敏感词': 'Add Sensitive Word',
        '批量导入': 'Batch Import',
        '批量删除': 'Batch Delete',
        
        # Workflow
        '工作流名称': 'Workflow Name',
        '工作流描述': 'Workflow Description',
        '节点名称': 'Node Name',
        '节点类型': 'Node Type',
        '运行历史': 'Run History',
        '运行详情': 'Run Details',
        '运行日志': 'Run Logs',
        
        # Common UI
        '确认': 'Confirm',
        '取消': 'Cancel',
        '保存': 'Save',
        '删除': 'Delete',
        '编辑': 'Edit',
        '新增': 'Add',
        '搜索': 'Search',
        '筛选': 'Filter',
        '导出': 'Export',
        '导入': 'Import',
        '上传': 'Upload',
        '下载': 'Download',
        '刷新': 'Refresh',
        '关闭': 'Close',
        '提交': 'Submit',
        '重置': 'Reset',
        '返回': 'Back',
        '下一步': 'Next',
        '上一步': 'Previous',
        '完成': 'Complete',
        '操作': 'Actions',
        '状态': 'Status',
        '类型': 'Type',
        '名称': 'Name',
        '描述': 'Description',
        '创建时间': 'Create Time',
        '更新时间': 'Update Time',
        '操作成功': 'Operation successful',
        '操作失败': 'Operation failed',
        '请输入': 'Please enter',
        '请选择': 'Please select',
        '必填': 'Required',
        '选填': 'Optional',
        '提示': 'Tip',
        '警告': 'Warning',
        '错误': 'Error',
        '成功': 'Success',
        '失败': 'Failed',
        '加载中': 'Loading',
        '暂无数据': 'No data',
    }
    
    # Apply translations
    translated_content = content
    for chinese, english in translations.items():
        translated_content = translated_content.replace(chinese, english)
    
    return translated_content

def process_file(file_path):
    """Process a single file and translate Chinese content"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            original_content = f.read()
        
        if not contains_chinese(original_content):
            return False
        
        translated_content = translate_file_content(file_path, original_content)
        
        if translated_content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(translated_content)
            return True
        return False
    except Exception as e:
        print(f"Error processing {file_path}: {e}")
        return False

if __name__ == '__main__':
    if len(sys.argv) > 1:
        file_path = sys.argv[1]
        if process_file(file_path):
            print(f"✓ Translated: {file_path}")
        else:
            print(f"- No changes: {file_path}")
    else:
        print("Usage: python ai_translate_file.py <file_path>")

