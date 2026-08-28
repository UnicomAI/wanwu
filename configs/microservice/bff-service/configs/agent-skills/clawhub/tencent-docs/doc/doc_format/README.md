# 文本格式化模块

纯文本 → 结构化 XML → 样式美化的工程化流程。

---

## 文件结构

```
doc_format/
├── prompt/
│   ├── scenario_recognition_prompt.txt    # 场景识别 Prompt
│   ├── pure_text_system_prompt.txt        # 文本转 XML Prompt
│   └── style_customization_prompt.txt     # 样式解析 Prompt
└── templates/
    ├── general.json                        # 通用场景模板
    ├── paper.json                          # 学术论文模板
    ├── contract.json                       # 合同模板
    ├── essay.json                          # 作文模板
    ├── government.json                     # 公文模板
```

---

## 工作流程

你需要按照以下步骤完成文本美化任务：

### 步骤 1: 场景识别与标题生成

分析用户提供的文本内容，识别所属场景并生成文档标题。

**参考规则：** `prompt/scenario_recognition_prompt.txt`

**你必须输出给用户：**
```json
{
  "scenario": "场景标识",
  "title": "生成的标题（2-25字符）"
}
```

---

### 步骤 2: 样式自定义（可选）

**仅当用户明确提出样式要求时执行此步骤**，例如：
- "标题用初号黑体"
- "正文改成小四"
- "标题居中显示"

**允许样式：** 参考 `templates/{scenario}.json` 中的 `schema.children[].structure` 字段，必须为叶节点的样式。
**参考规则：** `prompt/style_customization_prompt.txt`

**你必须输出给用户（JSON 数组格式）：**
```json
[
  {
    "structureName": "Title",
    "fontSize": 42,
    "fontFamily": "黑体",
    "fontColor": "AE2E19",
    "alignment": 2,
    "lineSpacing": 1.5
  }
]
```

如果用户没有样式要求，此步骤不输出。

---

### 步骤 3: 文本转 XML 结构化

根据识别的场景，加载对应模板，将纯文本转换为结构化 XML。

**模板位置：** `templates/{scenario}.json`

**参考规则：** `prompt/pure_text_system_prompt.txt`

**你必须输出给用户：**
```json
{
  "xml": "<root>...</root>"
}
```

---

### 步骤 4: 调用套用 MCP 工具

使用 `tencent-docs` MCP Server 对应的 MCP 工具 `doc.ai_format_pure_text` 调用套用 API，传入前面步骤的结果，生成在线腾讯文档链接。

**MCP 工具参数：**
- `title`: 文档标题（步骤 1 的输出）
- `xml`: 格式套用后的文档 XML 结构（步骤 3 的输出）
- `scenario`: 模板场景（步骤 1 的输出）
- `customStyles`: 对文档的自定义样式（步骤 2 的输出，可选，需序列化为 JSON 字符串）

**最终输出文档链接给用户。**

## 注意事项

### JSON 序列化
文本中的引号必须正确转义：

❌ 错误：
```json
{"text": "合同（以下简称"本合同"）"}
```

✅ 正确：
```json
{"text": "合同（以下简称\"本合同\"）"}
```
