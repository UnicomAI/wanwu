# Heartbeat Rules

Use heartbeat to keep `~/self-improving/` organized without creating churn or losing data.

## Source of Truth

Keep the workspace `HEARTBEAT.md` snippet minimal.
Treat this file as the stable contract for self-improving heartbeat behavior.
Store mutable run state only in `~/self-improving/heartbeat-state.md`.

## Start of Every Heartbeat

1. Ensure `~/self-improving/heartbeat-state.md` exists.
2. Write `last_heartbeat_started_at` immediately in ISO 8601.
3. Read the previous `last_reviewed_change_at`.
4. Scan `~/self-improving/` for files changed after that moment, excluding `heartbeat-state.md` itself.

## If Nothing Changed

- Set `last_heartbeat_result: HEARTBEAT_OK`
- Append a short "no material change" note if you keep an action log
- Return `HEARTBEAT_OK`

## If Something Changed

Only do conservative organization:

- refresh `index.md` if counts or file references drift
- compact oversized files by merging duplicates or summarizing repetitive entries
- move clearly misplaced notes to the right namespace only when the target is unambiguous
- preserve confirmed rules and explicit corrections exactly
- update `last_reviewed_change_at` only after the review finishes cleanly

## Safety Rules

- Most heartbeat runs should do nothing
- Prefer append, summarize, or index fixes over large rewrites
- Never delete data, empty files, or overwrite uncertain text
- Never reorganize files outside `~/self-improving/`
- If scope is ambiguous, leave files untouched and record a suggested follow-up instead

## State Fields

Keep `~/self-improving/heartbeat-state.md` simple:

- `last_heartbeat_started_at`
- `last_reviewed_change_at`
- `last_heartbeat_result`
- `last_actions`

## Behavior Standard

Heartbeat exists to keep the memory system tidy and trustworthy.
If no rule is clearly violated, do nothing.
