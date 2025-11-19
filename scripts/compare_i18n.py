import re
import os

def extract_keys(file_path):
    keys = set()
    stack = []
    
    with open(file_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()
        
    for line in lines:
        line = line.strip()
        # Remove comments
        if '//' in line:
            line = line.split('//')[0].strip()
            
        if not line:
            continue
            
        # Match key: {
        match_obj_start = re.match(r'([a-zA-Z0-9_]+):\s*{', line)
        if match_obj_start:
            key = match_obj_start.group(1)
            stack.append(key)
            current_path = '.'.join(stack)
            keys.add(current_path)
            continue
            
        # Match key: 'value' or key: "value" or key: value
        match_val = re.match(r'([a-zA-Z0-9_]+):\s*', line)
        if match_val and '{' not in line:
            key = match_val.group(1)
            current_path = '.'.join(stack + [key])
            keys.add(current_path)
            
        # Match closing brace
        if line.startswith('}'):
            if stack:
                stack.pop()
                
    return keys

zh_path = 'web/src/lang/zh.js'
en_path = 'web/src/lang/en.js'

zh_keys = extract_keys(zh_path)
en_keys = extract_keys(en_path)

missing_in_en = zh_keys - en_keys
missing_in_zh = en_keys - zh_keys

print("Missing in EN:")
for k in sorted(missing_in_en):
    print(k)
    
print("\nMissing in ZH:")
for k in sorted(missing_in_zh):
    print(k)
