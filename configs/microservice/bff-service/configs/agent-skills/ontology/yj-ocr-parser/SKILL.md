---
name: yj-ocr-parser
description: 文档解析技能，用于解析PDF文档和图片（jpg/png/jpeg）并将其转换为Markdown格式输出。支持标题层级、表格（HTML格式）、公式（LaTeX格式）、图片（链接形式）等复杂内容的解析。当用户提到"解析PDF"、"文档解析"、"PDF转Markdown"、"提取PDF内容"、"解析文档"、"文档内容提取"、"PDF内容识别"、"图片解析"、"图片转文字"、"识别图片内容"等场景时使用此技能。即使用户只是要求读取或查看PDF/图片文件内容，也应考虑使用此技能。
---

# 文档解析技能（yj-ocr-parser）

本技能调用文档解析模型 API，将 PDF 文件和图片（jpg/png/jpeg）解析为结构化的 Markdown 格式内容，支持标题层级、表格、公式、图片等复杂元素的提取与转换。

## 适用场景

- 用户需要解析 PDF 文档或图片（jpg/png/jpeg）并获取 Markdown 格式内容
- 用户需要提取 PDF/图片中的表格、公式、图片等信息
- 用户需要将 PDF 文档或图片内容转换为可编辑的文本格式
- 用户提到"解析文档"、"PDF转MD"、"提取文档内容"、"图片解析"、"图片转文字"等需求

## 变量要求

- 本技能需要 `MaaS_model_token`，其值为文档解析 API 的访问令牌（Access Token）
- 平台会把已配置的技能变量**注入到系统提示中**，形如 `MaaS_model_token = <token>`，位于"已为当前技能配置"的变量列表里。请**直接使用系统提示中提供的该变量值**，不要依赖 shell 环境变量（沙箱中不会导出 `$MaaS_model_token`）
- 如系统提示中未提供该变量，需提醒用户：请在技能变量中配置 `MaaS_model_token`，值为有效的 API 访问令牌

## 使用步骤

1. **确认文件**：确认用户提供了本地 PDF 或图片（jpg/png/jpeg）文件路径
2. **检查格式**：验证文件扩展名是否为支持的格式（pdf、jpg、jpeg、png），不支持 SVG 等矢量图格式
3. **获取变量值**：从系统提示中"已为当前技能配置"的变量列表读取 `MaaS_model_token` 的值
4. **调用 API**：使用 curl 发送 multipart/form-data 请求
5. **返回结果**：将解析结果呈现给用户

## API 调用方式

使用以下 curl 命令调用文档解析 API：

```bash
curl --location 'https://maas-api.ai-yuanjing.com/openapi/v1/rag/model_parser_file' \
  --header 'Authorization: Bearer {MaaS_model_token}' \
  -F 'file=@"{本地文件路径}"' \
  -F 'file_name={文件名}'
```

> 注：`{MaaS_model_token}` 需替换为系统提示中"已为当前技能配置"列出的该变量值，而非 shell 环境变量展开（`$MaaS_model_token` 在沙箱中为空）。

### 参数说明

| 参数名 | 必选 | 类型 | 说明 |
|--------|------|------|------|
| file | 是 | multipart file | 需解析的文件本地路径，支持 PDF 和图片（jpg/jpeg/png），以文件流形式上传 |
| file_name | 是 | string | 文档名称，例如：test.pdf、image.jpg |

### 认证方式

请求头中需携带 `Authorization: Bearer {MaaS_model_token}`，其中 `{MaaS_model_token}` 取自系统提示中为本技能配置的变量值（直接填入该值，不要使用 shell 环境变量）。

## 执行流程

当用户请求解析 PDF 文档或图片时，按以下步骤执行：

1. 获取用户提供的文件路径，验证文件存在且为支持的格式（pdf、jpg、jpeg、png）
2. 若文件为不支持的格式（如 SVG、gif、bmp 等），提示用户仅支持 PDF 和图片（jpg/jpeg/png）
3. 从系统提示中"已为当前技能配置"的变量列表读取 `MaaS_model_token` 的值，若未提供则提示用户在技能变量中配置
4. 提取文件名（从路径中获取文件名部分）
5. 执行 curl 命令调用 API
6. 检查返回结果：
   - `code` 为 `"200"` 表示成功，返回 `content` 字段中的 Markdown 内容
   - `code` 为 `"400"` 表示请求参数错误，提示用户检查文件和参数
   - `code` 为 `"429"` 表示令牌限流，提示用户稍后重试
   - `code` 为 `"500"` 表示服务内部错误，提示用户稍后重试
7. 将解析出的 Markdown 内容呈现给用户

## 返回结果处理

API 成功返回时的响应结构：

```json
{
  "code": "200",
  "status": "success",
  "message": "文档处理完成",
  "content": "解析出的Markdown内容",
  "trace_id": "请求追踪ID"
}
```

- `content`：主要字段，包含解析出的 Markdown 格式内容，其中表格以 HTML 格式表示，公式以 LaTeX 格式表示，图片以链接形式输出

## 注意事项

- 支持 PDF 文件和图片文件（jpg、jpeg、png），不支持 SVG 等矢量图格式
- 文件大小限制为 20MB，大文件建议拆分为小文件后解析
- API 访问频率限制为 3 次/分钟
- 使用前需确保已开通相关权限
- 该接口为同步调用，大文件解析可能需要较长时间，请耐心等待

## 示例

**用户输入**：解析 /home/user/documents/report.pdf

**执行命令**：

```bash
curl --location 'https://maas-api.ai-yuanjing.com/openapi/v1/rag/model_parser_file' \
  --header 'Authorization: Bearer <填入系统提示中提供的 MaaS_model_token 值>' \
  -F 'file=@"/home/user/documents/report.pdf"' \
  -F 'file_name=report.pdf'
```

**用户输入**：解析 /home/user/images/photo.jpg

**执行命令**：

```bash
curl --location 'https://maas-api.ai-yuanjing.com/openapi/v1/rag/model_parser_file' \
  --header 'Authorization: Bearer <填入系统提示中提供的 MaaS_model_token 值>' \
  -F 'file=@"/home/user/images/photo.jpg"' \
  -F 'file_name=photo.jpg'
```

**输出**：将 API 返回的 `content` 字段内容以 Markdown 格式展示给用户。
