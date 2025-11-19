import os
import re

# Final cleanup glossary
glossary = {
    'Please entervalue值': 'Please enter value',
    'Please开启元DataConfiguration后再进行Add': 'Please enable metadata configuration before adding',
    '存在未填Information,Please补充': 'Missing information, please complete',
    '文生图名字': 'Text-to-Image Name',
    'MaaS-简要Description': 'MaaS-Brief Description',
    '网页链接': 'Web Link',
    '字符串': 'String',
    '用于链接到网页': 'Used to link to webpage',
    '若没有Add过API Key,则显示输入框;若Add过,直接展示....': 'If API Key not added, show input box; if added, show directly...',
    'Please先输入API Key': 'Please enter API Key first',
    '暂无Download路径': 'No download path available',
    'Search区': 'Search Area',
    '表格区': 'Table Area',
    '多选区': 'Multi-select Area',
    '索引区': 'Index Area',
    '标题和Description': 'Title and Description',
    'API Key 部分': 'API Key Section',
    'Parameter表格': 'Parameter Table',
    '底部Confirm按钮': 'Bottom Confirm Button',
    'Please求Failed时的ErrorInformation': 'Error information when request failed',
    'Please求': 'Request',
    'Failed时': 'When Failed',
    '的ErrorInformation': ' Error Information',
    'Add条件': 'Add Condition',
    'Please select条件': 'Please select condition',
    '选择日期时间': 'Select Date Time',
    '条件': 'Condition',
    '且': 'AND',
    '或': 'OR',
    '早于': 'Earlier than',
    '晚于': 'Later than',
    '不Is Empty': 'Is Not Empty',
    '不Yes': 'Is Not',
    '包含': 'Contains',
    '不包含': 'Does not contain',
    '开始Yes': 'Starts With',
    '结束Yes': 'Ends With',
    '不等于': 'Not Equal',
    '大于等于': 'Greater Than or Equal',
    '小于等于': 'Less Than or Equal',
    '大于': 'Greater Than',
    '小于': 'Less Than',
    '等于': 'Equal',
    '是': 'Yes',
    '否': 'No',
    '的': ' of ',
    '时': ' when ',
    '求': 'Request',
    '值': 'Value',
    '名': 'Name',
    '名字': 'Name',
    '简要': 'Brief',
    '链接': 'Link',
    '用于': 'Used for',
    '若': 'If',
    '没有': 'No',
    '过': ' ',
    '则': 'Then',
    '显示': 'Show',
    '输入框': 'Input Box',
    '直接': 'Directly',
    '展示': 'Show',
    '先': 'First',
    '输入': 'Enter',
    '暂无': 'No',
    '路径': 'Path',
    '区': 'Area',
    '部分': 'Section',
    '表格': 'Table',
    '底部': 'Bottom',
    '按钮': 'Button',
    '标题': 'Title',
    '和': 'And',
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

print("Starting v4 automated translation of hardcoded strings...")
fixed_count = scan_and_fix('web/src')
print(f"Finished. Modified {fixed_count} files.")
