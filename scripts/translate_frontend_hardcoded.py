#!/usr/bin/env python3
"""
Automated script to translate hardcoded Chinese strings in frontend files.
This script replaces Chinese text with English equivalents.
"""

import os
import re
import sys

# Translation mappings for common Chinese phrases
TRANSLATIONS = {
    # File upload related
    '从FileUpload': 'Upload from File',
    'urlFileUpload': 'Upload from URL',
    'url单条Upload': 'Upload Single URL',
    '将File拖到此处，OR': 'Drag file here, or',
    '点击Upload': 'click to upload',
    '模版Download': 'Download Template',
    '当前Content不自动Update': 'Content will not auto-update',
    
    # Segmentation related
    '分段Setting': 'Segmentation Settings',
    '分段标识': 'Segment Identifier',
    '分段标识Setting': 'Segment Identifier Settings',
    '可分割MaxValue': 'Max Segment Size',
    '文本预Process规Then': 'Text Preprocessing Rules',
    '替换掉连续 of 空格、换行符And制表符': 'Replace consecutive spaces, newlines and tabs',
    'Delete所有URLAnd电子邮件地址': 'Delete all URLs and email addresses',
    'Parse方式': 'Parsing Method',
    '元Data管理': 'Metadata Management',
    
    # Button text
    '上一步': 'Previous',
    '下一步': 'Next',
    '确 定': 'Confirm',
    '重 置': 'Reset',
    '确定': 'Confirm',
    'Cancel': 'Cancel',
    
    # Messages
    'Please enter有效范围内 of 数Value': 'Please enter a valid number within range',
    '数Value范围0-0.25': 'Value range: 0-0.25',
    'FileUpload中...': 'File uploading...',
    'Please上Enterurl!': 'Please enter URL!',
    '元Data管理存在未填写 of Required字段': 'Metadata management has unfilled required fields',
    
    # Search and placeholders
    'Search分隔符': 'Search separator',
    'Create分隔符': 'Create separator',
    'Delete分隔符': 'Delete separator',
    'Confirm要Delete当前分隔符？': 'Confirm deletion of current separator?',
    
    # File operations
    'Delete已UploadFile': 'Delete uploaded file',
    'Cancel所HasRequest': 'Cancel all requests',
    'Cancel所Has正在进行 of UploadRequest': 'Cancel all ongoing upload requests',
    'ResetUpload相关Status': 'Reset upload related status',
    '开始切片Upload(IfNoFile正在Upload)': 'Start chunk upload (if no file is uploading)',
    'IfUpload当中Has新 of File加入': 'If new file is added during upload',
    '重新UploadFile': 'Re-upload file',
    '续传File': 'Resume file upload',
    '重新Upload': 'Re-upload',
    '释放Memory': 'Release memory',
    
    # Error messages
    'FileDoes not existOR服务器Error': 'File does not exist or server error',
    'FileDownloadFailed，PleaseLaterRetry！': 'File download failed, please try again later!',
    
    # Comments
    'rag一体机No此逻辑': 'RAG all-in-one machine does not have this logic',
    '元Data管理Data': 'Metadata management data',
    '0Yes通用分段，1Yes父子分段': '0 is general segmentation, 1 is parent-child segmentation',
    '父子分段Required': 'Required for parent-child segmentation',
    
    # Warnings
    '子分段MaxValue已调整为': 'Child segment max size adjusted to',
    '不能超 父分段 of MaxValue': 'cannot exceed parent segment max size',
    '子分段MaxValue不能超 父分段 of MaxValue': 'Child segment max size cannot exceed parent segment max size',
    
    # Common UI elements
    '头部': 'Header',
    '表头Area': 'Table Header Area',
    '分页Area': 'Pagination Area',
    '已Copy到剪贴板': 'Copied to clipboard',
    'Copy failed，Please手动Copy': 'Copy failed, please copy manually',
    
    # Text-to-image
    '文生图Configuration示例': 'Text-to-Image Configuration Example',
    '打开文生图Configuration': 'Open Text-to-Image Configuration',
    '当前Configuration：': 'Current Configuration:',
    '文生图Configuration弹窗': 'Text-to-Image Configuration Dialog',
    '文生图Configuration已Save': 'Text-to-Image configuration saved',
    '弹窗关闭': 'Dialog closed',
    
    # Search config
    '检索方式Configuration': 'Search Method Configuration',
    '语义': 'Semantic',
    '关键词': 'Keyword',
    '最长上下文': 'Max Context Length',
    'Score阈Value': 'Score Threshold',
    
    # Metadata
    'Add条件': 'Add Condition',
    'Please select条件': 'Please select condition',
    '选择日期时间': 'Select date and time',
    '条件': 'Condition',
    '且': 'And',
    '或': 'Or',
    '早于': 'Before',
    '晚于': 'After',
    '不Is Empty': 'Not Empty',
    '不Yes': 'Not',
    '包含': 'Contains',
    '不包含': 'Does not contain',
    
    # Agent detail
    '返回': 'Back',
    '使用概述': 'Usage Overview',
    '特性说明': 'Feature Description',
    'App场景': 'App Scenarios',
    'WorkflowConfiguration说明': 'Workflow Configuration Description',
    '更多推荐': 'More Recommendations',
}

def contains_chinese(text):
    """Check if text contains Chinese characters"""
    return bool(re.search(r'[\u4e00-\u9fff]', text))

def translate_line(line):
    """Translate a single line by replacing Chinese text with English"""
    original_line = line
    
    # Sort by length (longest first) to avoid partial replacements
    for chinese, english in sorted(TRANSLATIONS.items(), key=lambda x: len(x[0]), reverse=True):
        if chinese in line:
            line = line.replace(chinese, english)
    
    return line

def process_file(file_path):
    """Process a single file and translate Chinese strings"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        lines = content.split('\n')
        translated_lines = []
        changes_made = False
        
        for line in lines:
            translated_line = translate_line(line)
            if translated_line != line:
                changes_made = True
            translated_lines.append(translated_line)
        
        if changes_made:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write('\n'.join(translated_lines))
            return True
        return False
    except Exception as e:
        print(f"Error processing {file_path}: {e}")
        return False

def main():
    """Main function to process all frontend files"""
    web_src_dir = 'web/src'
    files_processed = 0
    files_changed = 0
    
    for root, dirs, files in os.walk(web_src_dir):
        # Skip lang and node_modules directories
        if 'lang' in dirs:
            dirs.remove('lang')
        if 'node_modules' in dirs:
            dirs.remove('node_modules')
        
        for file in files:
            if file.endswith(('.vue', '.js', '.html')):
                file_path = os.path.join(root, file)
                files_processed += 1
                
                if process_file(file_path):
                    files_changed += 1
                    print(f"✓ Translated: {file_path}")
    
    print(f"\nProcessed {files_processed} files, changed {files_changed} files")

if __name__ == '__main__':
    main()

