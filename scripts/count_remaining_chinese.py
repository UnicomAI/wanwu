#!/usr/bin/env python3
import re
import os

def contains_chinese(text):
    return bool(re.search(r'[\u4e00-\u9fff]', text))

def remove_comments(content):
    content = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)
    content = re.sub(r'/\*.*?\*/', '', content, flags=re.DOTALL)
    content = re.sub(r'<!--.*?-->', '', content, flags=re.DOTALL)
    return content

file_counts = {}
total_lines = 0
for root, dirs, files in os.walk('web/src'):
    if 'lang' in dirs:
        dirs.remove('lang')
    if 'node_modules' in dirs:
        dirs.remove('node_modules')
    
    for file in files:
        if file.endswith(('.js', '.vue', '.html')):
            file_path = os.path.join(root, file)
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                clean_content = remove_comments(content)
                lines = clean_content.split('\n')
                count = sum(1 for line in lines if contains_chinese(line) and line.strip())
                if count > 0:
                    file_counts[file_path] = count
                    total_lines += count
            except:
                pass

print(f'Total files with Chinese: {len(file_counts)}')
print(f'Total lines with Chinese: {total_lines}')
print(f'\nRemaining files (top 30):')
for file_path, count in sorted(file_counts.items(), key=lambda x: x[1], reverse=True)[:30]:
    print(f'{count:4d} {file_path}')

