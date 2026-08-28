# Heartbeat State Template

Use this file as the baseline for `~/self-improving/heartbeat-state.md`.
It stores only lightweight run markers and maintenance notes.

```markdown
# Self-Improving Heartbeat State

last_heartbeat_started_at: never
last_reviewed_change_at: never
last_heartbeat_result: never

## Last actions
- none yet
```

## Rules

- update `last_heartbeat_started_at` at the beginning of every heartbeat
- update `last_reviewed_change_at` only after a clean review of changed files
- keep `last_actions` short and factual
- never turn this file into another memory log
