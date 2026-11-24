"""OpenCode API checks inside the Docker fixture; no paid model calls."""
import argparse
import json
from pathlib import Path
import queue
import subprocess
import threading
import time
from urllib.parse import urlencode
from urllib.request import Request, urlopen
import uuid


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--confined", action="store_true")
    args = parser.parse_args()
    base = "http://127.0.0.1:4096"
    root = Path("/home/root/workspace/compatibility-" + uuid.uuid4().hex)
    (root / ".opencode").mkdir(parents=True)
    config = {
        "model": "fixture/fixture",
        "permission": {"*": "allow"},
        "provider": {"fixture": {
            "npm": "@ai-sdk/openai-compatible",
            "options": {"baseURL": "http://127.0.0.1:8000/v1", "apiKey": "fixture-only"},
            "models": {"fixture": {"name": "Fixture"}},
        }},
    }
    (root / ".opencode/opencode.json").write_text(json.dumps(config))
    query = "?" + urlencode({"directory": str(root)})

    def request(path, body=None, method=None, timeout=180):
        data = None if body is None else json.dumps(body).encode()
        req = Request(base + path + query, data, {"Content-Type": "application/json"}, method=method)
        with urlopen(req, timeout=timeout) as response:
            value = response.read()
            return json.loads(value) if value else None

    print("START OpenCode compatibility", {"confined": args.confined}, flush=True)
    for _ in range(90):
        try:
            health = request("/global/health", timeout=2)
            assert health["healthy"] is True, health
            break
        except (OSError, AssertionError):
            time.sleep(1)
    else:
        raise AssertionError("OpenCode did not become healthy")
    print("PASS health", health, flush=True)

    events = queue.Queue()

    def read_events():
        try:
            with urlopen(base + "/global/event", timeout=180) as response:
                assert response.headers.get_content_type() == "text/event-stream"
                for line in response:
                    if line.startswith(b"data:"):
                        events.put(json.loads(line[5:]))
        except Exception as error:
            events.put({"reader_error": str(error)})

    threading.Thread(target=read_events, daemon=True).start()
    connected = events.get(timeout=180)
    assert "reader_error" not in connected, connected
    session = request("/session", {})["id"]

    def shell(command):
        result = request("/session/" + session + "/shell", {
            "agent": "build", "model": {"providerID": "fixture", "modelID": "fixture"},
            "command": command,
        })
        tools = [part for part in result["parts"] if part["type"] == "tool"]
        assert tools and tools[-1]["state"]["status"] == "completed", result
        return tools[-1]["state"]["output"]

    output = shell("printf COMPAT_R1 > allowed.txt; cat allowed.txt")
    assert "COMPAT_R1" in output and (root / "allowed.txt").read_text() == "COMPAT_R1", output
    print("PASS actual OpenCode Bash tool wrote and read its workspace", flush=True)

    outside = Path("/var/lib/finsafe-compat-" + uuid.uuid4().hex + ".txt")
    output = shell("printf COMPAT_R2 > " + str(outside))
    if args.confined:
        assert not outside.exists(), "Out-of-scope tool write succeeded"
        assert "Permission denied" in output or "Read-only file system" in output, output
        print("PASS actual OpenCode tool out-of-scope write denied", flush=True)
    else:
        assert outside.read_text() == "COMPAT_R2", output
        outside.unlink()
        print("PASS baseline can write the same out-of-scope directory", flush=True)

    output = shell("cat allowed.txt; printf COMPAT_R3 > second.txt")
    assert "COMPAT_R1" in output and (root / "second.txt").read_text() == "COMPAT_R3", output
    print("PASS legitimate tool execution still works after denial", flush=True)

    observed = [connected]
    deadline = time.monotonic() + 30
    while time.monotonic() < deadline:
        try:
            event = events.get(timeout=1)
            observed.append(event)
            if "COMPAT_R1" in json.dumps(event):
                break
        except queue.Empty:
            continue
    assert any("COMPAT_R1" in json.dumps(event) for event in observed), observed[-5:]
    print("PASS SSE delivered the actual tool output", flush=True)

    pid = subprocess.check_output(["supervisorctl", "pid", "opencode"], text=True).strip()
    status = Path("/proc/" + pid + "/status").read_text()
    cgroup = Path("/proc/" + pid + "/cgroup").read_text()
    fields = dict(line.split(":", 1) for line in status.splitlines() if ":" in line)
    if args.confined:
        assert fields["NoNewPrivs"].strip() == "1", status
        assert fields["Seccomp"].strip() == "2", status
        assert "/finsafe.slice/wrapper-self-confine-" in cgroup, cgroup
        print("PASS Supervisor's actual OpenCode PID has FinSafe cgroup/seccomp/NoNewPrivs", flush=True)
    request("/session/" + session, method="DELETE")
    print(json.dumps({"result": "PASS", "confined": args.confined, "opencode": health["version"]}), flush=True)


if __name__ == "__main__":
    main()
