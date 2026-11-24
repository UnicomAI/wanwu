#!/bin/sh
# Run against an already-started sandbox. No model API calls or credentials needed.
set -eu
service=${1:-wga-sandbox}
case "$service" in
  wga-sandbox|wga-sandbox-ontology) ;;
  *) printf 'Usage: %s [wga-sandbox|wga-sandbox-ontology]\n' "$0" >&2; exit 2 ;;
esac
repo_dir=$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)
cd "$repo_dir"
set -- docker compose --env-file .env.image.amd64 --env-file .env
if [ -f .env.ontology ]; then
  set -- "$@" --env-file .env.ontology
fi
set -- "$@" --env-file "${FINSAFE_ENV_FILE:-.env.finsafe}" \
  -f docker-compose.yaml -f docker-compose.finsafe.yaml
"$@" exec -T "$service" /opt/finsafe/finsafe --version
"$@" exec -T "$service" python3 - <<'PY'
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
from urllib.request import urlopen

with urlopen("http://127.0.0.1:4096/global/health", timeout=5) as response:
    health = json.load(response)
assert health.get("healthy") is True, health
print("PASS OpenCode HTTP health", health)
with urlopen("http://127.0.0.1:4097/health", timeout=5) as response:
    assert response.read() == b"ok"
print("PASS Eino HTTP health")

# self-confine execs OpenCode, so check the server PID, not a wrapper process.
servers = {"/root/.opencode/bin/opencode": [], "/usr/local/bin/eino-agent": []}
for process in Path("/proc").iterdir():
    if not process.name.isdigit():
        continue
    try:
        executable = os.readlink(process / "exe").removesuffix(" (deleted)")
        if executable not in servers:
            continue
        status = (process / "status").read_text()
        cgroup = (process / "cgroup").read_text()
    except (FileNotFoundError, ProcessLookupError, PermissionError):
        continue
    fields = dict(line.split(":", 1) for line in status.splitlines() if ":" in line)
    assert fields["NoNewPrivs"].strip() == "1", (process.name, status)
    assert fields["Seccomp"].strip() == "2", (process.name, status)
    assert "/finsafe.slice/wrapper-self-confine-" in cgroup, (process.name, cgroup)
    servers[executable].append(process.name)
assert all(servers.values()), ("Missing FinSafe-confined runtime", servers)
print("PASS OpenCode and Eino cgroup, seccomp and NoNewPrivs; PIDs:", servers)

# Synthetic probe files only; this smoke test makes no deny-read guarantee.
outside = Path(tempfile.mkdtemp(prefix="wanwu-finsafe-probe-", dir="/var/lib"))
workspace = Path(tempfile.mkdtemp(prefix="finsafe-probe-", dir="/home/root/workspace"))
secret = outside / "synthetic-secret"
secret.write_text("fixture-only")
assert secret.read_text() == "fixture-only"
payload = r'''
import errno, pathlib, subprocess, sys
workspace, outside, policy = map(pathlib.Path, sys.argv[1:])
subprocess.run(["/bin/sh", "-c", 'printf allowed > "$1"', "probe", str(workspace / "allowed")], check=True)
assert (workspace / "allowed").read_text() == "allowed"
(workspace / ".env").write_text("FIXTURE=allowed\n")
assert (workspace / ".env").read_text() == "FIXTURE=allowed\n"
try:
    (outside / "synthetic-secret").read_text()
except OSError:
    print("DIAGNOSTIC synthetic read denied; sensitive-read coverage requires separate validation")
else:
    print("DIAGNOSTIC synthetic read allowed; no sensitive-read isolation is claimed")
for operation in (
    lambda: (outside / "forbidden").write_text("forbidden"),
    lambda: policy.write_text("forbidden"),
):
    try:
        operation()
    except OSError as error:
        assert error.errno in (errno.EACCES, errno.EPERM, errno.EROFS), error
    else:
        raise AssertionError("Out-of-scope write succeeded")
print("PASS child-process write, workspace .env and denied out-of-scope writes")
'''
try:
    for name in ("opencode", "eino-agent"):
        policy = "/etc/finsafe/" + name + ".yaml"
        result = subprocess.run(
            ["/opt/finsafe/finsafe", "--policy", policy,
             "self-confine", "--", "/usr/bin/python3", "-c", payload,
             str(workspace), str(outside), policy],
            cwd="/home/root/workspace",
            env={**os.environ, "HOME": "/home/root"},
            timeout=30,
        )
        assert result.returncode == 0, (name, result.returncode)
        assert not (outside / "forbidden").exists()
finally:
    shutil.rmtree(outside)
    shutil.rmtree(workspace)
print("PASS direct process-boundary checks; network egress and remote services are not restricted")
PY
