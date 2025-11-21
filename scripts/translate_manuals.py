import os
import re
import shutil

# Base directory
BASE_DIR = 'configs/microservice/bff-service/static/manual'

# Filename mapping (Chinese -> English)
FILENAME_MAP = {
    # Root
    '0.平台介绍.md': '0.Platform Introduction.md',
    '1.模型管理.md': '1.Model Management.md',
    '10.模板广场.md': '10.Template Square.md',
    '11.设置.md': '11.Settings.md',
    '2.知识库': '2.Knowledge Base',
    '3.资源库.md': '3.Resource Library.md',
    '4.安全护栏.md': '4.Safety Guardrails.md',
    '5.文本问答.md': '5.Text QA.md',
    '6.工作流': '6.Workflow',
    '7.智能体.md': '7.Agent.md',
    '8.应用广场.md': '8.App Square.md',
    '9.MCP广场.md': '9.MCP Square.md',
    '模型导入方式-详细版.md': 'Model Import Guide - Detailed.md',
    
    # 2.Knowledge Base
    '关键词管理.md': 'Keyword Management.md',
    '创建知识库': 'Create Knowledge Base',
    '命中测试': 'Hit Testing',
    '管理知识库': 'Manage Knowledge Base',
    
    # 6.Workflow
    'HTTP请求.md': 'HTTP Request.md',
    'JSON反序列化.md': 'JSON Deserialization.md',
    'JSON序列化.md': 'JSON Serialization.md',
    '代码.md': 'Code.md',
    '变量聚合.md': 'Variable Aggregation.md',
    '多文档解析.md': 'Multi-doc Parsing.md',
    '大模型.md': 'LLM.md',
    '工作流.md': 'Workflow.md',
    '工作流创建及发布.md': 'Workflow Creation and Publishing.md',
    '工具.md': 'Tools.md',
    '开始.md': 'Start.md',
    '循环.md': 'Loop.md',
    '意图识别.md': 'Intent Recognition.md',
    '批处理.md': 'Batch Processing.md',
    '文本处理.md': 'Text Processing.md',
    '文档生成.md': 'Document Generation.md',
    '文档解析.md': 'Document Parsing.md',
    '知识库.md': 'Knowledge Base.md',
    '结束.md': 'End.md',
    '输入.md': 'Input.md',
    '输出.md': 'Output.md',
    '选择器.md': 'Selector.md',
    
    # wanwu-openapi
    '1.文本问答API.md': '1.Text QA API.md',
    '2.智能体API.md': '2.Agent API.md',
    '3.工作流API.md': '3.Workflow API.md',
}

# Content translation glossary
GLOSSARY = {
    '平台介绍': 'Platform Introduction',
    '模型管理': 'Model Management',
    '模板广场': 'Template Square',
    '设置': 'Settings',
    '知识库': 'Knowledge Base',
    '资源库': 'Resource Library',
    '安全护栏': 'Safety Guardrails',
    '文本问答': 'Text QA',
    '工作流': 'Workflow',
    '智能体': 'Agent',
    '应用广场': 'App Square',
    'MCP广场': 'MCP Square',
    '模型导入方式': 'Model Import Guide',
    '关键词管理': 'Keyword Management',
    '创建知识库': 'Create Knowledge Base',
    '命中测试': 'Hit Testing',
    '管理知识库': 'Manage Knowledge Base',
    'HTTP请求': 'HTTP Request',
    'JSON反序列化': 'JSON Deserialization',
    'JSON序列化': 'JSON Serialization',
    '代码': 'Code',
    '变量聚合': 'Variable Aggregation',
    '多文档解析': 'Multi-doc Parsing',
    '大模型': 'LLM',
    '工作流创建及发布': 'Workflow Creation and Publishing',
    '工具': 'Tools',
    '开始': 'Start',
    '循环': 'Loop',
    '意图识别': 'Intent Recognition',
    '批处理': 'Batch Processing',
    '文本处理': 'Text Processing',
    '文档生成': 'Document Generation',
    '文档解析': 'Document Parsing',
    '结束': 'End',
    '输入': 'Input',
    '输出': 'Output',
    '选择器': 'Selector',
    '文本问答API': 'Text QA API',
    '智能体API': 'Agent API',
    '工作流API': 'Workflow API',
    '简介': 'Introduction',
    '功能': 'Features',
    '使用说明': 'Usage Instructions',
    '参数说明': 'Parameter Description',
    '返回结果': 'Return Result',
    '示例': 'Example',
    '注意': 'Note',
    '步骤': 'Steps',
    '配置': 'Configuration',
    '名称': 'Name',
    '描述': 'Description',
    '类型': 'Type',
    '必填': 'Required',
    '默认值': 'Default Value',
    '操作': 'Operation',
    '新增': 'Add',
    '编辑': 'Edit',
    '删除': 'Delete',
    '查询': 'Query',
    '详情': 'Details',
    '上传': 'Upload',
    '下载': 'Download',
    '导入': 'Import',
    '导出': 'Export',
    '成功': 'Success',
    '失败': 'Failed',
    '错误': 'Error',
    '状态': 'Status',
    '时间': 'Time',
    '用户': 'User',
    '角色': 'Role',
    '权限': 'Permission',
    '组织': 'Organization',
    '应用': 'App',
    '服务': 'Service',
    '接口': 'Interface',
    '请求方式': 'Request Method',
    '请求地址': 'Request URL',
    '请求头': 'Request Header',
    '请求体': 'Request Body',
    '响应体': 'Response Body',
    '错误码': 'Error Code',
    '错误信息': 'Error Message',
    '数据': 'Data',
    '列表': 'List',
    '文件': 'File',
    '图片': 'Image',
    '视频': 'Video',
    '音频': 'Audio',
    '链接': 'Link',
    '跳转': 'Jump',
    '刷新': 'Refresh',
    '重置': 'Reset',
    '提交': 'Submit',
    '取消': 'Cancel',
    '确认': 'Confirm',
    '保存': 'Save',
    '发布': 'Publish',
    '下架': 'Unpublish',
    '启用': 'Enable',
    '禁用': 'Disable',
    '是': 'Yes',
    '否': 'No',
    '无': 'None',
    '全部': 'All',
    '其他': 'Other',
}

def translate_content(content):
    # Sort glossary by length descending
    sorted_keys = sorted(GLOSSARY.keys(), key=len, reverse=True)
    
    # Translate
    for key in sorted_keys:
        if key in content:
            content = content.replace(key, GLOSSARY[key])
            
    return content

def update_links(content):
    # Update links to renamed files
    # Pattern: [text](link) or src="link"
    for cn_name, en_name in FILENAME_MAP.items():
        # URL encode Chinese characters for link matching might be needed, 
        # but let's try direct matching first as markdown often uses raw paths
        content = content.replace(f'({cn_name})', f'({en_name})')
        content = content.replace(f'"{cn_name}"', f'"{en_name}"')
        
        # Also handle relative paths if necessary (e.g. ./Chinese.md)
        content = content.replace(f'/{cn_name}', f'/{en_name}')
        
    return content

def process_directory(root_dir):
    # First rename files in this directory
    for filename in os.listdir(root_dir):
        if filename in FILENAME_MAP:
            old_path = os.path.join(root_dir, filename)
            new_path = os.path.join(root_dir, FILENAME_MAP[filename])
            os.rename(old_path, new_path)
            print(f"Renamed: {filename} -> {FILENAME_MAP[filename]}")

    # Then process files (translate content) and recurse into subdirectories
    # We need to list again because names changed
    for filename in os.listdir(root_dir):
        file_path = os.path.join(root_dir, filename)
        
        if os.path.isdir(file_path):
            process_directory(file_path)
        elif filename.endswith('.md'):
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            new_content = translate_content(content)
            new_content = update_links(new_content)
            
            if new_content != content:
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(new_content)
                print(f"Translated: {filename}")

print("Starting documentation translation...")
process_directory(BASE_DIR)
print("Finished.")
