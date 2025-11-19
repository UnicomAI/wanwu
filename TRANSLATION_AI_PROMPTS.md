# AI Translation Prompts for Wanwu Platform

This document contains ready-to-use prompts for translating different parts of the Wanwu platform. Copy and paste these into lingo.dev, Claude, ChatGPT, or DeepL.

---

## Prompt 1: Frontend UI Translation (Phase 1)

```markdown
# TRANSLATION TASK: Wanwu Platform Frontend UI (Chinese → English)

## Context
Enterprise-grade AI Agent Development Platform
Target Audience: Business users, developers, system administrators
Tone: Professional, clear, user-friendly

## Technical Constraints
1. PRESERVE all JavaScript object keys (left side of colon)
2. TRANSLATE only values (right side of colon)
3. MAINTAIN all formatting: quotes, commas, brackets
4. KEEP placeholders intact: {variable}, {{name}}, ${value}
5. PRESERVE HTML tags: <span>, </span>, etc.

## Translation Glossary (CRITICAL - Use These Exactly)

### Core Platform Terms
- 智能体 → Agent
- 知识库 → Knowledge Base
- 工作流 → Workflow
- 应用广场 → App Marketplace
- 模板广场 → Template Gallery
- 资源库 → Resource Library
- 工作室 → Studio
- 安全护栏 → Safety Guardrails

### Feature Terms
- 文本问答 → Text Q&A
- 问答 → Q&A
- 模型管理 → Model Management
- 组织管理 → Organization Management
- 角色管理 → Role Management
- 用户管理 → User Management
- 个人信息 → Profile
- 系统管理员 → System Administrator

### Technical Terms (Keep in English)
- API, Token, MCP, RAG, OCR, URL, JSON, HTTP, GPU
- Rerank, Embedding, OpenAPI, OAuth

### Action Terms
- 创建/创 建 → Create
- 编辑 → Edit
- 删除 → Delete
- 发布 → Publish
- 取消发布 → Unpublish
- 保存/保 存 → Save
- 确定/确 定 → Confirm
- 取消/取 消 → Cancel
- 查看 → View
- 搜索 → Search
- 导入 → Import
- 导出 → Export

### Knowledge Base Terms
- 分段 → Chunk/Segment (use "Chunk" for technical, "Segment" for UI)
- 命中测试 → Hit Test
- 元数据 → Metadata
- 向量检索 → Vector Search
- 全文检索 → Full-text Search
- 混合检索 → Hybrid Search
- 父子分段 → Parent-Child Chunking

### Status Terms
- 已发布 → Published
- 未发布 → Unpublished
- 处理中 → Processing
- 处理完成 → Completed
- 解析失败 → Failed
- 正在解析 → Parsing
- 待处理 → Pending

## Style Guidelines
1. Use American English spelling (e.g., "color" not "colour")
2. Use title case for menu items: "Model Management" not "model management"
3. Use sentence case for descriptions: "Create a new agent" not "Create A New Agent"
4. Button text: Keep concise (max 2-3 words)
5. Error messages: Be clear and actionable
6. Placeholders: Use "Enter..." or "Select..." format

## Examples

✅ CORRECT:
```javascript
login: {
    title: 'Login',
    form: {
        username: 'Username',
        password: 'Password'
    },
    button: 'Login'
}
```

❌ INCORRECT:
```javascript
denglu: {  // WRONG - key should stay as 'login'
    title: 'Sign In',  // Inconsistent with 'Login'
    form: {
        username: 'User Name',  // Should be one word
    }
}
```

## Source Text to Translate

[PASTE THE web/src/lang/zh.js CONTENT HERE]

## Output Required
Provide the complete translated JavaScript object maintaining exact structure.
```

---

## Prompt 2: Backend Error Messages (Phase 2)

```markdown
# TRANSLATION TASK: Backend Error Messages (Chinese → English)

## Context
Backend error messages for API responses and system logs
Must be clear, actionable, and professional

## Requirements
1. Translate error messages concisely
2. Include technical details when needed
3. Use passive voice sparingly
4. Provide actionable guidance when possible

## Error Message Format
- Pattern: "[Context] [What went wrong] [Suggested action]"
- Example: "Model import failed: Invalid API key format. Please verify your API key."

## Common Error Patterns

### Database Errors
- 创建失败 → "Failed to create [resource]"
- 更新失败 → "Failed to update [resource]"
- 删除失败 → "Failed to delete [resource]"
- 查询失败 → "Failed to retrieve [resource]"
- 记录不存在 → "[Resource] not found"
- 记录已存在 → "[Resource] already exists"

### Validation Errors
- 参数错误 → "Invalid parameter: [field]"
- 格式错误 → "Invalid format for [field]"
- 必填项为空 → "[Field] is required"
- 超出限制 → "[Field] exceeds maximum limit of [value]"

### Permission Errors
- 无权限 → "Insufficient permissions"
- 未授权 → "Unauthorized access"
- 已停用 → "[Resource] is disabled"

### Business Logic Errors
- 操作失败 → "Operation failed"
- 状态不允许 → "Operation not allowed in current state"
- 重复操作 → "Duplicate operation detected"

## Technical Terms
Keep these in English:
- Token, API Key, JWT, OAuth
- Elasticsearch, MinIO, MySQL, Redis
- gRPC, HTTP, WebSocket
- UUID, ID, URL

## Source Text
[PASTE Proto file or error code definitions HERE]

## Output
Translated error messages maintaining the same format structure.
```

---

## Prompt 3: Documentation Translation (Phase 4)

```markdown
# TRANSLATION TASK: Technical Documentation (Chinese → English)

## Context
User-facing documentation for Wanwu AI Agent Platform
Markdown format with screenshots and code examples

## Requirements
1. Translate markdown text preserving formatting
2. Keep code blocks unchanged
3. Translate image alt text
4. Update image paths if screenshots need English versions
5. Preserve all links and anchors

## Style Guide
- Use second person ("you") for instructions
- Active voice preferred
- Short paragraphs (2-4 sentences)
- Numbered lists for procedures
- Bullet points for features/benefits

## Technical Writing Standards
1. **Headers**: Title Case
2. **Instructions**: "To do X, follow these steps:"
3. **Notes**: Use "**Note:**" prefix
4. **Warnings**: Use "**Important:**" or "**Warning:**"
5. **Code**: Keep unchanged, translate only comments

## Section-Specific Guidelines

### Model Management Docs
- Emphasize compatibility and standards
- Include provider names in English
- Keep model names as-is (e.g., "yuanjing-70b-chat")

### Knowledge Base Docs
- Explain chunking strategies clearly
- Use diagrams where helpful
- Provide concrete examples

### Workflow Docs
- Step-by-step tutorials
- Screenshot each major step
- Include troubleshooting section

### Agent Docs
- Use case examples
- Configuration best practices
- Performance optimization tips

## Preserve These Elements
- File paths: `/path/to/file.js`
- Commands: `docker compose up -d`
- URLs: `https://example.com`
- Code variables: `${VARIABLE_NAME}`
- Keyboard shortcuts: `Ctrl+C`

## Example Transformation

**Source (Chinese):**
```markdown
## 模型管理

用户可以在模型管理页面导入各种大语言模型。

### 导入步骤

1. 点击"导入模型"按钮
2. 填写模型信息
3. 点击"确认"保存
```

**Target (English):**
```markdown
## Model Management

You can import various large language models on the Model Management page.

### Import Steps

1. Click the **Import Model** button
2. Fill in the model information
3. Click **Confirm** to save
```

## Source Text
[PASTE Markdown documentation HERE]

## Output
Translated markdown preserving all formatting and structure.
```

---

## Prompt 4: Code Comments Translation (Phase 3)

```markdown
# TRANSLATION TASK: Code Comments (Chinese → English)

## Context
Inline code comments for Go and Python services
Target audience: International developers

## Requirements
1. Translate comments to clear English
2. Maintain technical accuracy
3. Keep comment format (// or # or /**/)
4. Preserve [EN] markers if present
5. Don't translate code itself

## Comment Style
- **Go**: Use complete sentences with proper punctuation
- **Python**: Follow PEP 257 docstring conventions
- Be concise but informative
- Explain "why" not just "what"

## Examples

**Go Before:**
```go
// 检查是否已存在相同的记录
if err := tx.Where("name = ?", name).First(&record).Error; err == nil {
    return errors.New("记录已存在")
}
```

**Go After:**
```go
// Check if a record with the same name already exists
if err := tx.Where("name = ?", name).First(&record).Error; err == nil {
    return errors.New("record already exists")
}
```

**Python Before:**
```python
# 解析单个文档
def parse_doc(file_url, sentence_size):
    """
    参数:
    file_url (str): 文件URL
    sentence_size (int): 句子大小
    """
```

**Python After:**
```python
# Parse a single document
def parse_doc(file_url, sentence_size):
    """
    Parameters:
    file_url (str): File URL
    sentence_size (int): Sentence size
    """
```

## Technical Term Preservation
DO NOT translate these in code:
- Variable names: `user_id`, `kb_name`
- Function names: `create_user()`, `get_model()`
- Constants: `MAX_SIZE`, `DEFAULT_TIMEOUT`
- Package names: `langchain`, `elasticsearch`

## Source Code
[PASTE code file HERE]

## Output
Code file with translated comments, preserving all code and structure.
```

---

## Prompt 5: Validation & Review (Use After Translation)

```markdown
# VALIDATION TASK: Review Translation Quality

## Check These Items

### 1. Completeness
- [ ] All Chinese text translated
- [ ] No mixed Chinese-English
- [ ] All keys/values accounted for

### 2. Consistency
- [ ] Same term translated consistently
- [ ] Glossary terms used correctly
- [ ] Style guide followed

### 3. Technical Accuracy
- [ ] Technical terms correct
- [ ] Code examples work
- [ ] Links functional

### 4. Format Integrity
- [ ] JSON/JavaScript valid
- [ ] Markdown renders correctly
- [ ] Code syntax preserved

### 5. User Experience
- [ ] Messages clear and helpful
- [ ] Instructions easy to follow
- [ ] Tone appropriate

## Review This Translation:
[PASTE translated content HERE]

## Provide:
1. List of issues found (bullet points)
2. Suggested corrections for each issue
3. Overall quality score (1-10)
4. Any inconsistencies with glossary
5. Suggested improvements
```

---

## Quick Reference: Common UI Patterns

```javascript
// Buttons
'创 建' → 'Create'
'确 定' → 'Confirm'
'取 消' → 'Cancel'
'保 存' → 'Save'
'编 辑' → 'Edit'
'删 除' → 'Delete'

// Placeholders
'请输入' → 'Enter'
'请选择' → 'Select'
'请输入搜索内容' → 'Enter search term'
'请输入用户名' → 'Enter username'

// Messages
'操作成功' → 'Operation successful'
'操作失败' → 'Operation failed'
'确定要删除吗？' → 'Are you sure you want to delete this?'
'删除成功' → 'Deleted successfully'
'创建成功' → 'Created successfully'
'更新成功' → 'Updated successfully'

// Status
'加载中...' → 'Loading...'
'暂无数据' → 'No data available'
'正在处理' → 'Processing...'
'处理完成' → 'Completed'
'处理失败' → 'Failed'

// Time
'创建时间' → 'Created At'
'更新时间' → 'Updated At'
'最近一周' → 'Last 7 Days'
'最近一月' → 'Last 30 Days'

// Form Validation
'必填项' → 'Required'
'格式错误' → 'Invalid format'
'超出长度限制' → 'Exceeds character limit'
'密码不一致' → 'Passwords do not match'

// Common Actions
'上传文件' → 'Upload File'
'下载文件' → 'Download File'
'复制链接' → 'Copy Link'
'分享' → 'Share'
'收藏' → 'Favorite'
'取消收藏' → 'Unfavorite'
```

---

## Batch Translation Script Template

For automated translation of multiple files:

```bash
#!/bin/bash
# translate-batch.sh

# Usage: ./translate-batch.sh <input_dir> <output_dir> <prompt_file>

INPUT_DIR=$1
OUTPUT_DIR=$2
PROMPT_FILE=$3

# Create output directory if not exists
mkdir -p "$OUTPUT_DIR"

# Loop through all files
for file in "$INPUT_DIR"/*.md; do
    filename=$(basename "$file")
    echo "Translating: $filename"
    
    # Combine prompt with file content
    cat "$PROMPT_FILE" > temp_prompt.txt
    echo "\n\n## Source Text\n" >> temp_prompt.txt
    cat "$file" >> temp_prompt.txt
    
    # Send to translation API (example with Claude API)
    # Adjust this based on your chosen tool
    # claude translate temp_prompt.txt > "$OUTPUT_DIR/$filename"
    
    rm temp_prompt.txt
done

echo "Translation complete!"
```

---

## Testing Checklist After Translation

### Phase 1 - Frontend UI
```bash
# Start dev server
cd web && npm run dev

# Test these areas:
1. Login page - all text English
2. Main navigation menu - check all menu items
3. Model Management page - buttons, forms, tables
4. Knowledge Base - upload dialogs, settings
5. Agent creation - all configuration options
6. Workflow editor - node descriptions
7. Settings pages - all tabs
8. Error messages - trigger various errors

# Browser console:
- No Chinese characters in console logs
- Check network responses for i18n
```

### Phase 2 - Backend Errors
```bash
# Test API errors
curl -X POST http://localhost:8080/api/test \
  -H "Accept-Language: en" \
  -d '{"invalid": "data"}'

# Check logs
tail -f logs/app.log | grep -v "[\u4e00-\u9fa5]"
```

### Phase 4 - Documentation
```markdown
# Manual checks:
1. Open each .md file in VS Code
2. Preview in markdown viewer
3. Check all images load
4. Click all internal links
5. Test code examples work
6. Verify screenshots show English UI
```

---

## Translation Progress Tracker

Use this template to track progress:

```markdown
# Translation Progress

## Phase 1: Frontend UI
- [ ] Login module (zh.js lines 1-50)
- [ ] Common components (zh.js lines 51-226)
- [ ] Menu items (zh.js lines 57-95)
- [ ] User management (zh.js lines 291-331)
- [ ] Role management (zh.js lines 332-361)
- [ ] Organization management (zh.js lines 362-387)
- [ ] Model access (zh.js lines 423-466)
- [ ] Knowledge management (zh.js lines 583-880)
- [ ] Tool/MCP sections (zh.js lines 900-1059)
- [ ] Agent sections (zh.js lines 1060-1153)
- [ ] Statistics (zh.js lines 1179-1197)

## Phase 2: Backend Errors
- [ ] /pkg/i18n files
- [ ] /proto/err-code/err-code.proto
- [ ] Service error messages

## Phase 3: Code Comments
- [ ] /internal/**/*.go (50 files)
- [ ] /pkg/**/*.go (20 files)
- [ ] /agent/**/*.py (30 files)
- [ ] /rag/**/*.py (20 files)

## Phase 4: Documentation
- [ ] Platform introduction
- [ ] Model management guide
- [ ] Knowledge base docs (10 files)
- [ ] Workflow docs (15 files)
- [ ] Agent docs
- [ ] Settings docs

## Quality Checks
- [ ] Native English speaker review
- [ ] Technical accuracy check
- [ ] Consistency validation
- [ ] User testing
```

---

## Tips for Using These Prompts

1. **Start Small**: Test with one file first before batch processing
2. **Review Output**: Always manually review AI translations
3. **Iterate**: If output quality is poor, refine the prompt
4. **Use Context**: Provide file purpose in prompt for better results
5. **Maintain Glossary**: Update glossary as you find new terms
6. **Track Changes**: Use git branches for each phase
7. **Get Feedback**: Have someone else review critical sections

---

## Common Translation Mistakes to Avoid

❌ **Don't**:
- Translate technical terms like API, JSON, URL
- Change placeholder variable names
- Break JSON/JS syntax
- Use inconsistent terminology
- Over-translate (some English words are fine in Chinese version)

✅ **Do**:
- Keep technical accuracy
- Maintain consistent tone
- Preserve formatting
- Test after translation
- Use the glossary
- Get peer review

---

## Support Resources

- **Glossary**: See TRANSLATION_PLAN.md
- **Full Plan**: TRANSLATION_PLAN.md
- **Questions**: Open GitHub issue with tag `translation`

**Last Updated**: 2025-01-19
