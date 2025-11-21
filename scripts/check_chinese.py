import re

def check_chinese(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()
        
    for i, line in enumerate(lines):
        # Ignore comments
        if '//' in line:
            line = line.split('//')[0]
            
        # Check for Chinese characters range
        if re.search(r'[\u4e00-\u9fff]', line):
            print(f"Line {i+1}: {line.strip()}")

print("Checking en.js for Chinese characters...")
check_chinese('web/src/lang/en.js')
