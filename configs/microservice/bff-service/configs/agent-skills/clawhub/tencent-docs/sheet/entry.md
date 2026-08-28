# Excel 文档（sheet）品类操作指引

本目录提供 Excel 文档（sheet）品类的专业操作能力，包括计算、筛选、统计、Excel操作相关场景。sheet 工具已集成到 tencent-docs MCP 服务中，工具名以 `sheet.` 为前缀。

## 使用场景

> **操作优先级说明：请按以下顺序选择合适的操作方式。**

**🥇 优先使用（重点1）：** 对于以下明确支持的操作，**必须优先**使用 `sheet.*` 系列工具处理：
插入图片、设置单元格值、批量设置单元格值、设置单元格样式、合并/取消合并单元格、插入/删除行列、设置行高列宽、冻结/取消冻结行列、筛选、超链接、清除内容/样式、获取子表信息、获取单元格数据、获取合并单元格信息、插入删除重命名子表。

**🥈 次选使用（重点2）：** 当上述工具无法满足需求时（如涉及更复杂的表格操作），再考虑使用 `sheet.operation_sheet`（`api/operation-api.md`）的 JS 脚本方式完成。

---

## 服务信息

| 项目     | 说明                                                         |
| -------- | ------------------------------------------------------------ |
| 所属服务 | `tencent-docs`                                               |
| 工具前缀 | `sheet.*`（如 `sheet.get_cell_data`、`sheet.set_cell_value`）|
| 调用方式 | 与 tencent-docs 其他工具相同，无需额外配置                   |
| Token    | 使用 tencent-docs 统一 Token，完成授权后自动配置             |
| 文档类型 | 仅支持 Sheet 文档类型                                        |

---

## 文档标识

所有 sheet 工具使用 `file_id` 标识文档：
- `file_id` (string, 必填): 在线表格的唯一标识符

> 💡 **获取 file_id**：可通过 `manage.search_file` 搜索文档获取，或从文档链接中解析。

---

## 工具列表

| 工具名称              | 功能说明             |
| --------------------- | -------------------- |
| sheet.insert_image    | 在指定单元格插入图片 |
| sheet.set_cell_value  | 设置单个单元格的值   |
| sheet.set_range_value | 批量设置单元格的值   |
| sheet.set_cell_style  | 设置单元格的样式     |
| sheet.merge_cell      | 合并单元格           |
| sheet.insert_dimension| 插入行或列           |
| sheet.delete_dimension| 删除行或列           |
| sheet.set_freeze      | 设置冻结行列         |
| sheet.set_filter      | 设置筛选             |
| sheet.remove_filter   | 移除筛选             |
| sheet.set_link        | 设置单元格超链接     |
| sheet.clear_link      | 清除单元格超链接     |
| sheet.clear_range_cells  | 清除区域单元格内容|
| sheet.clear_range_style  | 清除区域单元格样式|
| sheet.get_sheet_info  | 获取子表信息         |
| sheet.clear_range_all | 清空区域内容和样式   |
| sheet.unset_freeze    | 删除所有冻结         |
| sheet.unmerge_cell    | 取消合并单元格       |
| sheet.get_cell_data   | 获取单元格数据       |
| sheet.get_merged_cells| 获取合并单元格信息   |
| sheet.set_dimension_size | 设置行高或列宽    |
| sheet.add_sheet       | 增加子表             |
| sheet.delete_sheet    | 删除子表             |
| sheet.rename_sheet    | 重命名子表           |

---

## 注意事项

- 工具名带 `sheet.` 前缀（如 `sheet.get_cell_data`、`sheet.set_cell_value` 等）
- 操作前需确保拥有文档的写入权限
- 详细 API 参数和调用示例请参考 `api/mcp-api.md`

---

## 按场景工作流

### 设置单元格内容和样式

```
1. 按需调用 sheet.* 工具更新单元格内容或者样式
    - 更新单个单元格内容：sheet.set_cell_value
    - 更新多个单元格内容：sheet.set_range_value
    - 更新单元格样式: sheet.set_cell_style
```

### 插入图片

```
1. 调用 sheet.insert_image，在指定单元格插入图片
2. 小图可以直接传base64编码后的图片内容content
3. 若图片过大导致base64内容超出传输限制，应先调用upload_image工具获取image_id，再调用 sheet.insert_image 传入image_id
4. 需要提供目标sheet_id、row_index、col_index，以及content或image_id
```

### 清除单元格内容和样式

```
1. 按需调用 sheet.* 工具清除单元格内容或者样式
    - 清除单元格内容：sheet.clear_range_cells
    - 清除单元格样式：sheet.clear_range_style
    - 同时清除内容和样式：sheet.clear_range_all
```

### 设置和取消合并单元格

```
1. 调用 sheet.merge_cell，可以生成合并单元格
2. 调用 sheet.unmerge_cell，可以取消合并单元格
```

### 设置和取消筛选

```
1. 调用 sheet.set_filter，可以设置筛选
2. 调用 sheet.remove_filter，可以取消筛选
```

### 设置和取消冻结

```
1. 调用 sheet.set_freeze，可以设置冻结区域
2. 调用 sheet.unset_freeze，可以取消冻结区域
```

### 添加和删除链接

```
1. 调用 sheet.set_link，可以设置链接
2. 调用 sheet.clear_link，可以删除链接
```

### 增删行列

```
1. 调用 sheet.insert_dimension，可以增加行或者列
2. 调用 sheet.delete_dimension，可以删除行或者列
```

### 设置行高列宽

```
1. 调用 sheet.set_dimension_size，可以设置指定行的行高或指定列的列宽，支持批量设置和清除自定义尺寸
```

### 子表管理
```
1. 调用 sheet.add_sheet，可以增加子表，支持指定位置插入和尾部追加两种
2. 调用 sheet.delete_sheet，可以删除指定的子表
3. 调用 sheet.rename_sheet，可以重命名子表
```

### 查询接口

```
1. 调用 sheet.get_sheet_info，获取在线表格的子表信息，包括子表ID、名称、类型、行列数量
2. 调用 sheet.get_cell_data，获取在线表格指定区域的单元格数据，支持返回CSV格式或结构化单元格数据
3. 调用 sheet.get_merged_cells，获取在线表格指定区域内与该区域相交的合并单元格信息，返回合并单元格范围列表
```
