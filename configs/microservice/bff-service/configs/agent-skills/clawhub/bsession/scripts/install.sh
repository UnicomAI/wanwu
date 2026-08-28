#!/usr/bin/env bash
# Install and configure bsession browser automation environment.
#
# Installs bsession to ~/.bsession/ so it's available globally,
# independent of any specific project repo.
#
# What this script does:
#   1. Ensures Docker is available
#   2. Installs uv (Python runtime manager) if missing
#   3. Installs Python 3.12 via uv
#   4. Clones/copies bsession source to ~/.bsession/
#   5. Builds the agent-browser Docker image
#   6. Starts the container
#   7. Sets up workspace directories
#   8. Makes bsession CLI available on PATH
#   9. Verifies the full stack
#
# Usage:
#   bash install.sh [options]
#
# Options:
#   --repo <url>         Git repo to clone bsession from (default: use local source)
#   --workspace <path>   Custom workspace directory (default: ~/.bsession/workspace)
#   --vnc-password <pw>  Set a VNC password (default: none)
#   --no-start           Build only, don't start the container

set -euo pipefail

# ── Defaults ──────────────────────────────────────────────────────────
BSESSION_HOME="$HOME/.bsession"
WORKSPACE_DIR=""
VNC_PASSWORD=""
NO_START=false
REPO_URL=""

# Where this script lives (inside the skill or the repo)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── Parse args ────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo)         REPO_URL="$2"; shift 2 ;;
        --workspace)    WORKSPACE_DIR="$2"; shift 2 ;;
        --vnc-password) VNC_PASSWORD="$2"; shift 2 ;;
        --no-start)     NO_START=true; shift ;;
        *)              echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

# ── Helpers ───────────────────────────────────────────────────────────
info()  { printf "\033[32m[+]\033[0m %s\n" "$*"; }
warn()  { printf "\033[33m[!]\033[0m %s\n" "$*"; }
fail()  { printf "\033[31m[x]\033[0m %s\n" "$*"; exit 1; }
check() { command -v "$1" &>/dev/null; }

# ── Step 1: Check Docker ─────────────────────────────────────────────
info "Checking Docker..."
if ! check docker; then
    fail "Docker not found. Install Docker Desktop (macOS) or docker-ce (Linux) first."
fi
if ! docker info &>/dev/null; then
    fail "Docker daemon not running. Start Docker Desktop or the docker service."
fi
info "Docker OK."

# ── Step 2: Install uv + Python ──────────────────────────────────────
info "Checking uv..."
if ! check uv; then
    info "Installing uv..."
    curl -LsSf https://astral.sh/uv/install.sh | sh
    export PATH="$HOME/.local/bin:$PATH"
    if ! check uv; then
        fail "uv installed but not on PATH. Add ~/.local/bin to your PATH and rerun."
    fi
fi
info "uv OK: $(uv --version)"

info "Ensuring Python 3.12 via uv..."
uv python install 3.12 2>/dev/null || true
PYTHON_BIN="$(uv python find 3.12 2>/dev/null || echo "")"
if [[ -z "$PYTHON_BIN" ]]; then
    warn "Could not find Python 3.12 via uv. Host Python not required for Docker mode, continuing."
else
    info "Python OK: $PYTHON_BIN"
fi

# ── Step 3: Get bsession source into ~/.bsession/ ────────────────────
REQUIRED_FILES="Dockerfile docker-compose.yml session.py entrypoint.sh bsession lib/browser.py"

if [[ -n "$REPO_URL" ]]; then
    # Clone from remote
    info "Cloning bsession from $REPO_URL..."
    if [[ -d "$BSESSION_HOME" ]]; then
        info "$BSESSION_HOME already exists — pulling latest..."
        cd "$BSESSION_HOME"
        git pull --ff-only || warn "git pull failed, using existing files."
    else
        git clone "$REPO_URL" "$BSESSION_HOME"
    fi
else
    # Try to find source files: check if we're inside the bsession repo
    SOURCE_DIR=""

    # Check if the skill is inside a bsession repo (.claude/skills/browser/scripts/)
    CANDIDATE="$(cd "$SCRIPT_DIR/../../../.." 2>/dev/null && pwd)"
    if [[ -f "$CANDIDATE/Dockerfile" && -f "$CANDIDATE/session.py" ]]; then
        SOURCE_DIR="$CANDIDATE"
    fi

    # Check current working directory
    if [[ -z "$SOURCE_DIR" && -f "./Dockerfile" && -f "./session.py" ]]; then
        SOURCE_DIR="$(pwd)"
    fi

    if [[ -z "$SOURCE_DIR" ]]; then
        fail "Cannot find bsession source files. Either:
  - Run this script from the bsession repo directory, or
  - Use --repo <git-url> to clone it"
    fi

    if [[ "$SOURCE_DIR" == "$BSESSION_HOME" ]]; then
        info "Already at $BSESSION_HOME, skipping copy."
    else
        info "Copying bsession from $SOURCE_DIR to $BSESSION_HOME..."
        mkdir -p "$BSESSION_HOME"
        # Copy core files (not workspace data — that's runtime)
        for item in Dockerfile docker-compose.yml session.py entrypoint.sh bsession lib .gitignore; do
            if [[ -e "$SOURCE_DIR/$item" ]]; then
                cp -r "$SOURCE_DIR/$item" "$BSESSION_HOME/"
            fi
        done
        info "Source files copied."
    fi
fi

# Verify files
cd "$BSESSION_HOME"
for f in $REQUIRED_FILES; do
    [[ -f "$f" ]] || fail "Missing required file: $BSESSION_HOME/$f"
done
info "Project files OK at $BSESSION_HOME"

# ── Step 4: Configure workspace ──────────────────────────────────────
if [[ -z "$WORKSPACE_DIR" ]]; then
    WORKSPACE_DIR="$BSESSION_HOME/workspace"
fi

mkdir -p "$WORKSPACE_DIR"/{conf,data,scripts}

# If workspace is not the default location, create override
if [[ "$WORKSPACE_DIR" != "$BSESSION_HOME/workspace" ]]; then
    cat > "$BSESSION_HOME/docker-compose.override.yml" <<YAML
services:
  agent-browser:
    volumes:
      - ${WORKSPACE_DIR}:/workspace
YAML
    info "Created docker-compose.override.yml for custom workspace: $WORKSPACE_DIR"
else
    # Remove stale override if workspace is default
    rm -f "$BSESSION_HOME/docker-compose.override.yml"
fi
info "Workspace directories ready: $WORKSPACE_DIR"

# ── Step 5: Configure .env ───────────────────────────────────────────
if [[ ! -f "$BSESSION_HOME/.env" ]]; then
    touch "$BSESSION_HOME/.env"
fi

if [[ -n "$VNC_PASSWORD" ]]; then
    if grep -q '^VNC_PASSWORD=' "$BSESSION_HOME/.env"; then
        sed -i.bak "s/^VNC_PASSWORD=.*/VNC_PASSWORD=$VNC_PASSWORD/" "$BSESSION_HOME/.env" && rm -f "$BSESSION_HOME/.env.bak"
    else
        echo "VNC_PASSWORD=$VNC_PASSWORD" >> "$BSESSION_HOME/.env"
    fi
    info "VNC password configured."
else
    info "No VNC password set (open access)."
fi

# ── Step 6: Build Docker image ───────────────────────────────────────
info "Building agent-browser Docker image..."
cd "$BSESSION_HOME"
docker compose build
info "Image built."

# ── Step 7: Start container ──────────────────────────────────────────
if [[ "$NO_START" == true ]]; then
    info "Skipping container start (--no-start)."
else
    # Stop any existing agent-browser container (may be from a different project dir)
    if docker inspect agent-browser &>/dev/null 2>&1; then
        info "Stopping existing agent-browser container..."
        docker rm -f agent-browser &>/dev/null || true
        # Clean up orphan networks
        docker compose down --remove-orphans &>/dev/null 2>&1 || true
    fi

    info "Starting container..."
    docker compose up -d
    info "Container started."

    info "Waiting for container to be ready..."
    for i in $(seq 1 15); do
        if docker exec agent-browser echo ok &>/dev/null; then
            break
        fi
        sleep 1
    done

    if ! docker exec agent-browser echo ok &>/dev/null; then
        fail "Container not responding after 15s. Check: docker compose -f $BSESSION_HOME/docker-compose.yml logs"
    fi
fi

# ── Step 8: Set up bsession CLI ──────────────────────────────────────
chmod +x "$BSESSION_HOME/bsession"

# Determine bin directory
BIN_DIR=""
if [[ -d "$HOME/.local/bin" ]]; then
    BIN_DIR="$HOME/.local/bin"
elif [[ -d "/usr/local/bin" && -w "/usr/local/bin" ]]; then
    BIN_DIR="/usr/local/bin"
else
    mkdir -p "$HOME/.local/bin"
    BIN_DIR="$HOME/.local/bin"
fi

LINK_PATH="$BIN_DIR/bsession"
if [[ -L "$LINK_PATH" || -f "$LINK_PATH" ]]; then
    rm -f "$LINK_PATH"
fi
ln -sf "$BSESSION_HOME/bsession" "$LINK_PATH"
info "bsession linked: $LINK_PATH → $BSESSION_HOME/bsession"

if ! echo "$PATH" | tr ':' '\n' | grep -qx "$BIN_DIR"; then
    warn "$BIN_DIR is not on your PATH. Add it:"
    warn "  export PATH=\"$BIN_DIR:\$PATH\""
fi

# ── Step 9: Install skills for supported platforms ───────────────────

# Find the repo source (for SKILL.md files)
REPO_SOURCE=""
if [[ -f "$TMPDIR/bsession/.claude/skills/browser/SKILL.md" ]]; then
    REPO_SOURCE="$TMPDIR/bsession"
elif [[ -f "$BSESSION_HOME/.claude/skills/browser/SKILL.md" ]]; then
    REPO_SOURCE="$BSESSION_HOME"
fi

# ── Claude Code ──
CLAUDE_SKILL_DIR="$HOME/.claude/skills/browser"
CLAUDE_SCRIPT_DIR="$CLAUDE_SKILL_DIR/scripts"
mkdir -p "$CLAUDE_SCRIPT_DIR"

if [[ -n "$REPO_SOURCE" && -f "$REPO_SOURCE/.claude/skills/browser/SKILL.md" ]]; then
    cp "$REPO_SOURCE/.claude/skills/browser/SKILL.md" "$CLAUDE_SKILL_DIR/SKILL.md"
fi
if [[ "$SCRIPT_DIR" != "$CLAUDE_SCRIPT_DIR" ]]; then
    cp "$SCRIPT_DIR/install.sh" "$CLAUDE_SCRIPT_DIR/install.sh"
    chmod +x "$CLAUDE_SCRIPT_DIR/install.sh"
fi
echo "$BSESSION_HOME" > "$CLAUDE_SKILL_DIR/.bsession-home"
info "Claude Code skill installed: $CLAUDE_SKILL_DIR"

# ── OpenClaw ──
OPENCLAW_SKILL_DIR="$HOME/.openclaw/workspace/skills/browser"
OPENCLAW_SCRIPT_DIR="$OPENCLAW_SKILL_DIR/scripts"
mkdir -p "$OPENCLAW_SCRIPT_DIR"

if [[ -n "$REPO_SOURCE" && -f "$REPO_SOURCE/.openclaw/skills/browser/SKILL.md" ]]; then
    cp "$REPO_SOURCE/.openclaw/skills/browser/SKILL.md" "$OPENCLAW_SKILL_DIR/SKILL.md"
fi
cp "$SCRIPT_DIR/install.sh" "$OPENCLAW_SCRIPT_DIR/install.sh"
chmod +x "$OPENCLAW_SCRIPT_DIR/install.sh"
echo "$BSESSION_HOME" > "$OPENCLAW_SKILL_DIR/.bsession-home"
info "OpenClaw skill installed: $OPENCLAW_SKILL_DIR"

# ── Step 10: Verify stack ────────────────────────────────────────────
if [[ "$NO_START" == false ]]; then
    info "Verifying stack..."
    ERRORS=0

    if docker exec agent-browser pgrep Xvfb &>/dev/null; then
        info "  Xvfb display: OK"
    else
        warn "  Xvfb display: NOT RUNNING"; ((ERRORS++))
    fi

    if docker exec agent-browser which agent-browser &>/dev/null; then
        info "  agent-browser CLI: OK"
    else
        warn "  agent-browser CLI: NOT FOUND"; ((ERRORS++))
    fi

    if docker exec agent-browser python3 -c "import sys; sys.path.insert(0, '/app'); from lib.browser import ab; print('ok')" 2>/dev/null | grep -q ok; then
        info "  browser.py: OK"
    else
        warn "  browser.py: IMPORT FAILED"; ((ERRORS++))
    fi

    if docker exec agent-browser ls /workspace/conf /workspace/data /workspace/scripts &>/dev/null; then
        info "  workspace mount: OK"
    else
        warn "  workspace mount: NOT MOUNTED"; ((ERRORS++))
    fi

    if [[ $ERRORS -gt 0 ]]; then
        warn "$ERRORS verification(s) failed. Check: docker compose -f $BSESSION_HOME/docker-compose.yml logs"
    else
        info "All checks passed!"
    fi
fi

# ── Done ──────────────────────────────────────────────────────────────
echo ""
info "=========================================="
info "  bsession is ready!"
info "=========================================="
info ""
info "  Install location:  $BSESSION_HOME"
info "  Workspace:         $WORKSPACE_DIR"
info "  VNC web access:    http://localhost:6080/vnc.html"
info ""
info "  Quick start (works from any directory):"
info "    bsession list              # list sessions"
info "    bsession run <name>        # start a session"
info ""
info "  /browser skill (Claude Code + OpenClaw):"
info "    /browser fetch <url>       # grab data from a URL"
info "    /browser new <name>        # create a new automation"
info "    /browser <session-id>      # debug a session"
info ""
