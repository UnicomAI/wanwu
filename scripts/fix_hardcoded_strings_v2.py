import os
import re

# Additional glossary for remaining strings
glossary = {
    '本次回答已被终止': 'This response has been terminated',
    '必填': 'Required',
    'mcp-工具名': 'mcp-tool name',
    '工作流': 'Workflow',
    '文本问答': 'Text Q&A',
    '智能体': 'Agent',
    '智能体模版': 'Agent Template',
    '涉政': 'Political',
    '辱骂': 'Abusive',
    '涉黄': 'Pornographic',
    '暴恐': 'Violent/Terrorist',
    '违禁': 'Prohibited',
    '信息安全': 'Information Security',
    '其他': 'Other',
    '心知天气API': 'Seniverse Weather API',
    '提供当前天气信息的API，包括温度、天气状况等。': 'API providing current weather info, including temperature, conditions, etc.',
    '天气Query工具': 'Weather Query Tool',
    '根据地点获取当前的天气情况，包括温度和天气状况Description。': 'Get current weather for a location, including temperature and conditions.',
    'Query的地点，可以Yes城市名、邮编等。': 'Query location, can be city name, zip code, etc.',
    'Success获取天气信息': 'Successfully retrieved weather info',
    '知识库': 'Knowledge Base',
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
}

def replace_chinese(text):
    sorted_keys = sorted(glossary.keys(), key=len, reverse=True)
    for key in sorted_keys:
        if key in text:
            text = text.replace(key, glossary[key])
    return text

def process_file(file_path):
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_content = content
        new_content = replace_chinese(content)
        
        if new_content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(new_content)
            return True
    except Exception as e:
        print(f"Error processing {file_path}: {e}")
    return False

def scan_and_fix(root_dir):
    count = 0
    for root, dirs, files in os.walk(root_dir):
        if 'lang' in dirs:
            dirs.remove('lang')
            
        for file in files:
            if file.endswith(('.js', '.vue', '.html')):
                file_path = os.path.join(root, file)
                if process_file(file_path):
                    print(f"Fixed: {file_path}")
                    count += 1
    return count

print("Starting automated translation of remaining hardcoded strings...")
fixed_count = scan_and_fix('web/src')
print(f"Finished. Modified {fixed_count} files.")
