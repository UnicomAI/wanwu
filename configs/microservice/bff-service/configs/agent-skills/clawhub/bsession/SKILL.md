---
name: browser
description: Browser automation — setup the bsession environment, fetch info from a website (one-shot), create scripted automations (one-shot or recurring), or debug existing sessions. Works from any repo.
user-invocable: true
metadata: {"openclaw":{"requires":{"bins":["docker"]}}}
---

# /browser skill

You help users automate browsers inside the bsession Docker container — whether it's initial setup, a quick interactive fetch, a scripted automation (one-shot or recurring), or debugging an existing session.

**This is a global skill** — it works from any repo. bsession is installed at `~/.bsession/`, and the `bsession` CLI is on PATH.

## Resolve paths

Before doing anything, determine how to reach bsession. Check in this order:

1. `bsession` on PATH → use `bsession`
2. `~/.bsession/bsession` exists → use `~/.bsession/bsession`
3. `./bsession` in current directory → use `./bsession`
4. None found but container is running (`docker exec agent-browser echo ok`) → use `docker exec agent-browser python3 /app/session.py` as the CLI

Similarly, resolve workspace:
1. `~/.bsession/workspace/` exists → use it
2. `./workspace/` in current directory → use it
3. Ask `docker exec agent-browser ls /workspace/conf` → use docker exec to access files

Use these resolved paths for **all** commands throughout the session.

## Constants (defaults)

- **BSESSION_HOME**: `~/.bsession/` — where bsession source + docker-compose live
- **WORKSPACE**: `~/.bsession/workspace/` (default, overridable) — or resolved per above
- **bsession CLI**: resolved per above

## Routing

Parse the user's slash command arguments:

- **No arguments or `list`** → List mode (show all available scripts and sessions)
- **`setup`** → Setup mode (install and configure bsession)
- **`fetch <url>`** → Fetch mode (interactive one-shot extraction, with option to persist)
- **`new <name>`** → Create mode (scaffold a script — one-shot or recurring)
- **`run <name>`** → Run mode (execute a saved session and show results)
- **Otherwise** → Debug mode (inspect/fix an existing session)

## Pre-check (all modes except setup)

Before running any mode except setup, verify the container is running:

```bash
docker exec agent-browser echo ok 2>/dev/null
```

If this fails, tell the user to either:
- Run `/browser setup` for a fresh install, or
- Run `docker compose up -d` from the bsession project directory

---

## List mode (`/browser` or `/browser list`)

Show all available scripts, their status, and what they do.

### Step 1: Get session status

```bash
bsession list
```

### Step 2: Read script docstrings

For each `.py` file in `~/.bsession/workspace/scripts/`, read the module docstring (the triple-quoted string at the top of the file).

### Step 3: Read conf files

For each `.conf` file in `~/.bsession/workspace/conf/`, read the `[env]` section to show current configuration.

### Step 4: Present as a table

Display a summary like:

```
Session       Status    Type        Description
─────────────────────────────────────────────────────────────────
uscis         running   recurring   USCIS case status monitor
price-check   stopped   one-shot    Amazon product price scraper

Available commands:
  /browser <name>           debug a session
  /browser new <name>       create a new automation
  /browser fetch <url>      quick one-shot fetch
```

---

## Setup mode (`/browser setup`)

Run the install script:

```bash
bash ~/.openclaw/workspace/skills/browser/scripts/install.sh
```

Or with options:
```bash
bash ~/.openclaw/workspace/skills/browser/scripts/install.sh --workspace /path/to/workspace
bash ~/.openclaw/workspace/skills/browser/scripts/install.sh --vnc-password secret
bash ~/.openclaw/workspace/skills/browser/scripts/install.sh --repo https://github.com/gaxxx/bsession.git
```

Ask the user for custom options before running. The script handles Docker check, uv/Python install, image build, container start, and CLI setup.

---

## Fetch mode (`/browser fetch <url>`)

One-shot: open a URL, extract information, return it. No script, no conf file, no loop.

### Step 1: Find an available CDP port

```bash
docker exec agent-browser python3 -c "
import urllib.request
try:
    urllib.request.urlopen('http://localhost:9222/json/version', timeout=2)
    print('IN_USE')
except:
    print('FREE')
"
```

If 9222 is in use, try 9223, 9224, etc. Start a temporary Chrome on a free port:

```bash
docker exec agent-browser python3 -c "
import sys; sys.path.insert(0, '/app')
from lib.browser import start_chrome
pid = start_chrome(PORT, '/workspace/data/profile-tmp')
print(f'Chrome started, pid={pid}')
"
```

### Step 2: Navigate and extract

```bash
docker exec agent-browser agent-browser --cdp PORT open "URL"
sleep 5
docker exec agent-browser agent-browser --cdp PORT snapshot
```

Handle Cloudflare if detected:

```bash
docker exec agent-browser python3 -c "
import sys; sys.path.insert(0, '/app')
from lib.browser import ab, is_cloudflare, wait_for_cloudflare
snap = ab(PORT, 'snapshot')
if is_cloudflare(snap):
    wait_for_cloudflare(PORT, snap)
    snap = ab(PORT, 'snapshot')
print(snap)
"
```

### Step 3: Parse and interact

```bash
docker exec agent-browser agent-browser --cdp PORT fill REF "value"
docker exec agent-browser agent-browser --cdp PORT click REF
docker exec agent-browser agent-browser --cdp PORT snapshot
```

### Step 4: Return results

Parse the relevant information and present it cleanly.

### Step 5: Offer to persist

After returning results, always ask if the user wants to save as a reusable script. If yes, create a one-shot script + conf in `~/.bsession/workspace/`.

### Step 6: Cleanup

```bash
docker exec agent-browser python3 -c "
import sys; sys.path.insert(0, '/app')
from lib.browser import stop_chrome
stop_chrome(PORT)
"
```

---

## Create mode (`/browser new <name>`)

Ask the user:
1. What URL(s) to target
2. One-shot or recurring?
3. What to detect / extract
4. Where to send results (webhook, file, etc.)
5. Env vars needed

Then scaffold `~/.bsession/workspace/conf/<name>.conf` and `~/.bsession/workspace/scripts/<name>.py` following the conventions in the reference section below.

---

## Run mode (`/browser run <name>`)

1. Verify session exists: `bsession show <name>`
2. Run it: `bsession run <name>`
3. Wait and tail logs: `bsession logs <name> -n 50`
4. Present results. If failed, switch to debug mode.

---

## Debug mode (`/browser <session-id>`)

1. Gather state: `bsession list`, `bsession show <id>`, read logs and script
2. Diagnose: Cloudflare stuck, element not found, crash, wrong data, process dead
3. Fix the script or conf, then `bsession restart <id>`

---

## Script conventions

**Imports:**
```python
import os, re, sys, time
sys.path.insert(0, "/app")
from lib.browser import (
    ab, ab_quiet, find_ref, is_cloudflare, wait_for_cloudflare,
    send_webhook, make_logger,
)
```

**Config from env vars:**
```python
port = int(os.environ.get("CDP_PORT", 9222))
session_name = os.environ.get("SESSION_NAME", "<name>")
webhook_url = os.environ.get("N8N_WEBHOOK_URL", "")
check_interval = int(os.environ.get("CHECK_INTERVAL", 1800))
```

**Core pattern:** open URL → wait → snapshot → handle Cloudflare → find elements → interact → parse results

**One-shot:** execute and exit. **Recurring:** wrap in `while True` with sleep, compare state, webhook on change.

## Reference: lib/browser.py

- `ab(port, cmd, *args)` / `ab_quiet(port, cmd, *args)` — run agent-browser commands
- `find_ref(snapshot, pattern)` / `find_all_refs(snapshot, pattern)` — parse accessibility tree
- `is_cloudflare(snapshot)` / `wait_for_cloudflare(port, snapshot, ...)` — Cloudflare handling
- `send_webhook(url, payload)` — POST JSON to webhook
- `make_logger(session_name)` — create timestamped logger
