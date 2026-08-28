# 版本管理指引

> 当 MCP 响应中出现版本相关提示时，必须按本文件指引处理。

---

## 总体规则

服务端（MCP Server）的版本检查仅做**提示**，不会拦截业务请求：

- 服务端会校验 `X-Skill-Version` 是否落后于最新版本；
- 若存在更新版本，会随业务响应一起返回提示信息；
- 不区分大版本/小版本，只要存在更新版本就会提示一次；
- 是否更新由 Skill 本地结合用户偏好决定。

---

## 收到“有新版本可更新”提示时的处理流程

按以下顺序处理：

1. **必须调用 `get_skill_update_preference` 工具查询本地偏好与 snooze 决策**（传入 `latest_version=hints 中的最新版本号`）：
   - **禁止自行读取或解析 `~/.tencent-meeting-mcp/update-state.json` 做时间戳运算**，必须以本工具返回为准。
   - 根据返回的 `shouldPrompt` 与 `reason` 决定下一步：
     - `shouldPrompt=false` 且 `reason=disabled`：用户已选择“不再提醒”，**静默忽略**。
     - `shouldPrompt=false` 且 `reason=auto_upgrade`：用户已预授权“默认自动更新”，可直接调用 `check_skill_version` 获取安装地址并执行安装，**无需再次询问**。
     - `shouldPrompt=false` 且 `reason=snoozed`：仍处于“暂不更新”静默期（可参考 `snooze.remainingSeconds` 查看剩余时间），**静默忽略**。
     - `shouldPrompt=true`（`reason ∈ {no_snooze, expired, version_changed}`）：进入下一步询问用户。
2. **需要询问时向用户展示选项**（带上当前版本与最新版本号）：
   - 立即更新
   - 暂不更新（本次跳过）
   - 以后自动更新
   - 不再提醒
3. **根据用户选择执行**（详见下一节）。

---

## 用户偏好的「四态互斥」约束

用户的版本更新偏好任意时刻仅处于以下其中一个状态，写入函数会自动维护互斥关系，模型无需感知：

| 状态 | 含义 | `optionalUpdateCheck` | `autoUpgrade` | `snooze` |
|---|---|---|---|---|
| A | 默认询问 | `true` | `false` | `null` |
| B | 暂不更新 | `true` | `false` | `{...}` |
| C | 自动更新 | `true` | `true` | `null` |
| D | 永不更新 | `false` | `false` | `null` |

**最近一次操作生效**：任意 `set_skill_update_preference` 写入入口在切换到目标状态时，会同步清空其它互斥字段，
即用户最近一次「暂不更新 / 自动更新 / 永不更新 / 恢复提醒」的意图始终压过此前的偏好，
不会出现「snooze + autoUpgrade」「snooze + 永不更新」等不可达组合。

---

## 用户选择对应的本地操作

| 用户选择 | 动作 | 切换到 |
|---|---|---|
| 立即更新 | 调用 `check_skill_version`（建议传 `output_format=json`）获取安装地址并执行安装；安装完成后提示用户重新开始新对话 | — |
| 暂不更新 | 调用 `set_skill_update_preference`，`action=snooze`、`version=最新版本`（静默时长由本地脚本决定，模型无需关心） | B |
| 以后自动更新 | 调用 `set_skill_update_preference`，`action=auto_upgrade` | C |
| 不再提醒 | 调用 `set_skill_update_preference`，`action=disable_optional_check` | D |
| 希望恢复提醒 | 调用 `set_skill_update_preference`，`action=enable_optional_check` | A |

> `set_skill_update_preference` 仅修改当前设备的本地偏好，不会影响其他设备，也不会上报到 MCP Server。
> `get_skill_update_preference` 是唯一授权的“是否提示”决策入口，模型不得跳过该工具自行判断 snooze 是否过期。

---

## 禁止行为

- **禁止自行修改请求头中的 `X-Skill-Version`** 来绕过检查
- **禁止跳过 `check_skill_version` 工具调用**，直接使用响应中的链接（链接可能过期）
- **禁止在用户未确认的情况下自动执行更新**（`autoUpgrade=true` 场景除外，那是用户明确预授权过的）
- **禁止重复打扰用户**：用户已设置 `disable_optional_check` 或处于 `snooze` 静默期内时，**必须保持静默**
- **禁止自行解析 `update-state.json` 做时间戳判断**：是否提示必须以 `get_skill_update_preference` 工具返回为准
