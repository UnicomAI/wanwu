---
name: bkn-creator
description: >-
  【首选技能】凡涉及知识网络的任何操作，优先使用 bkn-creator，不要直接触发 create-bkn。
  bkn-creator 是 BKN 全生命周期管理器，负责流程路由识别、阶段门禁、子流程编排与结果回执。
  覆盖新增、查找、更新、删除（CRUD）并采用渐进式执行。
  触发词：创建知识网络、新建BKN、BKN建模、本体建模、对象类、关系类、动作类、
  概念组、BKN文件、BKN push、BKN pull、对象类绑定、关系类映射、
  对象类提取、关系类提取、实体关系抽取、
  知识网络查询、知识网络更新、知识网络删除。
  不应触发：仅查询平台功能、Agent 对话、Vega/Catalog 操作、
  健康巡检、数据源连接器等非 BKN 生命周期操作时，应由 ontology-core 等其他技能处理。
---

# BKN 编排器（CRUD + 渐进式）

**角色定位**：流程编排器 + 生命周期管理器。  
**Never** 直接执行 `ontology` CLI；所有执行均委托子技能完成。

## 职能矩阵（谁做什么）

| 组件 | 输入 | 输出 | 禁止事项 |
|------|------|------|----------|
| **bkn-creator** | 用户意图、上下文、各子流程结果 | 阶段推进决策、确认门禁、最终回执 | Never 直接执行 CLI；Never 绕过确认门禁 |
| **ontology-core** | bkn-creator 的 CLI 委托 | 命令执行结果与回执 | Never 跳过 bkn-creator 直接接管流程 |
| **create-bkn** | 已确认的建模清单 | 标准 `.bkn` 文件树 | Never 改写流程门禁 |
| **bkn-modeling-advisor** | 模糊业务意图、领域上下文 | 结构化建模清单 | Never 直接推进阶段三/四/五 |

## 流程路由（第一层）

| 意图 | 关键词 | 路由目标 |
|------|--------|----------|
| **新增（Create）** | 创建、新建、建立、搭建、本体建模 | `./FLOW_CREATE.md` |
| **提取（Extract）** | 提取对象类、提取关系类、抽取实体关系、对象关系梳理 | `./FLOW_EXTRACT_TYPES.md` |
| **查找（Read）** | 查询、搜索、列出、有哪些、看看 | `./FLOW_READ.md` |
| **更新（Update）** | 修改、编辑、更新、调整、改一下 | `./FLOW_UPDATE.md` |
| **删除（Delete）** | 删除、清除、移除、去掉 | `./FLOW_DELETE.md` |

MUST 回显流程路由识别结果并等待用户确认后进入对应流程。意图不明确时先澄清。
MUST 在"流程路由确认（验证阶段）"通过前，Never 进入任何 `FLOW_*` 的阶段化执行与对用户阶段回显（包括"阶段一/Step 1"等内容）。

## 渐进式执行协议（第二层）

所有 CRUD 流程统一遵循：

`discover -> preview -> confirm -> execute -> verify -> report`

完整确认语义、门禁规则、全局硬约束见：`./COMMON_RULES.md`。

## 子流程调用规则（第三层）

仅当"流程路由确认"通过后，才可读取并执行对应 `FLOW_*.md` 的阶段内容。

- 新增：读取 `./FLOW_CREATE.md`
- 提取：读取 `./FLOW_EXTRACT_TYPES.md`
- 查找：读取 `./FLOW_READ.md`
- 更新：读取 `./FLOW_UPDATE.md`
- 删除：读取 `./FLOW_DELETE.md`

## 顶层约束（路由阶段即生效）

以下约束在加载任何 `FLOW_*.md` 之前即生效；细则统一以 `./COMMON_RULES.md` 为准：

- MUST：流程路由确认未通过时，仅允许输出“流程路由识别结果 + 确认请求”
- MUST：任何写操作必须先完成确认步骤
- MUST：凡需执行 CLI，统一委托 `../ontology-core/SKILL.md`
- MUST：对用户输出使用专业文本，不使用 emoji

## 统一目录与回执（顶层强制）

目录、落盘、归档回执规则统一以 `./COMMON_RULES.md` 与 `./FLOW_CREATE.md` 为准，不在本文件重复定义。

## 流程衔接

- 提取流程（Extract）的结果可作为新增流程（Create）的快速输入：若用户在完成提取后希望创建知识网络，可将提取清单直接带入新增流程，跳过阶段一的重复提取，从清单确认（A2/B2/C1）开始
- 查找流程（Read）完成后，若用户希望修改或删除结果，可直接发起流程路由切换到更新或删除流程
- 跨流程切换时仍需经过流程路由确认门禁

## 失败回退

- 输入不足：一次只追问最高优先级字段，不批量追问
- 工具调用失败：停在当前阶段，向用户报错并等待指令，Never 自动重试或跳过
- 结果不确定：标注低置信度，不得自动推进到下一阶段

## 示例

### 正例：新增流程

用户："根据这份 PRD 创建知识网络"  
执行：流程路由识别为新增 -> 路由 `FLOW_CREATE` -> 按门禁逐阶段推进。

### 反例：未确认推进

助手："是否确认删除？"  
用户："先这样，你继续说。"  
处理：必须停在流程路由确认门禁，不得执行删除。

## 上下文资源

- `./COMMON_RULES.md`
- `./FLOW_CREATE.md`
- `./FLOW_EXTRACT_TYPES.md`
- `./FLOW_READ.md`
- `./FLOW_UPDATE.md`
- `./FLOW_DELETE.md`
- `./references/DOMAIN_ROUTING.md`
- `./references/common/generic_extraction.md`
- `../ontology-core/SKILL.md`
- `../create-bkn/SKILL.md`
- `../bkn-modeling-advisor/SKILL.md`
