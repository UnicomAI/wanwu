import os
import re

def is_comment_line(line, file_ext):
    line = line.strip()
    if file_ext in ['.go', '.js', '.ts', '.vue', '.java', '.c', '.cpp', '.proto']:
        return line.startswith('//') or line.startswith('/*') or line.endswith('*/') or '*' in line
    elif file_ext in ['.py', '.sh', '.yaml', '.yml', '.toml', '.ini']:
        return line.startswith('#')
    return False

def contains_chinese(text):
    return bool(re.search(r'[\u4e00-\u9fff]', text))

def scan_comments(root_dir):
    results = {}
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
                
                chinese_comments = []
                for i, line in enumerate(lines):
                    if contains_chinese(line):
                        # Simple heuristic: if it has Chinese, check if it looks like a comment
                        # OR if it's in a file type where we expect comments.
                        # Actually, we just want to translate ALL Chinese comments.
                        # But we need to distinguish code strings from comments if possible.
                        # For now, let's just list all lines with Chinese that look like comments.
                        if is_comment_line(line, file_ext) or '//' in line or '#' in line or '/*' in line:
                             chinese_comments.append((i+1, line.strip()))
                
                if chinese_comments:
                    results[file_path] = chinese_comments
            except Exception:
                pass
                
    return results

print("Scanning for Chinese comments...")
dirs_to_scan = ['internal', 'pkg', 'agent', 'rag', 'web/src', 'proto']
total_files = 0
total_lines = 0

for d in dirs_to_scan:
    if os.path.exists(d):
        print(f"Scanning {d}...")
        res = scan_comments(d)
        for f, lines in res.items():
            total_files += 1
            total_lines += len(lines)
            # print(f"{f}: {len(lines)} lines")

print(f"Found {total_files} files with {total_lines} lines of Chinese comments.")
