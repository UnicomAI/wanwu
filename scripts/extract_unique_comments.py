import os
import re
from collections import Counter

def is_comment_line(line, file_ext):
    line = line.strip()
    if file_ext in ['.go', '.js', '.ts', '.vue', '.java', '.c', '.cpp', '.proto']:
        return line.startswith('//') or line.startswith('/*') or line.endswith('*/') or '*' in line
    elif file_ext in ['.py', '.sh', '.yaml', '.yml', '.toml', '.ini']:
        return line.startswith('#')
    return False

def contains_chinese(text):
    return bool(re.search(r'[\u4e00-\u9fff]', text))

def extract_comments(root_dir):
    comments = []
    for root, dirs, files in os.walk(root_dir):
        if 'node_modules' in dirs:
            dirs.remove('node_modules')
        if '.git' in dirs:
            dirs.remove('.git')
            
        for file in files:
            file_ext = os.path.splitext(file)[1]
            if file_ext not in ['.go', '.py', '.js', '.vue', '.ts', '.proto']:
                continue
                
            file_path = os.path.join(root, file)
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    lines = f.readlines()
                
                for line in lines:
                    if contains_chinese(line):
                        clean_line = line.strip()
                        # Remove comment markers for better frequency analysis
                        clean_line = re.sub(r'^//\s*', '', clean_line)
                        clean_line = re.sub(r'^#\s*', '', clean_line)
                        clean_line = re.sub(r'^/\*\s*', '', clean_line)
                        clean_line = re.sub(r'\*/$', '', clean_line)
                        clean_line = clean_line.strip()
                        if clean_line:
                            comments.append(clean_line)
            except Exception:
                pass
    return comments

print("Extracting unique Chinese comments...")
dirs_to_scan = ['internal', 'pkg', 'agent', 'rag', 'web/src', 'proto']
all_comments = []
for d in dirs_to_scan:
    if os.path.exists(d):
        all_comments.extend(extract_comments(d))

counter = Counter(all_comments)
print(f"Found {len(all_comments)} total Chinese comment lines.")
print(f"Found {len(counter)} unique Chinese comment lines.")

with open('unique_comments.txt', 'w', encoding='utf-8') as f:
    for comment, count in counter.most_common():
        f.write(f"{count}\t{comment}\n")

print("Saved unique comments to unique_comments.txt")
