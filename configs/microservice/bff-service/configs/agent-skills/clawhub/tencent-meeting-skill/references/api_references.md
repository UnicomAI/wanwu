# 腾讯会议 MCP 工具调用示例

本文件提供各工具的调用示例。参数说明已集成到 MCP 工具 Schema 中，可通过 `tools/list` 查看。

---

## 目录

- [会议管理](#会议管理)
  - [创建会议](#1-schedule_meeting--创建会议)
  - [修改会议](#2-update_meeting--修改会议)
  - [取消会议](#3-cancel_meeting--取消会议)
  - [查询会议详情](#4-get_meeting--查询会议详情)
  - [通过会议号查询](#5-get_meeting_by_code--通过会议号查询)
- [成员管理](#成员管理)
  - [获取参会成员明细](#6-get_meeting_participants--获取参会成员明细)
  - [获取受邀成员列表](#7-get_meeting_invitees--获取受邀成员列表)
  - [查询等候室成员](#8-get_waiting_room--查询等候室成员)
  - [查询用户会议列表](#9-get_user_meetings--查询用户会议列表)
  - [查询已结束会议](#10-get_user_ended_meetings--查询已结束会议)
- [录制与转写](#录制与转写)
  - [查询录制列表](#11-get_records_list--查询录制列表)
  - [获取录制下载地址](#12-get_record_addresses--获取录制下载地址)
  - [查询转写详情](#13-get_transcripts_details--查询转写详情)
  - [查询转写段落](#14-get_transcripts_paragraphs--查询转写段落)
  - [搜索转写内容](#15-search_transcripts--搜索转写内容)
  - [获取智能纪要](#16-get_smart_minutes--获取智能纪要)
- [录制权限申请](#录制权限申请)
  - [申请录制权限-预览](#17-apply_record_permission_prepare--申请录制权限-预览)
  - [申请录制权限-提交](#18-apply_record_permission_commit--申请录制权限-提交)
- [反馈](#反馈)
  - [提交反馈（Agent 意见箱）](#19-submit_feedback--提交反馈agent-意见箱)

---

## 会议管理

### 1. `schedule_meeting` — 创建会议

#### 调用示例

```bash
# 普通会议
python3 scripts/tencent_meeting.py tools/call '{
  "name": "schedule_meeting",
  "arguments": {
    "subject": "产品周会",
    "start_time": "2026-03-25T15:00:00+08:00",
    "end_time": "2026-03-25T16:00:00+08:00"
  }
}'

# 周期性会议（每周开会，共重复5次）
python3 scripts/tencent_meeting.py tools/call '{
  "name": "schedule_meeting",
  "arguments": {
    "subject": "每周例会",
    "start_time": "2026-03-25T15:00:00+08:00",
    "end_time": "2026-03-25T16:00:00+08:00",
    "meeting_type": 1,
    "recurring_rule": {
      "recurring_type": 2,
      "until_type": 1,
      "until_count": 5
    }
  }
}'
```

---

### 2. `update_meeting` — 修改会议

#### 调用示例

```bash
# 修改非周期性会议
python3 scripts/tencent_meeting.py tools/call '{
  "name": "update_meeting",
  "arguments": {
    "meeting_id": "xxx",
    "subject": "新主题",
    "start_time": "2026-03-25T16:00:00+08:00",
    "end_time": "2026-03-25T17:00:00+08:00"
  }
}'

# 修改周期性会议其中一场子会议
python3 scripts/tencent_meeting.py tools/call '{
  "name": "update_meeting",
  "arguments": {
    "meeting_id": "xxx",
    "start_time": "2026-03-26T10:00:00+08:00",
    "end_time": "2026-03-26T11:00:00+08:00",
    "meeting_type": 1,
    "recurring_rule": {
      "sub_meeting_id": "yyy"
    }
  }
}'
```

---

### 3. `cancel_meeting` — 取消会议

#### 调用示例

```bash
# 取消普通会议
python3 scripts/tencent_meeting.py tools/call '{
  "name": "cancel_meeting",
  "arguments": {
    "meeting_id": "xxx"
  }
}'

# 取消周期性会议的某个子会议
python3 scripts/tencent_meeting.py tools/call '{
  "name": "cancel_meeting",
  "arguments": {
    "meeting_id": "xxx",
    "sub_meeting_id": "yyy"
  }
}'

# 取消整场周期性会议
python3 scripts/tencent_meeting.py tools/call '{
  "name": "cancel_meeting",
  "arguments": {
    "meeting_id": "xxx",
    "meeting_type": 1
  }
}'
```

---

### 4. `get_meeting` — 查询会议详情

#### 调用示例

```bash
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_meeting",
  "arguments": {
    "meeting_id": "xxx"
  }
}'
```

---

### 5. `get_meeting_by_code` — 通过会议号查询

#### 调用示例

```bash
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_meeting_by_code",
  "arguments": {
    "meeting_code": "904854736"
  }
}'
```

---

## 成员管理

### 6. `get_meeting_participants` — 获取参会成员明细

#### 调用示例

```bash
# 基础查询（首页）
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_meeting_participants",
  "arguments": {
    "meeting_id": "xxx",
    "page_size": 20
  }
}'

# 翻页查询（使用上次返回的 next_page_token）
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_meeting_participants",
  "arguments": {
    "meeting_id": "xxx",
    "page_size": 20,
    "page_token": "上一次响应中的next_page_token"
  }
}'

# 按参会时间过滤
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_meeting_participants",
  "arguments": {
    "meeting_id": "xxx",
    "start_time": "2026-03-01T00:00:00+08:00",
    "end_time": "2026-03-25T23:59:59+08:00"
  }
}'
```

---

### 7. `get_meeting_invitees` — 获取受邀成员列表

#### 调用示例

```bash
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_meeting_invitees",
  "arguments": {
    "meeting_id": "xxx"
  }
}'
```

---

### 8. `get_waiting_room` — 查询等候室成员

#### 调用示例

```bash
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_waiting_room",
  "arguments": {
    "meeting_id": "xxx",
    "page_size": 20
  }
}'
```

---

### 9. `get_user_meetings` — 查询用户会议列表

#### 调用示例

```bash
# 查询即将开始/进行中的会议
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_user_meetings",
  "arguments": {
    "is_show_all_sub_meetings": 0
  }
}'

# 翻页查询（has_more 为 true 时，使用返回的 next_page_token）
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_user_meetings",
  "arguments": {
    "page_token": "上一次响应中的next_page_token",
    "is_show_all_sub_meetings": 0
  }
}'
```

---

### 10. `get_user_ended_meetings` — 查询已结束会议

#### 调用示例

```bash
# 查询指定日期已结束的会议
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_user_ended_meetings",
  "arguments": {
    "start_time": "2026-03-25T00:00:00+08:00",
    "end_time": "2026-03-25T23:59:59+08:00",
    "page_size": 10
  }
}'

# 翻页查询
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_user_ended_meetings",
  "arguments": {
    "start_time": "2026-03-25T00:00:00+08:00",
    "end_time": "2026-03-25T23:59:59+08:00",
    "page_size": 10,
    "page_token": "上一次响应中的next_page_token"
  }
}'
```

---

## 录制与转写

### 11. `get_records_list` — 查询录制列表

#### 调用示例

```bash
# 按时间范围查询
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_records_list",
  "arguments": {
    "start_time": "2026-03-25T00:00:00+08:00",
    "end_time": "2026-03-25T23:59:59+08:00",
    "page_size": 10
  }
}'

# 按会议ID查询（无需传时间）
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_records_list",
  "arguments": {
    "meeting_id": "xxx"
  }
}'

# 按会议号查询（无需传时间）
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_records_list",
  "arguments": {
    "meeting_code": "904854736"
  }
}'

# 翻页查询
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_records_list",
  "arguments": {
    "start_time": "2026-03-25T00:00:00+08:00",
    "end_time": "2026-03-25T23:59:59+08:00",
    "page_size": 10,
    "page_token": "上一次响应中的next_page_token"
  }
}'
```

---

### 12. `get_record_addresses` — 获取录制下载地址

#### 调用示例

```bash
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_record_addresses",
  "arguments": {
    "meeting_record_id": "xxx"
  }
}'
```

---

### 13. `get_transcripts_details` — 查询转写详情

#### 调用示例

```bash
# 查询转写内容（从第0段开始）
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_transcripts_details",
  "arguments": {
    "record_file_id": "xxx",
    "pid": "0"
  }
}'

# 从指定段落开始查询
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_transcripts_details",
  "arguments": {
    "record_file_id": "xxx",
    "pid": "100"
  }
}'

# 限制查询段落数
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_transcripts_details",
  "arguments": {
    "record_file_id": "xxx",
    "pid": "0",
    "limit": "10"
  }
}'
```

---

### 14. `get_transcripts_paragraphs` — 查询转写段落

#### 调用示例

```bash
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_transcripts_paragraphs",
  "arguments": {
    "record_file_id": "xxx"
  }
}'
```

---

### 15. `search_transcripts` — 搜索转写内容

#### 调用示例

```bash
python3 scripts/tencent_meeting.py tools/call '{
  "name": "search_transcripts",
  "arguments": {
    "record_file_id": "xxx",
    "text": "产品需求"
  }
}'
```

---

### 16. `get_smart_minutes` — 获取智能纪要

#### 调用示例

```bash
# 获取原文智能纪要
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_smart_minutes",
  "arguments": {
    "record_file_id": "xxx"
  }
}'

# 获取英文版智能纪要
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_smart_minutes",
  "arguments": {
    "record_file_id": "xxx",
    "lang": "en"
  }
}'

# 录制文件有密码时
python3 scripts/tencent_meeting.py tools/call '{
  "name": "get_smart_minutes",
  "arguments": {
    "record_file_id": "xxx",
    "pwd": "123456"
  }
}'
```

---

## 录制权限申请

> 录制权限申请采用 **两步流程**：先调用 `apply_record_permission_prepare` 获取预览信息向用户展示并确认，
> 用户明确同意后再调用 `apply_record_permission_commit` 正式提交。**严禁跳过预览阶段直接提交申请。**
> 触发场景：用户访问录制相关内容（下载地址/转写/智能纪要）出现无权限错误，或用户主动请求申请录制权限。

### 17. `apply_record_permission_prepare` — 申请录制权限-预览

#### 调用示例

```bash
# 基础调用：仅传 meeting_record_id
python3 scripts/tencent_meeting.py tools/call '{
  "name": "apply_record_permission_prepare",
  "arguments": {
    "meeting_record_id": "xxx"
  }
}'

# 同时携带 meeting_id（推荐，便于服务端校验）
python3 scripts/tencent_meeting.py tools/call '{
  "name": "apply_record_permission_prepare",
  "arguments": {
    "meeting_record_id": "xxx",
    "meeting_id": "yyy"
  }
}'
```

#### 返回字段说明

| 字段 | 说明 |
|------|------|
| `preview.meeting_record_id` | 会议录制 ID |
| `preview.approval_name` | 申请类型文案 |
| `preview.subject` | 会议标题 |
| `preview.file_owner` | 录制所有者名称 |
| `preview.apply_note` | 权限申请备注信息 |
| `preview.applicant` | 申请人名称 |
| `expires_in` | 预览有效期（秒），接近过期时建议重新调用 |

> **调用后必须**：向用户完整展示预览信息（标题/录制所有者/申请人/申请说明），等待用户明确同意后再调用 commit 工具。

---

### 18. `apply_record_permission_commit` — 申请录制权限-提交

> **前置条件**：已调用 prepare 工具向用户展示预览信息，并获得用户明确同意。

#### 调用示例

```bash
# 基础调用
python3 scripts/tencent_meeting.py tools/call '{
  "name": "apply_record_permission_commit",
  "arguments": {
    "meeting_record_id": "xxx"
  }
}'

# 同时携带 meeting_id
python3 scripts/tencent_meeting.py tools/call '{
  "name": "apply_record_permission_commit",
  "arguments": {
    "meeting_record_id": "xxx",
    "meeting_id": "yyy"
  }
}'
```

#### 返回字段说明

| 字段 | 说明 |
|------|------|
| `unique_id` | 申请 ID |
| `status` | 审批状态 |
| `message` | 审批状态描述 |
| `approval_url` | 审批链接（**必须展示给用户**，便于跟进审批进度） |
| `share_text` | 申请说明描述 |

---

## 反馈

### 19. `submit_feedback` — 提交反馈（Agent 意见箱）

> 由 Agent 调用，用于上报工具缺失、工具报错、能力不足、结果异常或改进建议。
> 参数定义、枚举、长度限制、关联字段条件必填等约束以工具自身的 MCP schema 为准（通过 `tools/list` 获取）；调用时机与输出规范详见 SKILL.md 场景7。

#### 调用示例

```bash
# 1. 找不到对应工具
python3 scripts/tencent_meeting.py tools/call '{
  "name": "submit_feedback",
  "arguments": {
    "category": "tool_not_found",
    "intent": "订阅会议变更事件，实时感知会议被修改/取消"
  }
}'

# 2. 调用工具报错
python3 scripts/tencent_meeting.py tools/call '{
  "name": "submit_feedback",
  "arguments": {
    "category": "tool_error",
    "intent": "查询某场会议的参会成员明细",
    "actions_tried": "调用 get_meeting_participants(meeting_id=xxx)",
    "result": "返回 9042 无权限",
    "tool_name": "get_meeting_participants",
    "error_code": "9042"
  }
}'

# 3. 工具能力/参数不足
python3 scripts/tencent_meeting.py tools/call '{
  "name": "submit_feedback",
  "arguments": {
    "category": "tool_inadequate",
    "intent": "按时间范围筛选用户的未来会议列表",
    "actions_tried": "调用 get_user_meetings",
    "result": "工具不支持 start_time / end_time 参数，无法按时间过滤",
    "tool_name": "get_user_meetings"
  }
}'

# 4. 结果不符预期
python3 scripts/tencent_meeting.py tools/call '{
  "name": "submit_feedback",
  "arguments": {
    "category": "unexpected_result",
    "intent": "获取某场会议的智能纪要",
    "actions_tried": "调用 get_smart_minutes(record_file_id=xxx)",
    "result": "返回空内容，但用户确认会议已生成纪要",
    "tool_name": "get_smart_minutes"
  }
}'

# 5. 一般性建议
python3 scripts/tencent_meeting.py tools/call '{
  "name": "submit_feedback",
  "arguments": {
    "category": "suggestion",
    "intent": "希望 get_records_list 支持按会议主题模糊搜索"
  }
}'
```

---

## 相关文档

- **SKILL.md** — 完整的业务规范与触发场景（时间处理、敏感操作、错误处理等通用规则以 SKILL.md 为准）
- **error_dictionary.md** — 错误处理指引
- **version_management.md** — 版本管理指引
