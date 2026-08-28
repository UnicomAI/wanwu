# 腾讯文档 Sheet MCP 工具完整参考

本文件包含腾讯文档 Sheet MCP 所有工具的通用 API 说明、详细调用示例、参数说明和返回值说明。

---

## 通用说明

### 公共参数

所有工具都包含以下公共参数：
- `file_id` (string, 可选): 文档唯一标识符
- `sheet_id` (string, 必填): 子表 ID（`get_sheet_info` 不需要此参数）

### 响应结构

所有 API 成功时返回空对象 `{}`，失败时会抛出对应错误信息。

## 工具调用示例

## 1. set_cell_value

### 功能说明
设置在线表格指定单元格的值，支持文本、数字、布尔、公式等类型（SHEET）。

> 💡 **建议**：单次写入操作的请求体内容尽量不超过 **1MB**，超大内容请拆分为多次写入。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "cell": {
    "row": 0,
    "col": 0,
    "value_type": "STRING",
    "string_value": "Hello World"
  }
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `cell` (object, 必填): 单元格值参数
  - `row` (int64, 必填): 行索引（0-based）
  - `col` (int64, 必填): 列索引（0-based）
  - `value_type` (string, 必填): 值类型，可选值：`STRING`、`NUMBER`、`BOOL`、`FORMULA`
  - `number_value` (double, 可选): 数值，`value_type` 为 `NUMBER` 时使用
  - `string_value` (string, 可选): 字符串值，`value_type` 为 `STRING` 时使用
  - `bool_value` (bool, 可选): 布尔值，`value_type` 为 `BOOL` 时使用
  - `formula` (string, 可选): 公式，`value_type` 为 `FORMULA` 时使用，例如 `"=SUM(A1:A10)"`

### 返回值说明
```json
{}
```

---

## 2. set_range_value

### 功能说明
批量设置在线表格多个单元格的值（SHEET）。

> 💡 **建议**：单次写入操作的请求体内容尽量不超过 **1MB**（大约几千个单元格，视单元格内容长度而定），超大批量请拆分为多次写入。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "values": [
    {
      "row": 0,
      "col": 0,
      "value_type": "STRING",
      "string_value": "Name"
    },
    {
      "row": 0,
      "col": 1,
      "value_type": "STRING",
      "string_value": "Score"
    },
    {
      "row": 1,
      "col": 0,
      "value_type": "STRING",
      "string_value": "Alice"
    },
    {
      "row": 1,
      "col": 1,
      "value_type": "NUMBER",
      "number_value": 95.5
    }
  ]
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `values` (array, 必填): 单元格值列表，每个元素与 `set_cell_value` 的 `cell` 参数结构相同

### 返回值说明
```json
{}
```

---

## 3. set_cell_style

### 功能说明
设置在线表格指定范围单元格的样式，包括字体、颜色、对齐等（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "start_row": 0,
  "start_col": 0,
  "end_row": 5,
  "end_col": 3,
  "format": {
    "bold": true,
    "italic": false,
    "font_size": 12,
    "font_color": "FF000000",
    "bg_color": "FFFFFF00",
    "horizontal_align": "center",
    "vertical_align": "center",
    "wrap_text": true
  }
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `start_row` (int64, 必填): 起始行索引（0-based）
- `start_col` (int64, 必填): 起始列索引（0-based）
- `end_row` (int64, 必填): 结束行索引
- `end_col` (int64, 必填): 结束列索引
- `format` (object, 必填): 样式参数对象  - `bold` (bool, 可选): 是否粗体
  - `italic` (bool, 可选): 是否斜体
  - `font_family` (string, 可选): 字体名称
  - `font_size` (int32, 可选): 字号（pt）
  - `font_color` (string, 可选): 字体颜色，ARGB hex，如 `"FF000000"`
  - `bg_color` (string, 可选): 背景色，ARGB hex，如 `"FFFFFFFF"`
  - `horizontal_align` (string, 可选): 水平对齐：`general` / `left` / `center` / `right` / `fill` / `justify`
  - `vertical_align` (string, 可选): 垂直对齐：`top` / `center` / `bottom` / `justify`
  - `wrap_text` (bool, 可选): 是否自动换行
  - `strike_through` (bool, 可选): 是否删除线
  - `underline` (string, 可选): 下划线类型：`none` / `single` / `double` / `single_accounting` / `double_accounting`
  - `number_format_pattern` (string, 可选): 数字格式，如 `"0.00%"`
- `is_clear` (bool, 可选): 若为 true，则清除格式

### 返回值说明
```json
{}
```

---

## 4. merge_cell

### 功能说明
合并在线表格指定范围的单元格，支持全部合并、按行合并、按列合并（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "start_row": 0,
  "start_col": 0,
  "end_row": 3,
  "end_col": 3,
  "merge_type": "all"
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `start_row` (int64, 必填): 起始行索引（0-based）
- `start_col` (int64, 必填): 起始列索引（0-based）
- `end_row` (int64, 必填): 结束行索引
- `end_col` (int64, 必填): 结束列索引
- `merge_type` (string, 必填): 合并类型
  - `"all"`: 全部合并（默认）
  - `"columns"`: 按列合并
  - `"rows"`: 按行合并

### 返回值说明
```json
{}
```

---

## 5. insert_dimension

### 功能说明
在在线表格指定位置插入行或列（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "dimension_type": "row",
  "index": 2,
  "count": 3,
  "direction": "before"
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `dimension_type` (string, 必填): 行列类型：`"row"` | `"col"`
- `index` (int64, 必填): 起始索引（0-based）
- `count` (int64, 必填): 插入数量
- `direction` (string, 可选): 插入方向：`"before"`（默认）| `"after"`

### 返回值说明
```json
{}
```

---

## 6. delete_dimension

### 功能说明
删除在线表格指定位置的行或列（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "dimension_type": "col",
  "index": 3,
  "count": 2
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `dimension_type` (string, 必填): 行列类型：`"row"` | `"col"`
- `index` (int64, 必填): 起始索引（0-based）
- `count` (int64, 必填): 删除数量

### 返回值说明
```json
{}
```

---

## 7. set_freeze

### 功能说明
设置在线表格的冻结行列数，传 0 可取消冻结（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "row_count": 1,
  "col_count": 2
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `row_count` (int64, 必填): 冻结行数（0 = 取消冻结行）
- `col_count` (int64, 必填): 冻结列数（0 = 取消冻结列）

### 返回值说明
```json
{}
```

---

## 8. set_filter

### 功能说明
为在线表格指定数据区域设置筛选（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "start_row": 0,
  "start_col": 0,
  "end_row": 100,
  "end_col": 5
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `start_row` (int64, 必填): 数据区域起始行（0-based）
- `start_col` (int64, 必填): 数据区域起始列（0-based）
- `end_row` (int64, 必填): 数据区域结束行
- `end_col` (int64, 必填): 数据区域结束列
- `filter_id` (string, 可选): 筛选 ID（不传则自动生成）

### 返回值说明
```json
{}
```

---

## 9. remove_filter

### 功能说明
移除在线表格的筛选，可按筛选 ID 精确移除或移除全部（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "filter_id": "filter_001"
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `filter_id` (string, 可选): 筛选 ID（不传则移除该子表所有筛选）

### 返回值说明
```json
{}
```

---

## 10. set_link

### 功能说明
为在线表格指定单元格设置超链接，可指定链接 URL 和显示文本（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "row": 0,
  "col": 0,
  "url": "https://docs.qq.com",
  "display_text": "腾讯文档"
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `row` (int64, 必填): 单元格行（0-based）
- `col` (int64, 必填): 单元格列（0-based）
- `url` (string, 必填): 超链接 URL
- `display_text` (string, 可选): 单元格显示文本

### 返回值说明
```json
{}
```

---

## 11. clear_link

### 功能说明
清除在线表格指定单元格的超链接，可按链接 ID 精确清除或清除全部超链接（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "row": 0,
  "col": 0,
  "link_id": "link_001"
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `row` (int64, 必填): 单元格行（0-based）
- `col` (int64, 必填): 单元格列（0-based）
- `link_id` (string, 可选): 链接 ID（不传则按位置清除）

### 返回值说明
```json
{}
```

---

## 12. unmerge_cell

### 功能说明
取消在线表格指定区域的单元格合并（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "start_row": 0,
  "start_col": 0,
  "end_row": 3,
  "end_col": 3
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `start_row` (int64, 必填): 起始行索引（0-based）
- `start_col` (int64, 必填): 起始列索引（0-based）
- `end_row` (int64, 必填): 结束行索引
- `end_col` (int64, 必填): 结束列索引

### 返回值说明
```json
{}
```

---

## 13. clear_range_cells

### 功能说明
清除在线表格指定区域内所有单元格的内容，不影响样式（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "start_row": 0,
  "start_col": 0,
  "end_row": 9,
  "end_col": 4
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `start_row` (int64, 必填): 起始行索引（0-based）
- `start_col` (int64, 必填): 起始列索引（0-based）
- `end_row` (int64, 必填): 结束行索引
- `end_col` (int64, 必填): 结束列索引

### 返回值说明
```json
{}
```

---

## 14. clear_range_style

### 功能说明
清除在线表格指定区域内所有单元格的样式，不影响内容（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "start_row": 0,
  "start_col": 0,
  "end_row": 9,
  "end_col": 4
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `start_row` (int64, 必填): 起始行索引（0-based）
- `start_col` (int64, 必填): 起始列索引（0-based）
- `end_row` (int64, 必填): 结束行索引
- `end_col` (int64, 必填): 结束列索引

### 返回值说明
```json
{}
```

---

## 15. clear_range_all

### 功能说明
清空在线表格指定区域内所有单元格的内容和样式（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "start_row": 0,
  "start_col": 0,
  "end_row": 9,
  "end_col": 4
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `start_row` (int64, 必填): 起始行索引（0-based）
- `start_col` (int64, 必填): 起始列索引（0-based）
- `end_row` (int64, 必填): 结束行索引
- `end_col` (int64, 必填): 结束列索引

### 返回值说明
```json
{}
```

---

## 16. unset_freeze

### 功能说明
删除在线表格指定子表的所有冻结行列（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001"
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID

### 返回值说明
```json
{}
```

---

## 17. get_sheet_info

### 功能说明
获取在线表格的子表信息，包括子表 ID、名称、类型、行列数量（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890"
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID

> 注意：此工具不需要 `sheet_id` 参数，返回文档下所有子表的信息。

### 返回值说明
```json
{
  "sheets": [
    {
      "sheet_id": "sub_sheet_001",
      "sheet_name": "Sheet1",
      "sheet_type": "worksheet",
      "row_count": 100,
      "col_count": 26
    }
  ]
}
```
- `sheets` (array): 子表信息列表
  - `sheet_id` (string): 子表 ID
  - `sheet_name` (string): 子表名称
  - `sheet_type` (string): 子表类型：`worksheet` / `smartsheet` / `smartcanvas`
  - `row_count` (int32): 行数
  - `col_count` (int32): 列数

---

## 18. get_cell_data

### 功能说明
获取在线表格指定区域的单元格数据，支持返回 CSV 格式或结构化单元格数据（SHEET）。

> ⚠️ **限制**：单次请求的单元格范围不得超过 **20000** 个（即 `(end_row - start_row + 1) × (end_col - start_col + 1) ≤ 20000`），超出将返回错误。如需获取更大范围的数据，请分多次请求。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "start_row": 0,
  "start_col": 0,
  "end_row": 9,
  "end_col": 4,
  "return_csv": false
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `start_row` (int64, 必填): 起始行索引（0-based）
- `start_col` (int64, 必填): 起始列索引（0-based）
- `end_row` (int64, 必填): 结束行索引
- `end_col` (int64, 必填): 结束列索引
- `return_csv` (bool, 可选): 是否以 CSV 格式返回数据，`true` 返回 `csv_data`，`false` 返回 `cells` 结构化数据（默认 `false`）

### 返回值说明
```json
{
  "csv_data": "Name,Score\nAlice,95.5\n",
  "cells": [
    {
      "row": 0,
      "col": 0,
      "value_type": "STRING",
      "string_value": "Name"
    },
    {
      "row": 0,
      "col": 1,
      "value_type": "STRING",
      "string_value": "Score"
    }
  ]
}
```
- `csv_data` (string): CSV 格式数据（`return_csv=true` 时返回）
- `cells` (array): 结构化单元格数据（`return_csv=false` 时返回）
  - `row` (int32): 行索引（0-based）
  - `col` (int32): 列索引（0-based）
  - `value_type` (string): 值类型：`NUMBER` / `STRING` / `BOOL` / `FORMULA` / `ERROR` / `TIME_STRING` / `RICH_STRING`
  - `number_value` (double): 数值
  - `string_value` (string): 字符串值
  - `bool_value` (bool): 布尔值
  - `formula` (string): 公式

---

## 19. get_merged_cells

### 功能说明
获取在线表格指定区域内与该区域相交的合并单元格信息，返回合并单元格范围列表（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "start_row": 0,
  "start_col": 0,
  "end_row": 9,
  "end_col": 9
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `start_row` (int64, 必填): 查询区域起始行索引（0-based）
- `start_col` (int64, 必填): 查询区域起始列索引（0-based）
- `end_row` (int64, 必填): 查询区域结束行索引
- `end_col` (int64, 必填): 查询区域结束列索引

### 返回值说明
```json
{
  "merged_cells": [
    "sub_sheet_001$A1:B2",
    "sub_sheet_001$C3:D5"
  ]
}
```
- `merged_cells` (array): 与查询区域相交的合并单元格范围列表，格式为 `"SheetID$A1:B2"`（列使用字母表示，A=第0列，B=第1列，以此类推）

---

## 20. set_dimension_size

### 功能说明
设置在线表格指定行的行高或指定列的列宽，支持批量设置多个行列的尺寸，也支持清除自定义尺寸恢复默认值（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "dimensions": [
    {
      "dimension_type": "row",
      "index": 0,
      "size": 40
    },
    {
      "dimension_type": "col",
      "index": 2,
      "size": 120
    },
    {
      "dimension_type": "row",
      "index": 5,
      "is_clear": true
    }
  ]
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `dimensions` (array, 必填): 行高/列宽参数列表，每个元素包含：
  - `dimension_type` (string, 必填): 行列类型：`"row"` | `"col"`
  - `index` (int64, 必填): 行或列的索引（0-based）
  - `size` (number, 可选): 行高或列宽的值（行高单位为pt，列宽单位为像素），`is_clear` 为 `true` 时该字段将被忽略
  - `is_clear` (bool, 可选): 是否清除自定义行高/列宽并恢复默认值，为 `true` 时 `size` 字段将被忽略

### 返回值说明
```json
{}
```

---

## 21. add_sheet

### 功能说明
在在线表格中添加一个新的子表，支持指定子表名称和位置（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "name": "新子表",
  "index": 0,
  "append_index": false
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `name` (string, 可选): 子表名称，长度限制为 31 个字符，不传则使用默认名称
- `index` (int64, 可选): 子表位置索引（0-based），不传或 `append_index` 为 `true` 时追加到末尾
- `append_index` (bool, 可选): 是否追加到末尾，为 `true` 时 `index` 字段将被忽略

> 注意：此工具不需要 `sheet_id` 参数，用于创建新的子表。

### 返回值说明
```json
{
  "sheet_id": "new_sheet_001"
}
```
- `sheet_id` (string): 新创建的子表 ID

---

## 22. delete_sheet

### 功能说明
删除在线表格中指定的子表（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001"
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 要删除的子表 ID

### 返回值说明
```json
{}
```

---

## 23. rename_sheet

### 功能说明
重命名在线表格中指定的子表（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "name": "新名称"
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `name` (string, 必填): 新的子表名称，长度限制为 31 个字符

### 返回值说明
```json
{}
```

---

## 24. insert_image

### 功能说明
在在线表格指定单元格插入一张图片，图片内容可通过 base64 或 image_id 传入（SHEET）。

### 调用示例
```json
{
  "file_id": "sheet_1234567890",
  "sheet_id": "sub_sheet_001",
  "row_index": 0,
  "col_index": 0,
  "content": "iVBORw0KGgoAAAANSUhEUgAA..."
}
```

### 参数说明
- `file_id` (string, 可选): 文档 ID
- `sheet_id` (string, 必填): 子表 ID
- `row_index` (int64, 必填): 目标行索引（0-based）
- `col_index` (int64, 必填): 目标列索引（0-based）
- `content` (string, 可选): 图片的 base64 内容，与 `image_id` 二选一，适合图片体积较小的场景；若图片过大导致 base64 内容超出传输限制，请改用 `image_id` 方式
- `image_id` (string, 可选): 图片的 image_id，本质是对图片信息加密后的字符串，与 `content` 二选一，适合图片体积较大的场景。获取方式：
  - 通过 `upload_image` MCP 接口上传图片后获取
  - 通过[腾讯文档开放平台 OpenAPI](https://docs.qq.com/open/developers/?nlc=1#/login) 图片上传接口获取（需先完成 OAuth 授权流程获取 `Access-Token`），示例命令：

```bash
curl --location --request POST 'https://docs.qq.com/openapi/resources/v2/images' \
  --header 'Access-Token: ACCESS_TOKEN' \
  --header 'Client-Id: CLIENT_ID' \
  --header 'Open-Id: OPEN_ID' \
  --form 'image=@"/path/to/your/image.png"'
```

上传成功后，取返回结果中的 `imageID` 字段值传入此参数

### 返回值说明
```json
{}
```
