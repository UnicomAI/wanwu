"""Exercise the actual Eino Skill/Bash path, with independent AIO file checks.

The fixture model is used only for the baseline/kernel-boundary comparison.
The live mode calls the platform's real configured model and loads a real Skill.
No production key is read, copied, or logged by this test.
"""
import argparse
import base64
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import hashlib
import json
from pathlib import Path
import re
import shlex
from urllib.parse import urlencode
from urllib.request import Request, urlopen
import uuid


class Fixture(BaseHTTPRequestHandler):
    def log_message(self, *_):
        pass

    def do_POST(self):
        if self.path != "/v1/chat/completions":
            self.send_error(404)
            return
        body = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
        messages = body["messages"]
        if messages[-1]["role"] == "tool":
            delta, finish = {"role": "assistant", "content": "DONE"}, "stop"
        else:
            joined = json.dumps(messages)
            match = re.search(r"FIXTURE_COMMAND:([A-Za-z0-9+/=]+)", joined)
            if not match:
                self.send_error(400, "Missing fixture command")
                return
            command = base64.b64decode(match[1]).decode()
            delta = {"role": "assistant", "tool_calls": [{
                "index": 0, "id": "call_eino_probe", "type": "function",
                "function": {"name": "bash", "arguments": json.dumps({"command": command})},
            }]}
            finish = "tool_calls"
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.end_headers()
        for value in [
            {"id": "fixture", "object": "chat.completion.chunk", "model": "fixture",
             "choices": [{"index": 0, "delta": delta, "finish_reason": None}]},
            {"id": "fixture", "object": "chat.completion.chunk", "model": "fixture",
             "choices": [{"index": 0, "delta": {}, "finish_reason": finish}]},
        ]:
            self.wfile.write(("data: " + json.dumps(value) + "\n\n").encode())
        self.wfile.write(b"data: [DONE]\n\n")
        self.wfile.flush()


def request(base, path, body=None, timeout=180):
    req = Request(base + path, None if body is None else json.dumps(body).encode(),
                  {"Content-Type": "application/json"})
    return urlopen(req, timeout=timeout)


def walk(value):
    if isinstance(value, dict):
        yield value
        for child in value.values():
            yield from walk(child)
    elif isinstance(value, list):
        for child in value:
            yield from walk(child)


def posture(python, baseline=False):
    code = """from pathlib import Path
import os,json
out={}
for p in Path('/proc').iterdir():
 if not p.name.isdigit():continue
 try:
  exe=os.readlink(p/'exe').removesuffix(' (deleted)')
  if exe not in ('/root/.opencode/bin/opencode','/usr/local/bin/eino-agent'):continue
  s=dict(x.split(':',1) for x in (p/'status').read_text().splitlines() if ':' in x)
  cg=(p/'cgroup').read_text().strip()
  out[exe]={'pid':p.name,'NoNewPrivs':int(s['NoNewPrivs']),'Seccomp':int(s['Seccomp']),'cgroup':cg}
  if exe.endswith('/eino-agent'):
   env=dict(x.split('=',1) for x in (p/'environ').read_text().split(chr(0)) if '=' in x)
   out[exe]['guards_enabled']=all(env.get(k,'') not in ('1','true') for k in ('DISABLE_EINO_GENERIC_GUARD','DISABLE_EINO_SKILL_VAR_GUARD'))
 except (OSError,ProcessLookupError):continue
print(json.dumps(out))
"""
    result = json.loads(python(code))
    assert len(result) == 2, result
    assert result["/usr/local/bin/eino-agent"]["guards_enabled"], result
    for executable, info in result.items():
        if baseline and executable.endswith("/eino-agent"):
            assert "/wrapper-self-confine-" not in info["cgroup"], info
        else:
            assert info["NoNewPrivs"] == 1 and info["Seccomp"] == 2, info
            assert "/finsafe.slice/wrapper-self-confine-" in info["cgroup"], info
    return result


def opencode_regression(args, python):
    root = "/home/root/workspace/opencode-eino-regression-" + uuid.uuid4().hex
    outside = "/var/lib/opencode-eino-regression-" + uuid.uuid4().hex
    config = {"model": "fixture/fixture", "permission": {"*": "allow"}, "provider": {
        "fixture": {"npm": "@ai-sdk/openai-compatible", "options": {
            "baseURL": "http://finsafe-eino-fixture:8000/v1", "apiKey": "fixture-only"},
            "models": {"fixture": {"name": "Fixture"}}}}}
    python("from pathlib import Path; p=Path(" + repr(root) + "); (p/'.opencode').mkdir(parents=True); "
           "(p/'.opencode/opencode.json').write_text(" + repr(json.dumps(config)) + ")")
    query = "?" + urlencode({"directory": root})
    session = None

    def call(path, body=None, method=None):
        req = Request(args.opencode_url.rstrip("/") + path + query,
                      None if body is None else json.dumps(body).encode(),
                      {"Content-Type": "application/json"}, method=method)
        with urlopen(req, timeout=180) as response:
            data = response.read()
            return json.loads(data) if data else None

    def shell(command):
        result = call("/session/" + session + "/shell", {
            "agent": "build", "model": {"providerID": "fixture", "modelID": "fixture"},
            "command": command})
        return json.dumps(result)

    try:
        assert call("/global/health")["healthy"]
        session = call("/session", {})["id"]
        assert "OPENCODE_OK" in shell("printf OPENCODE_OK > first.txt; cat first.txt")
        denied = shell("printf synthetic > " + outside)
        assert "Permission denied" in denied or "Read-only file system" in denied, denied
        assert "AFTER_DENIAL_OK" in shell("printf AFTER_DENIAL_OK > second.txt; cat second.txt")
        python("from pathlib import Path; root=Path(" + repr(root) + "); "
               "assert (root/'first.txt').read_text()=='OPENCODE_OK'; "
               "assert (root/'second.txt').read_text()=='AFTER_DENIAL_OK'; "
               "assert not Path(" + repr(outside) + ").exists()")
        result = {"result": "PASS", "mode": "opencode-regression", "model_calls": 0,
                  "workspace_write_read": True, "out_of_scope_write_denied": True,
                  "valid_after_denial": True, "posture": posture(python)}
        if args.evidence:
            Path(args.evidence).write_text(json.dumps(result, indent=2) + "\n")
        print(json.dumps(result), flush=True)
    finally:
        if session:
            call("/session/" + session, method="DELETE")
        python("import shutil; shutil.rmtree(" + repr(root) + ",ignore_errors=True)")


def run(args):
    aio = args.aio_url.rstrip("/")
    eino = args.eino_url.rstrip("/")

    def python(code):
        with request(aio, "/v1/shell/exec", {
            "command": "python3 -c " + shlex.quote(code), "exec_dir": "/tmp",
        }, timeout=60) as response:
            value = json.load(response)
        assert value["success"], value.get("message")
        data = value["data"]
        assert data["exit_code"] == 0, data.get("output")
        return data.get("output", "").strip()

    if args.opencode_regression:
        return opencode_regression(args, python)
    token = uuid.uuid4().hex
    root = "/home/root/workspace/eino-compat-" + token
    outside = "/var/lib/eino-compat-" + token
    live = args.mode == "live"
    model_url = args.live_model_url if live else "http://finsafe-eino-fixture:8000/v1"
    model = "deepseek-flash" if live else "fixture"
    assert model_url, "Live mode requires --live-model-url (no key needed)"
    expected = "DENIED" if args.mode == "confined" else "WRITTEN"
    script = """from pathlib import Path
import errno,json
root=Path.cwd()
items=json.loads((root/'input/data.json').read_text())
total=sum(x['quantity']*x['price'] for x in items)
(root/'output/first.txt').write_text('EINO_OK')
result={'total':total,'valid_after_denial':True}
"""
    if not live:
        script += (
            "try:\n Path(" + repr(outside) + ").write_text('synthetic probe')\n"
            "except OSError as error:\n"
            " assert error.errno in (errno.EACCES,errno.EPERM,errno.EROFS),error\n"
            " result['boundary']='DENIED'\n"
            "else:\n result['boundary']='WRITTEN'\n"
        )
    script += "(root/'output/result.json').write_text(json.dumps(result))\nprint('EINO_SCRIPT_COMPLETE',total)\n"
    digest = hashlib.sha256(script.encode()).hexdigest()
    skill = """---
name: eino-compat
description: Validate a preinstalled sales calculation script using the bash tool.
---
Run `python3 tmp/check.py` from the workspace root, then read output/result.json.
The script and input are preinstalled. Do not modify them or install packages.
"""
    env = "\n".join([
        "OPENAI_API_KEY=fixture-only", "OPENAI_BASE_URL=" + model_url,
        "OPENAI_MODEL_ID=" + model,
        "OPENAI_LLM_PARAMS='" + json.dumps({"maxTokensEnable": True, "maxTokens": 2048,
                                          "thinkingEnable": False}) + "'", "",
    ])
    files = {".env": env, "tmp/check.py": script, "skills/eino-compat/SKILL.md": skill,
             "input/data.json": json.dumps([{"quantity": 3, "price": 7}, {"quantity": 4, "price": 8}]),
             ".skill_env.json": json.dumps({"EINO_FIXTURE_SETTING": "synthetic-only"})}
    setup = "from pathlib import Path\nimport json\nroot=Path(" + repr(root) + ")\n"
    setup += "files=json.loads(" + repr(json.dumps(files)) + ")\n"
    setup += "for name,text in files.items():\n p=root/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_text(text)\n"
    setup += "(root/'output').mkdir()\n"
    python(setup)
    with request(eino, "/health", timeout=10) as response:
        assert response.read() == b"ok"
    command = "python3 tmp/check.py"
    if live:
        prompt = ("Use the eino-compat skill to validate the preinstalled sales calculation. "
                  "Run exactly `python3 tmp/check.py` with bash from the workspace root, "
                  "then read output/result.json and report the total. Do not edit the script or input.")
    else:
        prompt = "Execute the compatibility probe. FIXTURE_COMMAND:" + base64.b64encode(command.encode()).decode()
    events, names, commands = [], set(), []
    print("START", args.mode, "actual Eino /chat", flush=True)
    try:
        with request(eino, "/chat?" + urlencode({"workspace": root, "agent_type": "chat-model"}),
                     {"messages": [{"role": "user", "content": prompt}]}) as response:
            assert response.headers.get_content_type() == "text/event-stream"
            for line in response:
                if not line.startswith(b"data:"):
                    continue
                event = json.loads(line[5:])
                events.append(event)
                for obj in walk(event):
                    name = obj.get("name", obj.get("Name"))
                    arguments = obj.get("arguments", obj.get("Arguments"))
                    if name and isinstance(arguments, str):
                        names.add(name)
                        if name == "bash":
                            commands.append(json.loads(arguments).get("command"))
        # Real models add cd, stderr capture and subsequent output inspection.
        # Identify the exact script invocation as a shell command segment; the
        # independent file/hash assertions below verify its actual effects.
        def invoked_script(value):
            lexer = shlex.shlex(value, posix=True, punctuation_chars=";&|<>")
            lexer.whitespace_split = True
            segments = [[]]
            for token in lexer:
                if token in ("&&", ";"):
                    segments.append([])
                else:
                    segments[-1].append(token)
            if segments[0][:1] == ["cd"] and segments[0] != ["cd", root]:
                return False
            return any(s in (["python3", "tmp/check.py"],
                             ["python3", "tmp/check.py", "2", ">&", "1"])
                       for s in segments)

        assert any(invoked_script(c) for c in commands if isinstance(c, str)), (
            "Requested Bash command was not observed", names, commands)
        check = """from pathlib import Path
import json,hashlib
root=Path(ROOT)
assert hashlib.sha256((root/'tmp/check.py').read_bytes()).hexdigest()==DIGEST
assert (root/'output/first.txt').read_text()=='EINO_OK'
result=json.loads((root/'output/result.json').read_text())
assert result['total']==53 and result['valid_after_denial']
print(json.dumps(result))
""".replace("ROOT", repr(root)).replace("DIGEST", repr(digest))
        result = json.loads(python(check))
        if not live:
            assert result["boundary"] == expected, result
            exists = python("from pathlib import Path; print(Path(" + repr(outside) + ").exists())")
            assert exists == str(expected == "WRITTEN"), exists
        result.update({"result": "PASS", "mode": args.mode, "actual_eino_server": True,
                       "real_model": live, "tool_names": sorted(names), "event_count": len(events),
                       "bash_command_observed": True, "bash_commands": commands,
                       "original_script_unchanged": True})
        result["posture"] = posture(python, baseline=args.mode == "baseline")
        if live:
            assert any("skill" in name.lower() for name in names), ("No Skill tool invocation", names)
        if args.evidence:
            Path(args.evidence).write_text(json.dumps(result, indent=2) + "\n")
        print(json.dumps(result), flush=True)
    finally:
        python("import shutil; from pathlib import Path; shutil.rmtree(" + repr(root) +
               ",ignore_errors=True); Path(" + repr(outside) + ").unlink(missing_ok=True)")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--serve-fixture", action="store_true")
    parser.add_argument("--mode", choices=["baseline", "confined", "live"], default="confined")
    parser.add_argument("--aio-url", default="http://finsafe-x64-host:18080")
    parser.add_argument("--eino-url", default="http://finsafe-x64-host:14097")
    parser.add_argument("--live-model-url")
    parser.add_argument("--opencode-regression", action="store_true")
    parser.add_argument("--opencode-url", default="http://finsafe-x64-host:14096")
    parser.add_argument("--evidence")
    args = parser.parse_args()
    if args.serve_fixture:
        print("EINO_FIXTURE_READY", flush=True)
        ThreadingHTTPServer(("0.0.0.0", 8000), Fixture).serve_forever()
    else:
        run(args)
