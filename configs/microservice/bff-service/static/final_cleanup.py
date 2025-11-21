#!/usr/bin/env python3
"""
Final cleanup script to fix translation inconsistencies
"""

import re
import glob

def fix_common_issues(text):
    """
    Fix common translation issues
    """
    # Fix mixed Chinese/English in headers and content
    text = re.sub(r'CreateAgent"', r'Create Agent"', text)
    text = re.sub(r'AgentCreate"', r'Agent Create"', text)
    text = re.sub(r'AgentEdit"', r'Agent Edit"', text)
    text = re.sub(r'Text QA Create"', r'Text QA Create"', text)
    text = re.sub(r'Text QA Edit"', r'Text QA Edit"', text)
    text = re.sub(r'(.+?)(以下几Class)', r'\1 following categories:', text)
    text = re.sub(r'(User可自Row)', r'Users can custom', text)
    text = re.sub(r'(AgentIcon)', r'Agent Icon', text)
    text = re.sub(r'(AgentName)', r'Agent Name', text)
    text = re.sub(r'(AgentDescription)', r'Agent Description', text)
    text = re.sub(r'(SelectModelService)', r'Select Model Service', text)
    text = re.sub(r'(开场白)', r'Opening greeting', text)
    text = re.sub(r'(Used forEdit)', r'Used to edit', text)
    text = re.sub(r'(开场问候语)', r'opening greeting', text)
    text = re.sub(r'Perform AppFeatures0CodeDevelopment', r'perform application features and code development', text)
    
    # Fix number formatting (1、2、3、 etc.)
    text = re.sub(r'(\d+)、', r'\1. ', text)
    
    # Fix section headers with mixed languages
    text = re.sub(r'### (\d+)、(.+)', r'### \1. \2', text)
    text = re.sub(r'## (\d+)、(.+)', r'## \1. \2', text)
    
    # Fix inconsistent spacing
    text = re.sub(r'([a-zA-Z])([一-龯])', r'\1 \2', text)
    text = re.sub(r'([一-龯])([a-zA-Z])', r'\1 \2', text)
    
    # Fix bullet points that got corrupted
    text = re.sub(r'^\* (\*.*?:)', r'- \1', text, flags=re.MULTILINE)
    
    return text

def cleanup_file(file_path):
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_content = content
        cleaned_content = fix_common_issues(content)
        
        if cleaned_content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(cleaned_content)
            print(f"Cleaned: {file_path}")
            return True
        else:
            print(f"No changes needed: {file_path}")
            return False
            
    except Exception as e:
        print(f"Error cleaning {file_path}: {e}")
        return False

def main():
    manual_en_dir = "/Users/mohankumarv/Desktop/SAFVR/wanwu/configs/microservice/bff-service/static/Manual_en"
    
    # Find all markdown files
    md_files = glob.glob(f"{manual_en_dir}/**/*.md", recursive=True)
    
    print(f"Found {len(md_files)} markdown files to clean")
    
    cleaned_count = 0
    for md_file in md_files:
        if cleanup_file(md_file):
            cleaned_count += 1
    
    print(f"Final cleanup complete! Cleaned {cleaned_count} files.")

if __name__ == "__main__":
    main()
