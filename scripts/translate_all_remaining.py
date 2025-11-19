#!/usr/bin/env python3
"""
Master script to translate all remaining Chinese content in the codebase.
Processes files one by one and applies comprehensive translations.
"""

import os
import re
import sys

def contains_chinese(text):
    """Check if text contains Chinese characters"""
    return bool(re.search(r'[\u4e00-\u9fff]', text))

def get_files_with_chinese(directory, extensions):
    """Get all files with Chinese content"""
    files_with_chinese = []
    
    for root, dirs, files in os.walk(directory):
        # Skip certain directories
        if 'lang' in dirs:
            dirs.remove('lang')
        if 'node_modules' in dirs:
            dirs.remove('node_modules')
        if '.git' in dirs:
            dirs.remove('.git')
        
        for file in files:
            if any(file.endswith(ext) for ext in extensions):
                file_path = os.path.join(root, file)
                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        content = f.read()
                    if contains_chinese(content):
                        files_with_chinese.append(file_path)
                except:
                    pass
    
    return files_with_chinese

def translate_content(content):
    """Comprehensive translation of Chinese to English"""
    
    # Dictionary of translations - comprehensive list
    translations = {
        # Model parameters and settings
        '温度': 'Temperature',
        '增加温度将使Model of 回答更具创造性': 'Increasing temperature will make the model\'s responses more creative',
        '增加温度将使模型的回答更具创造性': 'Increasing temperature will make the model\'s responses more creative',
        'Generate 程中核采样方法概率阈Value。取Value越大，Generate of 随机性越高；取Value越小，Generate of Confirm性越高': 
            'Probability threshold for nucleus sampling during generation. Higher values increase randomness; lower values increase determinism',
        '生成过程中核采样方法概率阈值。取值越大，生成的随机性越高；取值越小，生成的确定性越高':
            'Probability threshold for nucleus sampling during generation. Higher values increase randomness; lower values increase determinism',
        '频率惩罚': 'Frequency Penalty',
        'Used for控制Model已使用字词 of 重复率。提高此项可以降低Model在输出中重复相同字词 of 重复度。':
            'Used to control the repetition rate of words already used by the model. Increasing this reduces repetition of the same words in output.',
        '用于控制模型已使用字词的重复率。提高此项可以降低模型在输出中重复相同字词的重复度。':
            'Used to control the repetition rate of words already used by the model. Increasing this reduces repetition of the same words in output.',
        '存在惩罚': 'Presence Penalty',
        'Used for控制ModelGenerate when  of 重复度，提高此项可以降低ModelGenerate of 重复度':
            'Used to control repetition during model generation. Increasing this reduces repetition.',
        '用于控制模型生成时的重复度，提高此项可以降低模型生成的重复度':
            'Used to control repetition during model generation. Increasing this reduces repetition.',
        'Max标记': 'Max Tokens',
        '最大标记': 'Max Tokens',
        'Model回答 of tokens of Max长度': 'Maximum length of tokens in model response',
        '模型回答的tokens的最大长度': 'Maximum length of tokens in model response',
        
        # API and authentication
        'API身份认证': 'API Authentication',
        '选择样例': 'Select Example',
        '模板样例Import': 'Import Template Example',
        '模板样例导入': 'Import Template Example',
        'JSON样例Import': 'Import JSON Example',
        'JSON样例导入': 'Import JSON Example',
        'YAML样例Import': 'Import YAML Example',
        'YAML样例导入': 'Import YAML Example',
        'Please enter对应API of openapi3.0结构，可以选择示例了解Details。': 
            'Please enter the OpenAPI 3.0 structure for the corresponding API. You can select an example to learn more details.',
        '请输入对应API的openapi3.0结构，可以选择示例了解详情。':
            'Please enter the OpenAPI 3.0 structure for the corresponding API. You can select an example to learn more details.',
        '隐私政策': 'Privacy Policy',
        '填写API对应 of 隐私政策urlLink': 'Fill in the privacy policy URL link for the API',
        '填写API对应的隐私政策url链接': 'Fill in the privacy policy URL link for the API',
        '认证弹窗': 'Authentication Dialog',
        '认证': 'Authentication',
        '认证Type': 'Authentication Type',
        '认证类型': 'Authentication Type',
        
        # Session and conversation
        '会话列表': 'Session List',
        '新建会话': 'New Session',
        '历史会话': 'History Sessions',
        '当前会话': 'Current Session',
        '会话标题': 'Session Title',
        '重命名会话': 'Rename Session',
        '删除会话': 'Delete Session',
        '对话历史': 'Conversation History',
        '对话内容': 'Conversation Content',
        '清空对话': 'Clear Conversation',
        
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
        '知识库': 'Knowledge Base',
        '文档': 'Document',
        '片段': 'Segment',
        
        # Safety and moderation  
        '敏感词': 'Sensitive Word',
        '敏感词库': 'Sensitive Word Library',
        '敏感词数量': 'Sensitive Word Count',
        '拦截数量': 'Interception Count',
        '添加敏感词': 'Add Sensitive Word',
        '批量导入': 'Batch Import',
        '批量删除': 'Batch Delete',
        '安全护栏': 'Safety Guardrail',
        
        # Workflow
        '工作流': 'Workflow',
        '工作流名称': 'Workflow Name',
        '工作流描述': 'Workflow Description',
        '节点': 'Node',
        '节点名称': 'Node Name',
        '节点类型': 'Node Type',
        '运行历史': 'Run History',
        '运行详情': 'Run Details',
        '运行日志': 'Run Logs',
        '运行结果': 'Run Result',
        '运行状态': 'Run Status',
        
        # Common UI elements
        '确认': 'Confirm',
        '取消': 'Cancel',
        '保存': 'Save',
        '删除': 'Delete',
        '编辑': 'Edit',
        '新增': 'Add',
        '添加': 'Add',
        '搜索': 'Search',
        '查询': 'Query',
        '筛选': 'Filter',
        '排序': 'Sort',
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
        '详情': 'Details',
        '创建时间': 'Create Time',
        '更新时间': 'Update Time',
        '修改时间': 'Modify Time',
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
        '无数据': 'No data',
        '已复制': 'Copied',
        '复制成功': 'Copy successful',
        '复制失败': 'Copy failed',
        '已取消': 'Cancelled',
        '已删除': 'Deleted',
        '删除成功': 'Delete successful',
        '删除失败': 'Delete failed',
    }

    # Apply translations
    result = content
    for chinese, english in translations.items():
        result = result.replace(chinese, english)

    return result

def process_file(file_path):
    """Process and translate a single file"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            original_content = f.read()

        translated_content = translate_content(original_content)

        if translated_content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(translated_content)
            return True
        return False
    except Exception as e:
        print(f"Error processing {file_path}: {e}")
        return False

def main():
    """Main function to process all files"""
    print("Scanning for files with Chinese content...")

    # Frontend files
    frontend_files = get_files_with_chinese('web/src', ['.vue', '.js', '.html'])

    print(f"\nFound {len(frontend_files)} frontend files with Chinese content")
    print("\nTranslating files...\n")

    translated_count = 0
    for file_path in frontend_files:
        if process_file(file_path):
            translated_count += 1
            print(f"✓ {file_path}")

    print(f"\n{'='*60}")
    print(f"Translation complete!")
    print(f"Translated {translated_count} files")
    print(f"{'='*60}")

if __name__ == '__main__':
    main()


