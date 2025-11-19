import os

def is_ascii(s):
    return all(ord(c) < 128 for c in s)

def scan_dir(root_dir):
    for root, dirs, files in os.walk(root_dir):
        if 'lang' in dirs:
            dirs.remove('lang') # Skip lang directory
        
        for file in files:
            if file.endswith(('.js', '.vue', '.html')):
                file_path = os.path.join(root, file)
                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        lines = f.readlines()
                    
                    for i, line in enumerate(lines):
                        line = line.strip()
                        if '//' in line:
                            line = line.split('//')[0]
                        if not is_ascii(line):
                            print(f"{file_path}:{i+1}: {line.strip()}")
                except Exception as e:
                    pass

print("Scanning web/src for non-ASCII characters...")
scan_dir('web/src')
