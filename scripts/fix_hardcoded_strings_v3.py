import os
import re

# Final glossary for remaining strings and cleanup
glossary = {
    'Please求Failed时的ErrorInformation': 'Error information when request failed',
    'Please求': 'Request',
    'Failed时': 'When Failed',
    '的ErrorInformation': ' Error Information',
    '返回': 'Back',
    '使用概述': 'Usage Overview',
    '特性说明': 'Feature Description',
    'App场景': 'App Scenario',
    'WorkflowConfiguration说明': 'Workflow Configuration Description',
    '更多推荐': 'More Recommendations',
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
    '是': 'Yes', # Re-add simple ones just in case
    '否': 'No',
    '的': ' of ', # Risky but might be needed for some
    '时': ' when ',
    '求': 'Request', # Last resort for "Please求"
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

print("Starting v3 automated translation of hardcoded strings...")
fixed_count = scan_and_fix('web/src')
print(f"Finished. Modified {fixed_count} files.")
