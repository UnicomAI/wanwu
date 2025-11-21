#!/usr/bin/env python3
"""
Final cleanup script for remaining Chinese strings in frontend.
Handles edge cases and mixed Chinese-English text.
"""

import os
import re

# Final set of translation patterns for remaining strings
FINAL_TRANSLATIONS = [
    # Search config descriptions
    (r'重SortModel会根据候选Document与用户问题 of Semantic匹配度，对初步检索结果进行重新Sort从而进一步提升最终Back结果 of 相关性And准确性。',
     'The reranking model reorders initial search results based on semantic matching between candidate documents and user questions, further improving the relevance and accuracy of final returned results.'),
    (r'By向量Similarity找到Semantic相近、表达多样 of 文本Segment，适Used for理解And召回Semantic相关Information。',
     'Find semantically similar and diversely expressed text segments through vector similarity, suitable for understanding and recalling semantically related information.'),
    (r'全文检索', 'Full-text Search'),
    (r'基于Keyword匹配，能够高效QueryContains指定词汇 of 文本Segment，适Used for精确查找',
     'Based on keyword matching, efficiently query text segments containing specified words, suitable for precise search'),
    (r'混合检索', 'Hybrid Search'),
    (r'结合向量AndKeyword检索，融合Semantic理解与Keyword匹配，兼顾相关性And准确性，提升检索效果。',
     'Combines vector and keyword search, integrating semantic understanding with keyword matching, balancing relevance and accuracy to improve search results.'),
    (r'权重Setting', 'Weight Settings'),
    (r'权重Setting功能Used for调整不同检索方式 of 影响力。BySetting权重，可以控制SemanticSimilarityAndKeyword匹配在最终Sort中 of 占比。',
     'Weight settings function is used to adjust the influence of different search methods. By setting weights, you can control the proportion of semantic similarity and keyword matching in the final ranking.'),
    
    # Common mixed patterns
    (r'Add标志位，Used forArea分YesNoYes从configSetting of Value', 'Add flag to distinguish whether value is set from config'),
    (r'最长Context', 'Max context'),
    (r'IfYes从configSetting of Value，不TriggersendConfigInfo', 'If value is set from config, do not trigger sendConfigInfo'),
    (r'Setting标志位', 'Set flag'),
    (r'UsenextTick确保DOMUpdate完成后再Reset标志位', 'Use nextTick to ensure DOM update completes before resetting flag'),
    (r'预LoadData，避免首次打开下拉框 when  of Delay', 'Preload data to avoid delay when opening dropdown for the first time'),
    (r'DirectlyTriggerEvent，避免防抖Delay', 'Directly trigger event to avoid debounce delay'),
    
    # Image and file upload
    (r'图片Upload中', 'Image uploading'),
    (r'click to upload图片', 'Click to upload image'),
    (r'Please enter文本Description', 'Please enter text description'),
    (r'文本Description限制600字符以内', 'Text description limited to 600 characters'),
    (r'渲染完成', 'Rendering complete'),
    (r'渲染Failed', 'Rendering failed'),
    (r'预览Failed', 'Preview failed'),
    
    # App list and publish
    (r'私密', 'Private'),
    (r'公开', 'Public'),
    (r'CancelPublish后，历史引用了本Workflow of Agent将自动Cancel引用，AND此Operation不可撤回',
     'After unpublishing, agents that historically referenced this workflow will automatically cancel the reference, and this operation cannot be undone'),
    
    # URL and API creation
    (r'URL地址', 'URL Address'),
    (r'API地址', 'API Address'),
    (r'请求方式', 'Request Method'),
    (r'请求头', 'Request Headers'),
    (r'请求体', 'Request Body'),
    (r'响应', 'Response'),
    (r'超时设置', 'Timeout Settings'),
    (r'重试次数', 'Retry Count'),
    
    # Section and knowledge
    (r'片段编号', 'Segment Number'),
    (r'片段长度', 'Segment Length'),
    (r'原始文本', 'Original Text'),
    (r'向量化', 'Vectorization'),
    (r'索引状态', 'Index Status'),
    (r'已索引', 'Indexed'),
    (r'未索引', 'Not Indexed'),
    (r'索引中', 'Indexing'),
    
    # Model access
    (r'模型类别', 'Model Category'),
    (r'对话模型', 'Chat Model'),
    (r'嵌入模型', 'Embedding Model'),
    (r'重排序模型', 'Rerank Model'),
    (r'图像模型', 'Image Model'),
    (r'语音模型', 'Speech Model'),
    
    # Safety and moderation
    (r'敏感词类型', 'Sensitive Word Type'),
    (r'拦截级别', 'Interception Level'),
    (r'高风险', 'High Risk'),
    (r'中风险', 'Medium Risk'),
    (r'低风险', 'Low Risk'),
    (r'白名单', 'Whitelist'),
    (r'黑名单', 'Blacklist'),
    
    # Workflow
    (r'节点配置', 'Node Configuration'),
    (r'输入参数', 'Input Parameters'),
    (r'输出参数', 'Output Parameters'),
    (r'执行条件', 'Execution Condition'),
    (r'错误处理', 'Error Handling'),
    (r'重试策略', 'Retry Strategy'),
    
    # Session and chat
    (r'会话列表', 'Session List'),
    (r'会话详情', 'Session Details'),
    (r'消息记录', 'Message History'),
    (r'发送消息', 'Send Message'),
    (r'清空会话', 'Clear Session'),
    
    # Common UI text
    (r'操作成功', 'Operation successful'),
    (r'操作失败', 'Operation failed'),
    (r'加载中', 'Loading'),
    (r'加载失败', 'Load failed'),
    (r'暂无数据', 'No data'),
    (r'请选择', 'Please select'),
    (r'请输入', 'Please enter'),
    (r'必填项', 'Required field'),
    (r'选填项', 'Optional field'),

    # URL creation and expiry
    (r' 期 when 间', 'Expiry Time'),
    (r'url生效 when 间', 'URL effective time'),
    (r'版权', 'Copyright'),
    (r'YesNo在界面中Show版权Information', 'Whether to show copyright information in the interface'),
    (r'Please enter版权Information', 'Please enter copyright information'),
    (r'隐私协议', 'Privacy Policy'),
    (r'YesNo在界面中Show隐私协议Information', 'Whether to show privacy policy information in the interface'),
    (r'Please enter隐私政策Link', 'Please enter privacy policy link'),
    (r'免责声明', 'Disclaimer'),
    (r'YesNo在界面中Show免责声明', 'Whether to show disclaimer in the interface'),
    (r'Please enter免责声明', 'Please enter disclaimer'),
    (r'Link效验不合格', 'Link validation failed'),
    (r'Content已Copy到粘贴板', 'Content copied to clipboard'),
    (r'Confirm要Delete当前AccessURL吗？', 'Confirm deletion of current access URL?'),

    # Section and metadata
    (r'元Data', 'Metadata'),
    (r'无Data', 'No data'),
    (r'元Data规Then', 'Metadata Rules'),
    (r'BatchAdd分段Status', 'Batch Add Segment Status'),
    (r'Add分段', 'Add Segment'),
    (r'父子分段', 'Parent-child Segment'),
    (r'通用分段', 'General Segment'),
    (r'个子分段', ' child segments'),
    (r'Save并重新Parse子分段', 'Save and reparse child segments'),
    (r'Add子分段', 'Add Child Segment'),
    (r'Confirm要Delete这个子分段吗？', 'Confirm deletion of this child segment?'),
    (r'无Modify', 'No changes'),

    # Model access constants
    (r'DocumentParse服务', 'Document Parsing Service'),
    (r'联通元景', 'China Unicom Yuanjing'),
    (r'通义千问', 'Tongyi Qianwen'),
    (r'火山引擎', 'Volcano Engine'),
    (r'无问芯穹', 'Infini'),
    (r'支持', 'Supported'),

    # Session component
    (r'对话History', 'Conversation History'),
    (r'新建对话', 'New Conversation'),
    (r'History对话', 'History Conversations'),
    (r'当前对话', 'Current Conversation'),
    (r'对话标题', 'Conversation Title'),
    (r'重命名对话', 'Rename Conversation'),
    (r'Delete对话', 'Delete Conversation'),
    (r'Confirm要Delete这条对话吗？', 'Confirm deletion of this conversation?'),

    # Action and tools
    (r'工具名称', 'Tool Name'),
    (r'工具Description', 'Tool Description'),
    (r'工具类型', 'Tool Type'),
    (r'工具参数', 'Tool Parameters'),
    (r'参数名称', 'Parameter Name'),
    (r'参数类型', 'Parameter Type'),
    (r'参数Description', 'Parameter Description'),
    (r'YesNo必填', 'Required'),
    (r'默认Value', 'Default Value'),

    # Safety and moderation
    (r'敏感词库名称', 'Sensitive Word Library Name'),
    (r'敏感词Count', 'Sensitive Word Count'),
    (r'拦截Count', 'Interception Count'),
    (r'最近Update when 间', 'Last Update Time'),
    (r'Add敏感词', 'Add Sensitive Word'),
    (r'BatchImport', 'Batch Import'),
    (r'BatchDelete', 'Batch Delete'),
    (r'Confirm要Delete选中 of 敏感词吗？', 'Confirm deletion of selected sensitive words?'),

    # Workflow run
    (r'运行History', 'Run History'),
    (r'运行Details', 'Run Details'),
    (r'运行日志', 'Run Logs'),
    (r'运行结果', 'Run Result'),
    (r'运行 when 长', 'Run Duration'),
    (r'开始 when 间', 'Start Time'),
    (r'结束 when 间', 'End Time'),
    (r'节点执行Details', 'Node Execution Details'),
    (r'Input参数', 'Input Parameters'),
    (r'Output结果', 'Output Result'),

    # Knowledge base
    (r'Knowledge Base名称', 'Knowledge Base Name'),
    (r'Knowledge BaseDescription', 'Knowledge Base Description'),
    (r'DocumentCount', 'Document Count'),
    (r'片段Count', 'Segment Count'),
    (r'向量化Status', 'Vectorization Status'),
    (r'已向量化', 'Vectorized'),
    (r'未向量化', 'Not Vectorized'),
    (r'向量化中', 'Vectorizing'),
    (r'向量化Failed', 'Vectorization Failed'),

    # Keyword and tags
    (r'关键词', 'Keyword'),
    (r'关键词列表', 'Keyword List'),
    (r'Add关键词', 'Add Keyword'),
    (r'Delete关键词', 'Delete Keyword'),
    (r'标签', 'Tag'),
    (r'标签列表', 'Tag List'),
    (r'Add标签', 'Add Tag'),
    (r'Delete标签', 'Delete Tag'),
    (r'Confirm要Delete这个标签吗？', 'Confirm deletion of this tag?'),
]

def translate_content(content):
    """Translate content using pattern matching"""
    for pattern, replacement in FINAL_TRANSLATIONS:
        content = re.sub(pattern, replacement, content)
    return content

def process_file(file_path):
    """Process a single file"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            original_content = f.read()
        
        translated_content = translate_content(original_content)
        
        if translated_content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(translated_content)
            return True
        return False
    except Exception as e:
        print(f"Error processing {file_path}: {e}")
        return False

def main():
    web_src_dir = 'web/src'
    files_changed = 0
    
    for root, dirs, files in os.walk(web_src_dir):
        if 'lang' in dirs:
            dirs.remove('lang')
        if 'node_modules' in dirs:
            dirs.remove('node_modules')
        
        for file in files:
            if file.endswith(('.vue', '.js', '.html')):
                file_path = os.path.join(root, file)
                if process_file(file_path):
                    files_changed += 1
                    print(f"✓ {file_path}")
    
    print(f"\nTranslated {files_changed} files")

if __name__ == '__main__':
    main()

