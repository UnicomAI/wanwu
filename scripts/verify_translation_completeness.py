import os
import re

def remove_comments(text):
    # Remove block comments /* ... */
    text = re.sub(r'/\*[\s\S]*?\*/', '', text)
    # Remove single line comments // ...
    text = re.sub(r'//.*', '', text)
    return text

def contains_chinese(text):
    return bool(re.search(r'[\u4e00-\u9fff]', text))

def scan_dir(root_dir):
    found_issues = []
    for root, dirs, files in os.walk(root_dir):
        if 'lang' in dirs:
            dirs.remove('lang') # Skip lang directory
        
        for file in files:
            if file.endswith(('.js', '.vue', '.html')):
                file_path = os.path.join(root, file)
                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        content = f.read()
                    
                    # Remove comments first to avoid false positives
                    clean_content = remove_comments(content)
                    
                    lines = clean_content.split('\n')
                    for i, line in enumerate(lines):
                        if contains_chinese(line):
                            # Double check it's not just whitespace or special chars
                            if line.strip():
                                found_issues.append(f"{file_path}:{i+1}: {line.strip()}")
                except Exception as e:
                    pass
    return found_issues

print("Scanning web/src for Chinese characters (ignoring comments)...")
issues = scan_dir('web/src')
if issues:
    print(f"Found {len(issues)} potential untranslated lines:")
    for issue in issues[:20]: # Show top 20
        print(issue)
else:
    print("No Chinese characters found in code (excluding comments).")
