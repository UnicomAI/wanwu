# 步骤 1：接收 KN（透传，无运行时决策）

在 **问数主流程第 1 步** 接收上游 [smart-data-analysis](../../smart-data-analysis/SKILL.md) 透传过来的 `kn_id`。**本 skill 不在运行时选 KN、不调 `bkn list/get` 枚举或决策。**

## 输入

| 字段 | 来源 |
|------|------|
| `kn_id` | smart-data-analysis 从其 SKILL.md「知识网络声明」表"问数"行读取后透传 |
| `accountId` | 同样由 smart-data-analysis 注入 |

## 校验与停止

- 若 `kn_id` 为空或仍为 `<填入...>` 占位：**立即停止**，回复用户「未在 smart-data-analysis/SKILL.md 配置问数 KN」，**不得**用 `bkn list` 找一个凑数、也不得编造。
- 若 `accountId` 为空：让 smart-data-analysis 向用户索取后再进入本 skill。

## 不做的事

- **不**在本 skill 内调 `bkn list` / `bkn get` 枚举或挑 KN（运行时 KN 决策已被取消；维护者一次性在 smart-data-analysis/SKILL.md 表中填好即可）。
- **不**对透传过来的 `kn_id` 做"业务/元数据/职责类型"二次判定——上游已经按分支选定。

## 通过后

把校验通过的 `kn_id` 与 `accountId` 透传给步骤 2（[schema-discovery.md](schema-discovery.md)）。
