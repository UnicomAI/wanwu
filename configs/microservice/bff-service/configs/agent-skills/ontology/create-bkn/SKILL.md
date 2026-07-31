---
name: create-bkn
description: >-
  Guides creation of BKN (Business Knowledge Network) definition files following v2.0.1 spec.
  Covers network, object_type, relation_type, action_type, concept_group.
  Use when creating knowledge networks, BKN files, object types, relation types, action types,
  concept groups, or when user asks to model business knowledge in BKN format.
  When ontology-core is also loaded, use it to run ontology CLI (bkn push) after files exist.
---

# Create BKN

Generate well-formed BKN directories (Markdown + YAML frontmatter) per v2.0.1.

## Works with ontology-core

**create-bkn** authors the `.bkn` tree; **ontology-core** runs `ontology bkn push` / `pull` after files exist.

## What is BKN

BKN is Markdown + YAML frontmatter for schema; one file per definition under typed subfolders. Details (sections, required tables, types) live in [references/SPECIFICATION.llm.md](references/SPECIFICATION.llm.md).

## Directory layout

```
{network_dir}/
├── SKILL.md
├── network.bkn
├── CHECKSUM                 # optional; SDK may generate
├── object_types/
├── relation_types/
├── action_types/
├── concept_groups/
└── data/                    # optional CSV instance data
```

## Workflow

1. **Gather requirements** — objects, relations, actions, optional concept groups
2. **Read spec** — [references/SPECIFICATION.llm.md](references/SPECIFICATION.llm.md) (format rules, sections, frontmatter types)
3. **Pick templates** — copy/adapt from [assets/templates/](assets/templates/) (`network_type.bkn.template`, `object_type.bkn.template`, …)
4. **Create `network.bkn`** — root file; align with Network Overview
   - **MUST**: generate a fresh **UUID v4** locally (e.g. Python `uuid.uuid4()`) and write it as the `id` field in frontmatter at file creation time. Never leave `id` empty, `null`, `~`, or absent — `ontology bkn validate` / `push` both require a non-empty string id, and the bkn-creator flow does **not** call `ontology bkn create` to acquire a server-assigned id.
   - The locally generated UUID is the final `kn_id`; any other `.bkn` file that references the network id (e.g. `network_id` in `object_types/*.bkn`) must reuse the same UUID.
5. **Create `object_types/*.bkn`** — one file per object, `{id}.bkn`
6. **Create `relation_types/*.bkn`** — one file per relation
7. **Create `action_types/*.bkn`** — **optional, may be empty**. One file per genuine write-operation on instances. Decide per the rules below; if none apply, create no files and leave `action_types/` and the Network Overview ActionType list empty (keep them consistent).
   - **Build an action** when the PRD contains an explicit write-operation (create/update/delete of instances) with an identifiable trigger condition or a bound tool.
   - **Create none** when the PRD is read-only (query / monitor / analysis / stats / trace) — those belong to object-type query / subgraph / metric / semantic search, not ActionType.
   - **Ambiguous cases** (e.g. whether "approval" is a write-action vs. a status transition) — surface to the user for confirmation; never silently build or silently skip.
8. **Create `concept_groups/*.bkn`** — optional
9. **Update `network.bkn`** — list all IDs in Network Overview
10. **Add root `SKILL.md` in the BKN directory** — same folder as `network.bkn` (this is **not** the create-bkn skill file); agent-facing guide for that network (see [Delivered BKN: root SKILL.md](#delivered-bkn-root-skillmd))
11. **Review (MUST)** — cross-check [Validation checklist](#validation-checklist) and [Business rules placement](#business-rules-placement); fix IDs, cross-refs, headings
    - 特别核对 `action_types/*.bkn` 每个文件 frontmatter 是否含 `action_type: add|modify|delete` 这一行（**位置在 frontmatter，不在 markdown body 的 Bound Object 表**），遇到查询/监控/追溯/校验等只读语义，**删掉对应 ActionType 文件并在 `network.bkn` 的 Network Overview 同步移除该 id**
12. **Validate (MUST)** — `ontology bkn validate <dir>` (see [Validation](#validation))
13. **Import** (optional) — `ontology bkn push <dir>`

## Import (ontology CLI)

**无需安装**。`ontology` 已内置在执行环境，直接调用即可。

- **BKN validation** — If workflow step 12 (`ontology bkn validate <dir>`) **already succeeded** for this directory, **do not** repeat validate before `push` unless you changed `.bkn` files. If you have **not** validated yet, run `validate` before `push`.

```bash
ontology bkn push <dir> [-bd <business-domain>]
```

`-bd` / `--biz-domain` is optional. If you omit it, the CLI resolves the business domain automatically.

Export: `ontology bkn pull <kn-id> [<dir>]`. More subcommands: `ontology bkn --help` (see ontology-core skill if loaded).

## Validation

`ontology bkn validate <dir>` — must pass before delivery or upload. It loads `network.bkn` and sibling `.bkn` files. Success prints counts; on failure fix `.bkn` files and re-run.

## Per-type reference

| Kind | Spec (section) | Template | Example (k8s) |
|------|------------------|----------|---------------|
| Network | `knowledge_network` in spec | [assets/templates/network_type.bkn.template](assets/templates/network_type.bkn.template) | [references/examples/k8s-network/network.bkn](references/examples/k8s-network/network.bkn) |
| Object | `object_type` | [assets/templates/object_type.bkn.template](assets/templates/object_type.bkn.template) | [references/examples/k8s-network/object_types/pod.bkn](references/examples/k8s-network/object_types/pod.bkn) |
| Relation | `relation_type` | [assets/templates/relation_type.bkn.template](assets/templates/relation_type.bkn.template) | [references/examples/k8s-network/relation_types/pod_belongs_node.bkn](references/examples/k8s-network/relation_types/pod_belongs_node.bkn) |
| Action | `action_type` | [assets/templates/action_type.bkn.template](assets/templates/action_type.bkn.template) | [references/examples/k8s-network/action_types/restart_pod.bkn](references/examples/k8s-network/action_types/restart_pod.bkn) |
| Concept group | `concept_group` | [assets/templates/concept_group.bkn.template](assets/templates/concept_group.bkn.template) | [references/examples/k8s-network/concept_groups/k8s.bkn](references/examples/k8s-network/concept_groups/k8s.bkn) |

Full rules and optional sections: [references/SPECIFICATION.llm.md](references/SPECIFICATION.llm.md).

## Naming conventions

- **ID**: lowercase, digits, underscores; **file**: `{id}.bkn` under the matching folder
- **Headings**: `#` network title, `##` type block, `###` section, `####` logic property
- **Frontmatter**: at least `type`, `id`, `name` (see spec for each type)

## Business rules placement

Rules must sit in spec-defined places so import persists them. Full wording: [references/SPECIFICATION.llm.md](references/SPECIFICATION.llm.md#输出规则).

- **Network-level** — prose in `network.bkn` right after `# {title}` (before structured sections like `## Network Overview`)
- **Type-level** — prose in each type file after `## ObjectType:` / `## RelationType:` / … and **before** the first `###`; never in frontmatter
- **Property-level** — in **Data Properties** table **Description** column
- **No extra sections** — do not add Markdown outside the standard sections; parsers may drop unparsed content on import

## Validation checklist

- [ ] `network.bkn` at root; frontmatter matches spec
- [ ] Every `.bkn` has valid YAML frontmatter (`type`, `id`, `name`)
- [ ] Files live under folders matching `type` (`object_types/`, `relation_types/`, …); filename = `{id}.bkn`
- [ ] Network Overview lists **all** definition IDs — no missing/extra
- [ ] Relations/actions reference existing object-type IDs; concept groups list only existing objects
- [ ] Parameter binding `Source` ∈ `property` | `input` | `const`; YAML blocks (e.g. trigger) parse
- [ ] **每个 `action_types/*.bkn` 的 frontmatter 都包含 `action_type:` 字段，值 ∈ {`add`, `modify`, `delete`}**（注意：该字段在 **frontmatter**，不在 `### Bound Object` 表格里；缺失或值非法时 backend 返回 `BknBackend.ActionType.InvalidParameter`，错误细节里说的 `[create, update, delete]` 是 backend bug，实际白名单就是 `add/modify/delete`）
- [ ] **`network.bkn` 的 Network Overview ActionType 列表与 `action_types/*.bkn` 文件名一一对应**（多/少都会导致后端 schema 不一致；CLI 不校验，会静默放过）
- [ ] **没有把只读语义（查询/监控/追溯/校验）建为 ActionType**——这类应走 object-type query / subgraph / metric / semantic search
- [ ] Heading hierarchy has no skipped levels
- [ ] Business rules only in allowed places (see [Business rules placement](#business-rules-placement))

## Output rules

1. Emit raw `.bkn` content — do not wrap the whole file in a fenced `markdown` block
2. Reuse IDs consistently across relations/actions
3. IDs: lowercase + underscores; display text Chinese unless asked otherwise
4. Keep heading order per spec

## Examples

- [references/examples/k8s-network/](references/examples/k8s-network/) — modular sample (objects, relations, actions, concept group)

## Delivered BKN: root GUIDE.md

When you build a knowledge network directory `{network_dir}/`, add `{network_dir}/GUIDE.md` at the root (alongside `network.bkn`). Short overview + **index tables with file paths** (object | path | relation | path | action | path) so agents route to the right `.bkn` without scanning. Optional: topology sketch, usage scenarios. Example: [references/examples/k8s-network/GUIDE.md](references/examples/k8s-network/GUIDE.md).
