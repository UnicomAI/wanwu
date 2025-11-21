import json
import re
import os

file_path = 'configs/microservice/bff-service/configs/wanwu-i18n.jsonl'

glossary = {
    '智能体': 'Agent',
    '知识库': 'Knowledge Base',
    '提示词': 'Prompt',
    '插件': 'Plugin',
    '组件': 'Component',
    '应用': 'App',
    '敏感词表': 'Sensitive Word Table',
    '敏感词': 'Sensitive Word',
    '客户端': 'Client',
    '统计': 'Statistic',
    '浏览量': 'Views',
    '用户量': 'Users',
    '文件': 'File',
    '配置': 'Configuration',
    '历史记录': 'History Record',
    '收藏': 'Favorite',
    '发布': 'Publish',
    '自定义': 'Custom',
    '工具': 'Tool',
    '模型': 'Model',
    '系统': 'System',
    '参数': 'Parameter',
    '类型': 'Type',
    '状态': 'Status',
    '列表': 'List',
    '详情': 'Details',
    '信息': 'Information',
    '失败': 'Failed',
    '成功': 'Success',
    '错误': 'Error',
    '不存在': 'Not Found',
    '已存在': 'Already Exists',
    '已禁用': 'Disabled',
    '已过期': 'Expired',
    '无效': 'Invalid',
    '为空': 'Is Empty',
    '不支持': 'Not Supported',
    '权限不足': 'Permission Denied',
    '内部错误': 'Internal Error',
    '未知错误': 'Unknown Error',
    '超时': 'Timeout',
    '连接失败': 'Connection Failed',
    '认证失败': 'Authentication Failed',
    '未登录': 'Not Logged In',
    '非法': 'Illegal',
    '格式错误': 'Format Error',
    '上传': 'Upload',
    '下载': 'Download',
    '导入': 'Import',
    '导出': 'Export',
    '生成': 'Generate',
    '解析': 'Parse',
    '处理': 'Process',
    '执行': 'Execute',
    '调用': 'Call',
    '访问': 'Access',
    '保存': 'Save',
    '复制': 'Copy',
    '重试': 'Retry',
    '请': 'Please',
    '稍后': 'Later',
    '修改': 'Modify',
    '删除': 'Delete',
    '创建': 'Create',
    '更新': 'Update',
    '获取': 'Get',
    '查询': 'Query',
    '新增': 'Add',
    '添加': 'Add',
    '取消': 'Cancel',
    '确认': 'Confirm',
    '通过': 'By',
    'ID': 'ID',
    '名称': 'Name',
    '描述': 'Description',
    '内容': 'Content',
    '数据': 'Data',
    '总量': 'Total',
    '数量': 'Count',
    '最大': 'Max',
    '容量': 'Capacity',
    '超出': 'Exceed',
    '同名': 'Same Name',
    '回复': 'Reply',
    '设置': 'Setting',
    '批量': 'Batch',
    '文档': 'Document',
    'mcp': 'MCP',
    'tool': 'Tool',
    'ApiKey': 'API Key',
    'URL': 'URL',
    'OAuth': 'OAuth',
    '密钥': 'Key',
    '平台': 'Platform',
    '活跃': 'Active',
    '累计': 'Cumulative',
    '新增': 'New',
}

patterns = [
    (r'创建(.+)错误: %v', r'Create \1 Error: %v'),
    (r'删除(.+)错误: %v', r'Delete \1 Error: %v'),
    (r'获取(.+)错误: %v', r'Get \1 Error: %v'),
    (r'更新(.+)错误: %v', r'Update \1 Error: %v'),
    (r'查询(.+)错误: %v', r'Query \1 Error: %v'),
    (r'新增(.+)错误: %v', r'Add \1 Error: %v'),
    (r'添加(.+)错误: %v', r'Add \1 Error: %v'),
    (r'(.+)已存在', r'\1 Already Exists'),
    (r'(.+)不存在', r'\1 Not Found'),
    (r'(.+)失败: %v', r'\1 Failed: %v'),
    (r'(.+)失败', r'\1 Failed'),
    (r'(.+)错误: %v', r'\1 Error: %v'),
    (r'(.+)错误', r'\1 Error'),
]

def translate_text(text):
    # 1. Apply patterns
    for zh_pattern, en_pattern in patterns:
        match = re.match(zh_pattern, text)
        if match:
            # Translate the captured group
            inner_text = match.group(1)
            translated_inner = translate_text(inner_text) # Recursive for nested terms
            # Replace the group in the pattern
            # We need to construct the result manually because re.sub with function is complex with groups
            # Simple approach: replace the group placeholder in en_pattern with translated_inner
            # But en_pattern has \1.
            return en_pattern.replace(r'\1', translated_inner)
            
    # 2. Apply glossary (longest match first)
    sorted_glossary = sorted(glossary.keys(), key=len, reverse=True)
    translated = text
    for key in sorted_glossary:
        if key in translated:
            translated = translated.replace(key, glossary[key] + " ") # Add space for separation
            
    # Clean up spaces
    translated = re.sub(r'\s+', ' ', translated).strip()
    # Fix some common grammar issues roughly
    translated = translated.replace(" Error", " Error").replace(" Failed", " Failed")
    
    return translated

updated_lines = []
with open(file_path, 'r', encoding='utf-8') as f:
    for line in f:
        if not line.strip():
            continue
        try:
            data = json.loads(line)
            langs = data.get('langs', {})
            zh = langs.get('zh', '')
            en = langs.get('en', '')
            
            if not en or en == zh:
                # Translate
                new_en = translate_text(zh)
                # Clean up % v -> %v
                new_en = new_en.replace('% v', '%v').replace('%V', '%v')
                langs['en'] = new_en
                data['langs'] = langs
                
            updated_lines.append(json.dumps(data, ensure_ascii=False))
        except json.JSONDecodeError:
            updated_lines.append(line.strip())

# Write back
with open(file_path, 'w', encoding='utf-8') as f:
    for line in updated_lines:
        f.write(line + '\n')

print(f"Updated {len(updated_lines)} lines in {file_path}")
