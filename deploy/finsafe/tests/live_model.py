"""Real DeepSeek request through OpenCode; credentials come from server environment."""
import json
from pathlib import Path
import time
from urllib.error import HTTPError
from urllib.parse import urlencode
from urllib.request import Request, urlopen
import uuid

root = Path("/home/root/workspace/live-deepseek-" + uuid.uuid4().hex)
(root / ".opencode").mkdir(parents=True)
(root / ".opencode/opencode.json").write_text(json.dumps({
    "permission": {"bash": "allow", "edit": "allow", "read": "allow"},
}))
query = "?" + urlencode({"directory": str(root)})
base = "http://127.0.0.1:4096"


def request(path, body=None, timeout=180):
    req = Request(base + path + query, None if body is None else json.dumps(body).encode(),
                  {"Content-Type": "application/json"})
    try:
        with urlopen(req, timeout=timeout) as response:
            return json.load(response)
    except HTTPError as error:
        # Do not dump provider request objects or authentication headers.
        raise RuntimeError("OpenCode HTTP status " + str(error.code)) from None


for attempt in range(60):
    try:
        health = request("/global/health", timeout=2)
        assert health["healthy"]
        break
    except (OSError, AssertionError):
        time.sleep(1)
else:
    raise SystemExit("OpenCode did not become healthy")

session = request("/session", {})["id"]
reply = request("/session/" + session + "/message", {
    "model": {"providerID": "deepseek", "modelID": "deepseek-flash"},
    "parts": [{"type": "text", "text": (
        "Use a tool to create real-model.txt in the current working directory with exactly "
        "DEEPSEEK_LIVE_OK as its contents. Read the file back with a tool to verify it. "
        "Stay within the current working directory. Finally reply DEEPSEEK_LIVE_DONE."
    )}],
})
assert (root / "real-model.txt").read_text().strip() == "DEEPSEEK_LIVE_OK", "Wrong artifact content"
messages = request("/session/" + session + "/message")
tools = [part["tool"] for message in messages for part in message.get("parts", [])
         if part.get("type") == "tool" and part.get("state", {}).get("status") == "completed"]
assert tools, "No completed real tool invocation"
assert not reply.get("info", {}).get("error"), "Provider returned an error"
print(json.dumps({
    "result": "PASS", "model": "deepseek/deepseek-flash", "opencode": health["version"],
    "real_provider": True, "tool_names": tools, "artifact_content_verified": True,
    "tokens": reply.get("info", {}).get("tokens"),
}), flush=True)
