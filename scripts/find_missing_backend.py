import json

def find_missing_translations(file_path):
    missing = []
    with open(file_path, 'r', encoding='utf-8') as f:
        for line in f:
            if not line.strip():
                continue
            try:
                data = json.loads(line)
                langs = data.get('langs', {})
                zh = langs.get('zh', '')
                en = langs.get('en', '')
                
                if not en or en == zh: # Assuming if en is same as zh it might be untranslated, or just empty
                    # But sometimes they are same (e.g. numbers or simple words). 
                    # Let's focus on empty or missing 'en' key.
                    if 'en' not in langs or not langs['en']:
                        missing.append(data)
            except json.JSONDecodeError:
                pass
                
    return missing

missing_items = find_missing_translations('configs/microservice/bff-service/configs/wanwu-i18n.jsonl')
print(f"Found {len(missing_items)} missing translations.")
for item in missing_items:
    print(json.dumps(item, ensure_ascii=False))
