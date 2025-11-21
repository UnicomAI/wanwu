import os

# Base directory
BASE_DIR = 'configs/microservice/bff-service/static/manual/2.Knowledge Base'

# Filename mapping (Chinese -> English)
FILENAME_MAP = {
    'Hit Testing/元数据过滤.md': 'Hit Testing/Metadata Filtering.md',
    'Hit Testing/创建命中测试.md': 'Hit Testing/Create Hit Test.md',
    'Hit Testing/检索方式配置.md': 'Hit Testing/Retrieval Config.md',
    'Manage Knowledge Base/知识库权限.md': 'Manage Knowledge Base/Knowledge Base Permissions.md',
    'Manage Knowledge Base/分段内容编辑.md': 'Manage Knowledge Base/Chunk Content Editing.md',
    'Manage Knowledge Base/文档处理状态查看.md': 'Manage Knowledge Base/Document Processing Status.md',
    'Manage Knowledge Base/查看分段结果.md': 'Manage Knowledge Base/View Chunk Results.md',
    'Create Knowledge Base/元数据.md': 'Create Knowledge Base/Metadata.md',
    'Create Knowledge Base/创建知识库.md': 'Create Knowledge Base/Create Knowledge Base.md',
    'Create Knowledge Base/分段配置.md': 'Create Knowledge Base/Chunk Config.md',
    'Create Knowledge Base/解析方式配置.md': 'Create Knowledge Base/Parsing Config.md',
    'Create Knowledge Base/知识图谱.md': 'Create Knowledge Base/Knowledge Graph.md',
}

print("Starting manual cleanup...")
for cn_path, en_path in FILENAME_MAP.items():
    full_cn_path = os.path.join(BASE_DIR, cn_path)
    full_en_path = os.path.join(BASE_DIR, en_path)
    
    if os.path.exists(full_cn_path):
        os.rename(full_cn_path, full_en_path)
        print(f"Renamed: {cn_path} -> {en_path}")
    else:
        print(f"File not found: {cn_path}")

print("Finished cleanup.")
