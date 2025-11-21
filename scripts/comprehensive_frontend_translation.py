#!/usr/bin/env python3
"""
Comprehensive frontend translation script.
Handles mixed Chinese-English text and provides detailed translations.
"""

import os
import re
import json

# Comprehensive translation patterns (regex patterns and replacements)
PATTERN_TRANSLATIONS = [
    # File upload patterns
    (r'从FileUpload', 'Upload from File'),
    (r'urlFileUpload', 'Upload from URL'),  
    (r'url单条Upload', 'Upload Single URL'),
    (r'将File拖到此处，OR', 'Drag file here, or'),
    (r'点击Upload', 'click to upload'),
    (r'模版Download', 'Download Template'),
    (r'当前Content不自动Update', 'Content will not auto-update'),
    (r'您可单独OR者BatchUpload以下格式 of Document：', 'You can upload the following document formats individually or in batch: '),
    (r'FileMax为', 'Maximum file size: '),
    (r'File格式Upload大小限制', 'file format upload size limit'),
    (r'非压缩包File，一次可传', 'For non-compressed files, you can upload '),
    (r'个File，如File页数多，DocumentParse when 间较长，平均', ' files at a time. If the file has many pages, document parsing takes longer, averaging '),
    (r'秒/页，Please您耐心等待', ' seconds/page, please be patient'),
    (r'BatchUpload支持', 'Batch upload supports '),
    (r'格式，仅可Upload', ' format, can only upload '),
    (r'个。Document最多可Add', ' file. Document can add up to '),
    (r'条url，File不超 ', ' URLs, file size not exceeding '),
    
    # Segmentation
    (r'分段Setting', 'Segmentation Settings'),
    (r'分段标识Setting', 'Segment Identifier Settings'),
    (r'分段标识', 'Segment Identifier'),
    (r'可分割MaxValue', 'Max Segment Size'),
    (r'文本预Process规Then', 'Text Preprocessing Rules'),
    (r'替换掉连续 of 空格、换行符And制表符', 'Replace consecutive spaces, newlines and tabs'),
    (r'Delete所有URLAnd电子邮件地址', 'Delete all URLs and email addresses'),
    (r'Parse方式', 'Parsing Method'),
    (r'元Data管理', 'Metadata Management'),
    
    # Buttons
    (r'上一步', 'Previous'),
    (r'下一步', 'Next'),
    (r'确 定', 'Confirm'),
    (r'重 置', 'Reset'),
    
    # Messages
    (r'Please enter有效范围内 of 数Value', 'Please enter a valid number within range'),
    (r'数Value范围', 'Value range: '),
    (r'FileUpload中\.\.\.', 'File uploading...'),
    (r'Please上Enterurl!', 'Please enter URL!'),
    (r'元Data管理存在未填写 of Required字段', 'Metadata management has unfilled required fields'),
    
    # Search
    (r'Search分隔符', 'Search separator'),
    (r'Create分隔符', 'Create separator'),
    (r'Delete分隔符', 'Delete separator'),
    (r'Confirm要Delete当前分隔符？', 'Confirm deletion of current separator?'),
    
    # File operations
    (r'Delete已UploadFile', 'Delete uploaded file'),
    (r'Cancel所HasRequest', 'Cancel all requests'),
    (r'Cancel所Has正在进行 of UploadRequest', 'Cancel all ongoing upload requests'),
    (r'ResetUpload相关Status', 'Reset upload related status'),
    (r'开始切片Upload\(IfNoFile正在Upload\)', 'Start chunk upload (if no file is uploading)'),
    (r'IfUpload当中Has新 of File加入', 'If new file is added during upload'),
    (r'重新UploadFile', 'Re-upload file'),
    (r'续传File', 'Resume file upload'),
    (r'重新Upload', 'Re-upload'),
    (r'释放Memory', 'Release memory'),
    
    # Errors
    (r'FileDoes not existOR服务器Error', 'File does not exist or server error'),
    (r'FileDownloadFailed，PleaseLaterRetry！', 'File download failed, please try again later!'),
    
    # Comments
    (r'rag一体机No此逻辑', 'RAG all-in-one machine does not have this logic'),
    (r'元Data管理Data', 'Metadata management data'),
    (r'0Yes通用分段，1Yes父子分段', '0 is general segmentation, 1 is parent-child segmentation'),
    (r'父子分段Required', 'Required for parent-child segmentation'),
    
    # Warnings
    (r'子分段MaxValue已调整为', 'Child segment max size adjusted to'),
    (r'不能超 父分段 of MaxValue', 'cannot exceed parent segment max size'),
    (r'子分段MaxValue不能超 父分段 of MaxValue', 'Child segment max size cannot exceed parent segment max size'),
    
    # UI elements
    (r'头部', 'Header'),
    (r'表头Area', 'Table Header Area'),
    (r'分页Area', 'Pagination Area'),
    (r'已Copy到剪贴板', 'Copied to clipboard'),
    (r'Copy failed，Please手动Copy', 'Copy failed, please copy manually'),
    
    # Text-to-image
    (r'文生图Configuration示例', 'Text-to-Image Configuration Example'),
    (r'打开文生图Configuration', 'Open Text-to-Image Configuration'),
    (r'当前Configuration：', 'Current Configuration:'),
    (r'文生图Configuration弹窗', 'Text-to-Image Configuration Dialog'),
    (r'文生图Configuration已Save', 'Text-to-Image configuration saved'),
    (r'弹窗关闭', 'Dialog closed'),
    
    # Search config
    (r'检索方式Configuration', 'Search Method Configuration'),
    (r'语义', 'Semantic'),
    (r'关键词', 'Keyword'),
    (r'最长上下文', 'Max Context Length'),
    (r'Score阈Value', 'Score Threshold'),
    (r'重排序Model会根据候选Document与用户问题 of 语义匹配度，对初步检索结果进行重新排序从而进一步提升最终Back结果 of 相关性And准确性。', 
     'The reranking model will reorder the initial search results based on the semantic matching degree between candidate documents and user questions, thereby further improving the relevance and accuracy of the final returned results.'),
    (r'Used for控制检索阶段Back of 最相关 of Document片段 of Count。这些Document片段将被送入GenerateModel中，Used for Generate最终 of 回答。',
     'Used to control the number of most relevant document segments returned in the retrieval phase. These document segments will be sent to the generation model to generate the final answer.'),
    (r'Save of 最长 of 上下文对话轮数。', 'Maximum number of context dialogue rounds to save.'),
    (r'检索结果 of 相似度阈Value，低于该Value of 结果将被 滤。', 'Similarity threshold for search results, results below this value will be filtered.'),
    
    # Metadata conditions
    (r'Add条件', 'Add Condition'),
    (r'Please select条件', 'Please select condition'),
    (r'选择日期时间', 'Select date and time'),
    (r'条件', 'Condition'),
    (r'且', 'And'),
    (r'或', 'Or'),
    (r'早于', 'Before'),
    (r'晚于', 'After'),
    (r'不Is Empty', 'Not Empty'),
    (r'不Yes', 'Not'),
    (r'包含', 'Contains'),
    (r'不包含', 'Does not contain'),
    
    # Agent detail
    (r'返回', 'Back'),
    (r'使用概述', 'Usage Overview'),
    (r'特性说明', 'Feature Description'),
    (r'App场景', 'App Scenarios'),
    (r'WorkflowConfiguration说明', 'Workflow Configuration Description'),
    (r'更多推荐', 'More Recommendations'),

    # Form and UI
    (r'无Information', 'No information'),
    (r'API根地址', 'API Base URL'),
    (r'API密钥', 'API Key'),
    (r'私密Publish为App：仅自己可见', 'Private: Visible only to you'),
    (r'公开Publish为App：组织内可见', 'Public: Visible within organization'),
    (r'公开Publish为App：全局可见', 'Public: Globally visible'),
    (r'保 存', 'Save'),
    (r'Model选择', 'Model Selection'),
    (r'暂Not Supported选择图文问答类Model', 'Image-text Q&A models not currently supported'),
    (r'可EnterModelNameSearch', 'Enter model name to search'),
    (r'关联Knowledge Base', 'Associated Knowledge Base'),
    (r'元Data 滤', 'Metadata Filter'),
    (r'安全护栏Configuration', 'Safety Guardrail Configuration'),
    (r'实 when 拦截高风险Content of EnterAnd输出，保障Content安全合规。', 'Real-time interception of high-risk input and output content to ensure content safety and compliance.'),
    (r'知识图谱', 'Knowledge Graph'),
    (r'Knowledge Base选择', 'Knowledge Base Selection'),
    (r'元DataSetting', 'Metadata Settings'),
    (r'Configuration元Data 滤', 'Configure Metadata Filter'),
    (r'BySetting of 元Data，对Knowledge Base内Information进行更加细化 of 筛选与检索控制。', 'By setting metadata, perform more refined filtering and retrieval control on information within the knowledge base.'),
    (r'取 消', 'Cancel'),

    # Model and knowledge config
    (r'关Key词权重', 'Keyword weight'),
    (r'向量检索', 'Vector search'),
    (r'Text检索', 'Text search'),
    (r'混合检索：向量\+Text', 'Hybrid search: vector + text'),
    (r'权重Match，只Has在混合检索模式下，选择权重Setting后，这个才Setting为', 'Weight matching, only set to '),
    (r'Semantic权重', 'Semantic weight'),
    (r'Get最高 of 几行', 'Get top N rows'),
    (r' 滤分数阈Value', 'Filter score threshold'),

    # Comments and labels
    (r'防抖计 when 器', 'Debounce timer'),
    (r'防止重复Update标记', 'Prevent duplicate update flag'),
    (r'防止DetailsDataTriggerUpdate标记', 'Prevent detail data trigger update flag'),
    (r'IfYes从DetailsSetting of Data，不TriggerUpdate逻辑', 'If data is set from details, do not trigger update logic'),
    (r'GetDetails', 'Get details'),
    (r'Getapi跟Address', 'Get API base address'),
    (r'判断YesNoPublish', 'Check if published'),
    (r'Setting标志位，防止TriggerUpdate逻辑', 'Set flag to prevent triggering update logic'),
    (r'UsenextTick确保所HasDataSetting完成后再Reset标志位', 'Use nextTick to ensure all data is set before resetting flag'),
    (r'Please选rerank择Model！', 'Please select rerank model!'),
    (r'Please select关联Knowledge Base！', 'Please select associated knowledge base!'),
    (r'防止重复Call', 'Prevent duplicate calls'),
    (r'UpdateSuccess后，Update initialEditForm 避免重复Trigger', 'After successful update, update initialEditForm to avoid duplicate triggers'),

    # CSS comments
    (r'/\*通用\*/', '/* General */'),
    (r'/\*新建App\*/', '/* Create App */'),
    (r'/\*推荐问题\*/', '/* Recommended Questions */'),
    (r'/\*知识增强\*/', '/* Knowledge Enhancement */'),
    (r'Setting边框颜色', 'Set border color'),
    (r'Setting背景颜色', 'Set background color'),
    (r'Setting文字颜色', 'Set text color'),

    # Permission and user management
    (r'成员', 'Member'),
    (r'组织', 'Organization'),
    (r'权限', 'Permission'),
    (r'可读', 'Read'),
    (r'可Edit', 'Edit'),
    (r'管理员', 'Administrator'),
    (r'System管理员权限：只Show转让Button', 'System administrator permission: only show transfer button'),
    (r'转让', 'Transfer'),
    (r'非管理员权限：ShowEditAndDeleteButton', 'Non-administrator permission: show edit and delete buttons'),
    (r'Save原始Value', 'Save original value'),
    (r'权限ModifySuccess', 'Permission modified successfully'),
    (r'Confirm要转让管理员权限吗？转让后您将失去管理员权限。', 'Confirm transfer of administrator permission? You will lose administrator permission after transfer.'),
    (r'转让Confirm', 'Transfer Confirmation'),
    (r'Confirm转让', 'Confirm Transfer'),
    (r'已Cancel转让', 'Transfer cancelled'),
    (r'Confirm要Delete这条Data吗？', 'Confirm deletion of this data?'),
    (r'已CancelDelete', 'Deletion cancelled'),

    # Knowledge base and sections
    (r'片段', 'Segment'),
    (r'片段Information', 'Segment Information'),
    (r'片段Content', 'Segment Content'),
    (r'命中测试', 'Hit Test'),
    (r'测试问题', 'Test Question'),
    (r'测试结果', 'Test Results'),
    (r'相似度', 'Similarity'),
    (r'Document名称', 'Document Name'),
    (r'DocumentInformation', 'Document Information'),
    (r'DocumentContent', 'Document Content'),
    (r'Upload when 间', 'Upload Time'),
    (r'Update when 间', 'Update Time'),
    (r'Create when 间', 'Create Time'),
    (r'DocumentStatus', 'Document Status'),
    (r'Parse中', 'Parsing'),
    (r'ParseSuccess', 'Parse Successful'),
    (r'ParseFailed', 'Parse Failed'),
    (r'等待Parse', 'Waiting to Parse'),

    # Workflow
    (r'工作流名称', 'Workflow Name'),
    (r'工作流Description', 'Workflow Description'),
    (r'节点', 'Node'),
    (r'节点名称', 'Node Name'),
    (r'节点类型', 'Node Type'),
    (r'开始节点', 'Start Node'),
    (r'结束节点', 'End Node'),
    (r'条件节点', 'Condition Node'),
    (r'执行节点', 'Execution Node'),
    (r'连接线', 'Connection Line'),
    (r'运行Status', 'Run Status'),
    (r'运行中', 'Running'),
    (r'运行Success', 'Run Successful'),
    (r'运行Failed', 'Run Failed'),
    (r'等待运行', 'Waiting to Run'),

    # Model access
    (r'Model名称', 'Model Name'),
    (r'ModelType', 'Model Type'),
    (r'Model供应商', 'Model Provider'),
    (r'ModelDescription', 'Model Description'),
    (r'ModelStatus', 'Model Status'),
    (r'可用', 'Available'),
    (r'不可用', 'Unavailable'),
    (r'测试中', 'Testing'),

    # Safety
    (r'安全护栏', 'Safety Guardrail'),
    (r'敏感词', 'Sensitive Word'),
    (r'敏感词库', 'Sensitive Word Library'),
    (r'拦截规Then', 'Interception Rule'),
    (r'拦截类型', 'Interception Type'),
    (r'拦截Content', 'Interception Content'),
    (r'拦截 when 间', 'Interception Time'),

    # Session and chat
    (r'对话History', 'Conversation History'),
    (r'对话Content', 'Conversation Content'),
    (r'用户Input', 'User Input'),
    (r'SystemBack', 'System Response'),
    (r'对话 when 间', 'Conversation Time'),
    (r'清空对话', 'Clear Conversation'),
    (r'Delete对话', 'Delete Conversation'),

    # Common actions
    (r'新增', 'Add'),
    (r'编辑', 'Edit'),
    (r'Delete', 'Delete'),
    (r'Search', 'Search'),
    (r'筛选', 'Filter'),
    (r'排序', 'Sort'),
    (r'导出', 'Export'),
    (r'导入', 'Import'),
    (r'Download', 'Download'),
    (r'Upload', 'Upload'),
    (r'刷新', 'Refresh'),
    (r'重试', 'Retry'),
    (r'续传', 'Resume'),
    (r'暂停', 'Pause'),
    (r'继续', 'Continue'),
    (r'停止', 'Stop'),
    (r'启动', 'Start'),
    (r'重启', 'Restart'),
]

def translate_content(content):
    """Translate content using pattern matching"""
    for pattern, replacement in PATTERN_TRANSLATIONS:
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

