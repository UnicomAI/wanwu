---
name: tencent-meeting-mcp
description: "腾讯会议：会议管理与音视频协作助手。预约/创建/修改/取消会议、查询会议详情与会议号转换、查看参会成员/受邀人/等候室成员、查询用户会议列表（即将开始/进行中/已结束）、查询录制列表与下载地址、获取转写全文/段落/搜索、获取AI智能纪要（支持多语言翻译）、录制权限申请（预览+提交两步流程）、时间转换与版本检查、Agent意见箱反馈上报。当用户需要预约或管理腾讯会议、查看参会人员、查询会议录制或转写内容、获取智能纪要时使用；当用户访问录制相关内容（下载地址/转写/智能纪要）出现无权限错误时，自动发起录制权限申请流程；当Agent遇到以下任何情况时必须立即触发submit_feedback：工具缺失（功能不存在）、工具报错（返回错误码）、能力不足（参数/功能限制）、结果不符预期（数据异常）、改进建议（功能优化）；如果用户提供的是9位会议号而非meeting_id，优先使用get_meeting_by_code转换后再操作。不要在以下场景触发：日程管理（非腾讯会议日程）、即时通讯/聊天、腾讯文档操作、企业微信审批流程、电话/PSTN拨号、视频剪辑或视频编辑、其他视频会议平台（Zoom/Teams/飞书/钉钉）。"
homepage: https://meeting.tencent.com/
metadata:
  {
    "openclaw":
      {
        "requires": { "bins": ["python3"], "env": ["TENCENT_MEETING_TOKEN"] },
        "primaryEnv": "TENCENT_MEETING_TOKEN",
        "category": "tencent",
        "tencentTokenMode": "custom",
        "tokenUrl": "https://mcp.meeting.tencent.com/mcp/wemeet-open/v1",
        "emoji": "📅"
      }
  }
---

# 腾讯会议 MCP 服务

## 概述

本技能为腾讯会议提供完整的 MCP 工具集，涵盖会议管理、成员管理、录制、转写与智能纪要查询等核心功能。

**⚠️ 最高优先级提示**：当遇到以下任何情况时，必须立即暂停当前任务并主动触发 `submit_feedback` 工具上报反馈（详见场景7）：工具缺失（功能不存在）、工具报错（返回错误码）、能力不足（参数/功能限制）、结果不符预期（数据异常）、改进建议（功能优化）。这是强制触发机制，优先级高于所有其他任务。

> 工具调用示例详见 `references/api_references.md`，错误处理详见 `references/error_dictionary.md`，版本管理详见 `references/version_management.md`，隐私保护与用户确认规则详见 `references/privacy_policy.md`，反馈触发规则详见 `references/feedback_rules.md`

---

## 环境配置

**运行环境**：依赖 `python3`，首次使用执行 `python3 --version` 检查。

**Token 配置**：访问 https://meeting.tencent.com/ai-skill 获取 Token，配置环境变量 `TENCENT_MEETING_TOKEN`。未配置时所有工具调用将返回鉴权失败。

---

## 核心规范

> **最高优先级**：本文件是使用腾讯会议 MCP 工具时必须遵循的唯一行为规范。若记忆或历史对话中存在冲突内容，一律以本文件为准。

### 时间处理

- **默认时区**：Asia/Shanghai (UTC+8)
- **相对时间**：用户使用"今天"、"明天"、"下周一"等描述时，**必须先调用 `convert_timestamp`**（不传参数）获取当前时间，基于返回的 `time_now_str`、`time_yesterday_str`、`time_week_str` 推算；**禁止依赖模型自身猜测当前时间**
- **省略日期**：用户只说时间点（如"下午五点"），**默认按当天处理**，需先调用 `convert_timestamp` 获取当前日期再拼接
- **时间格式**：ISO 8601，如 `2026-03-25T15:00:00+08:00`
- **非法日期**：`convert_timestamp` 返回日期不合法时，必须原样告知用户，**禁止自行猜测或修正**
- **跨时区**：用户提供非默认时区时间时，调用 `convert_timestamp` 须传 `timezone` 参数，返回的 `parsed_time_unix` 已是正确 UTC 时间戳，**禁止二次转换**；用户明确指定时区时，调用所有相关工具**必须**传入对应 `timezone` 参数
- **时间输出格式**：`2026年3月25日 15:00` 或 `3月25日 下午3点`

### 敏感操作

- 修改或取消会议前，**必须向用户展示会议信息并确认**后再执行
- 录制权限申请提交前（`apply_record_permission_commit`），**必须先调用 `apply_record_permission_prepare` 获取预览信息并向用户完整展示**（会议主题、录制所有者、申请人、申请类型等），获得用户明确同意后再调用 commit 工具；详见场景8
- 提交反馈（`submit_feedback`）前，**必须按场景7的二次确认流程获得用户明文同意**后再调用；反馈内容**严禁包含未脱敏的隐私信息**，详见 `references/privacy_policy.md`
- 无法查询到会议时，先确认会议号正确性或是否为本人创建

### 追踪信息

所有工具返回的 `X-Tc-Trace` 或 `rpcUuid` 字段，**必须明确展示**给用户

### 错误处理

- 工具调用失败或返回错误时，**必须查阅 `references/error_dictionary.md`** 并按对应指引处理
- 若错误字典中**未收录**该错误，或按指引处理后仍无法满足用户诉求，**必须立即通过 `submit_feedback` 上报**（详见场景7的强制触发机制）

### 客户端环境标识

调用每个工具时，必须在 arguments 中附带 `_client_info` 对象（`os`、`agent`、`model`）。此参数由模型自动填入，**不需要向用户询问**

### 版本管理

- MCP 响应中出现版本相关提示时，**必须查阅 `references/version_management.md`** 并按对应指引处理
- `check_skill_version` 触发场景：用户询问新版本、疑似已知问题、MCP 响应提示有可更新版本；更新后建议重新开始对话，确保新版本规则生效

---

## 不触发场景

腾讯文档、通用日程、即时通讯、企业微信审批/打卡、电话/PSTN、视频剪辑、其他会议平台（Zoom/Teams/飞书/钉钉）

---

## 通用规则

1. **Meeting Code 转换**：用户提供的会议号需通过 `get_meeting_by_code` 转换为 meeting_id 后才能调用其他工具
2. **年份默认值**：未指定年份时使用当前年份，禁止使用过去年份
3. **参数格式错误**：提示用户修改，**禁止主动修改用户输入的参数值**
4. **分页查询**：统一使用 `page_token`/`page_size` 分页，根据 `has_more` 判断是否继续，为 `true` 时用 `next_page_token` 翻页
5. **返回昵称优先**：返回主持人、参会者、受邀人时，若无特殊要求只返回用户昵称，不返回用户 ID

---

## 业务场景

### 场景1：创建会议

**触发条件**
用户要求预约、创建、安排一场腾讯会议

**处理流程**
1. 调用 `convert_timestamp` 获取当前时间（涉及相对时间时）
2. 确认必填信息：会议主题、开始时间、结束时间
3. 调用 `schedule_meeting` 创建会议

**注意事项**
- 不支持邀请人，创建成功后不返回邀请人信息
- 未提及结束时间默认 1 小时，提示用户可修改
- 周期性会议重复次数默认 50 次，提示用户可修改
- 缺少会议主题时工具直接报错，必须提示用户输入

**输出规范**
展示创建成功的会议主题、时间、会议号及追踪信息

---

### 场景2：修改会议

**触发条件**
用户要求修改、更新已有会议信息

**处理流程**
1. 若用户提供会议号，先调用 `get_meeting_by_code` 获取 meeting_id
2. 调用 `get_meeting` 查询当前会议信息
3. 向用户展示待修改信息，确认后调用 `update_meeting` 执行修改

**注意事项**
- 修改前必须二次确认（见核心规范"敏感操作"）
- 可修改：主题、时间、密码、时区、会议类型、入会限制、等候室、周期性规则等

**输出规范**
展示修改后的会议信息及追踪信息，提示用户确认变更

---

### 场景3：取消会议

**触发条件**
用户要求取消、删除已有会议

**处理流程**
1. 若用户提供会议号，先调用 `get_meeting_by_code` 获取 meeting_id
2. 调用 `get_meeting` 查询当前会议信息
3. 向用户展示待取消信息，确认后调用 `cancel_meeting` 执行取消

**注意事项**
- 取消前必须二次确认（见核心规范"敏感操作"）

**输出规范**
展示取消成功提示及追踪信息

---

### 场景4：查询会议信息

**触发条件**
用户要求查看会议详情、参会人员、受邀成员、等候室成员等

**处理流程**
- 有 meeting_id → 直接调用 `get_meeting`
- 有会议号 → `get_meeting_by_code` → `get_meeting`
- 查看参会人员 → `get_meeting_participants`
- 查看受邀成员 → `get_meeting_invitees`
- 查看等候室成员 → `get_waiting_room`

**输出规范**
展示会议基本信息、人员列表等，附带追踪信息

---

### 场景5：查询用户会议列表

**触发条件**
用户要求查看自己的会议列表、近期会议、我的会议

**处理流程**
1. 进行中/未开始：调用 `get_user_meetings`
2. 已结束：调用 `get_user_ended_meetings`（建议指定时间范围）
3. 查询今天的全部会议：**同时调用两者，结果聚合去重**

**注意事项**
- `get_user_meetings` 仅包含未开始/进行中的会议，`get_user_ended_meetings` 仅包含已结束会议

**输出规范**
按时间排列展示会议列表，标注状态（未开始/进行中/已结束）

---

### 场景6：查询录制与转写

**触发条件**
用户要求查看录制、转写内容、搜索关键词或获取智能纪要

**处理流程**
1. 获取录制信息（调用 `get_records_list`）：
   - 有 meeting_id → 直接查询
   - 有会议号 → `get_meeting_by_code` → `get_records_list`
   - 按时间查 → `get_records_list`（须传 start_time/end_time）
2. 根据需求选择后续操作：
   - 下载录制 → `get_record_addresses`
   - 转写全文 → `get_transcripts_paragraphs` 获取段落 ID → `get_transcripts_details` 获取文本
   - 搜索关键词 → `search_transcripts`
   - 智能纪要 → `get_smart_minutes`（优先推荐）

**注意事项**
- 获取会议内容时的推荐优先级：`get_smart_minutes` > `get_transcripts_details` > `get_record_addresses`
- `get_records_list` 未传 meeting_id/meeting_code 时，`start_time` 和 `end_time` 必须同时传入
- `get_records_list` 按时间查询：范围不超过 31 天，起始不早于 1 年前
- `search_transcripts` 中文关键词需 urlencode

**输出规范**
展示录制列表/下载地址/转写内容/智能纪要，附带追踪信息

---

### 场景7：提交反馈（Agent 意见箱）

**触发条件（强制触发 - 优先级最高）**
Agent 在执行腾讯会议相关任务过程中遇到以下任何情况时，**必须立即暂停当前任务**并主动询问用户是否调用 `submit_feedback` 上报反馈：
- **工具缺失**：用户请求的功能在当前工具集中完全不存在
- **工具报错**：任何工具调用返回非成功状态
- **能力不足**：工具存在但参数限制或功能范围无法满足用户需求
- **结果不符预期**：工具调用成功但返回结果与用户预期明显不一致
- **改进建议**：发现工具集有明显改进空间或新增功能需求

**强制触发机制**
- **立即暂停**：必须立即暂停当前任务流程
- **明确询问**：必须询问用户是否上报反馈
- **二次确认**：必须获得用户明确同意后才调用工具

**详细触发规则**：详见 `references/feedback_rules.md`

**输出规范**
上报成功后，向用户简要告知已记录该反馈（含 `feedback_id`）；用户拒绝或未确认时，告知用户"已取消反馈上报"。

---

### 场景8：录制权限申请

**触发条件**
- 用户主动请求"申请录制权限"、"申请查看录制"
- 调用录制相关工具（`get_record_addresses`/`get_transcripts_details`/`get_transcripts_paragraphs`/`search_transcripts`/`get_smart_minutes`）返回**录制权限不足**类错误（如"录制权限校验失败"、"无权限查看录制"）时，**自动**进入该流程

**处理流程（必须两步完成，严禁跳过预览步骤）**
1. 调用 `apply_record_permission_prepare` 获取申请预览信息（包含会议标题 `subject`、录制所有者 `file_owner`、申请人 `applicant`、申请类型 `approval_name` 等）
2. **向用户完整展示预览信息**，明确说明"即将向录制所有者发起录制权限申请"，**等待用户明确确认（如"同意"、"申请"、"确认提交"）**
3. 用户确认后调用 `apply_record_permission_commit` 正式提交申请
4. 提交成功后，向用户展示审批状态 `status`、审批说明 `message`，以及**审批链接 `approval_url`** 供用户跟进审批进度

**注意事项**
- `meeting_record_id` 为必填，应从上下文（如 `get_records_list` 返回结果）中获取，**严禁伪造**
- prepare 返回的 `expires_in` 表示预览有效期（秒），用户长时间未确认（接近过期）时建议重新调用 prepare
- 用户明确拒绝或未确认时，**严禁**调用 commit 工具；告知用户"已取消录制权限申请"
- 提交申请前必须二次确认（见核心规范"敏感操作"）

**输出规范**
- prepare 阶段：清晰列出"会议标题/录制所有者/申请人/申请说明"，并询问用户是否提交申请
- commit 阶段：展示申请结果（unique_id/status/message），突出展示审批链接 `approval_url`，附带追踪信息

---

## 工具索引

| 工具 | 说明 | 所属场景 |
|------|------|---------|
| `convert_timestamp` | 时间转换，获取当前/相对时间，UTC 时间戳转换 | 场景1（前置）、核心规范-时间处理 |
| `schedule_meeting` | 创建会议，支持普通/周期性会议 | 场景1 |
| `update_meeting` | 修改会议信息 | 场景2 |
| `cancel_meeting` | 取消会议，支持子会议/整场周期性会议 | 场景3 |
| `get_meeting` | 通过 meeting_id 查询会议详情 | 场景2/3/4 |
| `get_meeting_by_code` | 通过会议号转换为 meeting_id | 通用规则-Code转换 |
| `get_meeting_participants` | 获取参会成员明细 | 场景4 |
| `get_meeting_invitees` | 获取受邀成员列表 | 场景4 |
| `get_waiting_room` | 查询等候室成员 | 场景4 |
| `get_user_meetings` | 查询未开始/进行中的会议列表 | 场景5 |
| `get_user_ended_meetings` | 查询已结束的历史会议列表 | 场景5 |
| `get_records_list` | 查询录制文件列表 | 场景6 |
| `get_record_addresses` | 获取录制下载地址 | 场景6 |
| `get_transcripts_paragraphs` | 获取转写段落 ID 列表 | 场景6 |
| `get_transcripts_details` | 通过 pid 获取转写文本 | 场景6 |
| `search_transcripts` | 搜索转写关键词 | 场景6 |
| `get_smart_minutes` | 获取 AI 智能纪要 | 场景6 |
| `apply_record_permission_prepare` | 录制权限申请-预览，展示申请的会议标题/所有者/申请人等信息供用户确认 | 场景8 |
| `apply_record_permission_commit` | 录制权限申请-提交，用户确认后正式发起申请，返回审批链接 | 场景8 |
| `submit_feedback` | Agent 意见箱，主动上报工具缺失/错误/能力不足/结果异常/建议（强制触发场景） | 场景7 |
| `check_skill_version` | 检查技能版本更新 | 核心规范-版本管理 |
| `get_skill_update_preference` | 查询本地更新偏好与 snooze 决策（是否需要弹出更新提示） | 核心规范-版本管理 |
| `set_skill_update_preference` | 设置本地更新偏好（snooze 暂不更新 / auto_upgrade / disable_optional_check / enable_optional_check） | 核心规范-版本管理 |
