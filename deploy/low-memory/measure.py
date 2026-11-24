"""Sample actual deployment memory while a separate client runs a real task.

Docker stats excludes processes moved into sibling FinSafe scopes. Read their
shared parent separately; on the ARM test host it is already inside the VM total.
No credentials or full Docker configuration are collected.
"""
import argparse
from datetime import datetime, timezone
import json
from pathlib import Path
import re
import shlex
import subprocess
import time
from urllib.request import Request, urlopen

MAIN_SERVICES = {
    "mysql", "redis", "minio", "kafka", "es", "bff-service", "iam-service",
    "model-service", "mcp-service", "knowledge-service", "rag-service",
    "assistant-service", "agent-service", "operate-service", "app-service",
    "channel-service", "callback", "workflow", "rag", "wga-sandbox", "nginx",
}


def docker(*args):
    return subprocess.check_output(["docker", *args], text=True).strip()


def memory_bytes(value):
    amount, unit = re.fullmatch(r"([\d.]+)\s*([A-Za-z]+)", value).groups()
    return int(float(amount) * {"B": 1, "KiB": 1024, "MiB": 1024**2,
                               "GiB": 1024**3, "kB": 1000, "MB": 1000**2,
                               "GB": 1000**3}[unit])


def shared_parent(aio):
    code = """from pathlib import Path
import json
r=Path('/run/finsafe-budget')
if not r.exists():r=Path('/sys/fs/cgroup')
print(json.dumps({'current':int((r/'memory.current').read_text()),
 'peak':int((r/'memory.peak').read_text()),'limit':int((r/'memory.max').read_text()),
 'events':dict((k,int(v)) for k,v in (line.split() for line in (r/'memory.events').read_text().splitlines())),
 'container_cgroup':Path('/proc/1/cgroup').read_text().strip()}))
"""
    req = Request(aio.rstrip("/") + "/v1/shell/exec", json.dumps({
        "command": "python3 -c " + shlex.quote(code), "exec_dir": "/tmp",
    }).encode(), {"Content-Type": "application/json"})
    with urlopen(req, timeout=90) as response:
        value = json.load(response)
    assert value["success"] and value["data"]["exit_code"] == 0
    value = json.loads(value["data"]["output"])
    assert value["limit"] == 1536 * 1024**2, value
    assert any(p in value["container_cgroup"] for p in (
        "/finsafe-wanwu/", "/finsafe-wanwu.slice/")), value
    assert not value["events"]["oom_kill"], value
    return value


def states(names):
    template = '{{json .State}}|{{.RestartCount}}'
    result = {}
    for name in names:
        raw, count = docker("inspect", name, "--format", template).rsplit("|", 1)
        state = json.loads(raw)
        assert state["Running"] and not state["OOMKilled"], (name, state)
        health = state.get("Health", {}).get("Status", "not-configured")
        assert health in ("healthy", "not-configured"), (name, health)
        result[name] = {"restarts": int(count), "health": health}
    return result


def cgroup_checks(names):
    result = {}
    for name in names:
        lines = docker("exec", name, "sh", "-c", "cat /sys/fs/cgroup/memory.current "
                       "/sys/fs/cgroup/memory.swap.current /sys/fs/cgroup/memory.events").splitlines()
        events = dict((k, int(v)) for k, v in (line.split() for line in lines[2:]))
        assert events["oom_kill"] == 0, (name, events)
        assert int(lines[1]) == 0, (name, "swap is in use", lines[1])
        result[name] = {"current_bytes": int(lines[0]), "swap_bytes": int(lines[1]), "events": events}
    return result


def run(args):
    # Include created/stopped containers: selecting only running ones can hide
    # services left unstarted after a dependency failed during recreation.
    rows = docker("ps", "-a", "--filter", "label=com.docker.compose.project=" + args.project,
                  "--format", '{{.Names}}|{{.Label "com.docker.compose.service"}}').splitlines()
    by_service = {service: name for name, service in (line.split("|", 1) for line in rows)}
    assert not MAIN_SERVICES.difference(by_service), ("Missing main services", MAIN_SERVICES.difference(by_service))
    names = sorted(set([by_service[s] for s in MAIN_SERVICES] + args.extra))
    initial = states(names)
    samples, peaks = [], {}
    start = time.monotonic()
    while True:
        rows = [json.loads(line) for line in docker(
            "stats", "--no-stream", "--format", "{{json .}}", *names).splitlines()]
        used = {r["Name"]: memory_bytes(r["MemUsage"].split(" / ")[0]) for r in rows}
        for name, size in used.items():
            peaks[name] = max(peaks.get(name, 0), size)
        sample = {"elapsed_s": round(time.monotonic() - start, 1),
                  "host_docker_bytes": sum(used.values()),
                  "shared_sandbox": shared_parent(args.aio_url)}
        samples.append(sample)
        print("SAMPLE", sample["elapsed_s"], "host Docker MiB", round(sum(used.values()) / 1024**2), flush=True)
        if time.monotonic() - start >= args.seconds:
            break
        time.sleep(5)
    final = states(names)
    assert all(final[n]["restarts"] == initial[n]["restarts"] for n in names), (initial, final)
    result = {"result": "PASS", "timestamp": datetime.now(timezone.utc).isoformat(),
              "verified_main_services": sorted(MAIN_SERVICES),
              "host_docker_memory_bytes": int(docker("info", "--format", "{{.MemTotal}}")),
              "host_docker_peak_observed_bytes": max(x["host_docker_bytes"] for x in samples),
              "container_peaks_bytes": peaks, "final_states": final, "samples": samples,
              "cgroup_checks": cgroup_checks(names),
              "note": "Observed interval only; includes QEMU overhead when supplied in --extra. Not a capacity benchmark."}
    Path(args.evidence).write_text(json.dumps(result, indent=2) + "\n")
    print("PASS memory samples, shared sandbox cap, health, no OOM kills or new restarts", flush=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--project", default="wanwu-finsafe-ui")
    parser.add_argument("--extra", action="append", default=[])
    parser.add_argument("--aio-url", default="http://127.0.0.1:18080")
    parser.add_argument("--seconds", type=int, default=60)
    parser.add_argument("--evidence", required=True)
    run(parser.parse_args())
