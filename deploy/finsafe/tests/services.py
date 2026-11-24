"""Test fixtures only: deterministic model responses and the AIO file/shell transport.

Wanwu and OpenCode are real. These endpoints avoid external model charges and
the full AIO image; their use must be stated in the integration report.
"""
import base64
from email.parser import BytesParser
from email.policy import default
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
from pathlib import Path
import re
import subprocess
import threading
from urllib.parse import parse_qs, urlparse


class Fixture(BaseHTTPRequestHandler):
    def log_message(self, *_):
        pass

    def reply(self, data, status=200):
        body = json.dumps(data).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        parsed = urlparse(self.path)
        if parsed.path != "/v1/file/download":
            return self.reply({"error": "unknown fixture route"}, 404)
        path = self.workspace_path(parse_qs(parsed.query)["path"][0])
        body = path.read_bytes()
        self.send_response(200)
        self.send_header("Content-Type", "application/octet-stream")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    @staticmethod
    def workspace_path(value):
        path = Path(value).resolve()
        assert path.is_relative_to("/home/root/workspace"), path
        return path

    def do_POST(self):
        raw = self.rfile.read(int(self.headers.get("Content-Length", 0)))
        if self.path.endswith("/chat/completions"):
            return self.model(json.loads(raw))
        if self.path == "/v1/shell/exec":
            data = json.loads(raw)
            result = subprocess.run(data["command"], shell=True, cwd=data.get("exec_dir", "/"),
                                    capture_output=True, text=True, timeout=30)
            return self.reply({"success": True, "data": {
                "status": "completed", "session_id": "fixture", "exit_code": result.returncode,
                "output": result.stdout + result.stderr,
            }})
        if self.path == "/v1/file/upload":
            message = BytesParser(policy=default).parsebytes(
                ("Content-Type: " + self.headers["Content-Type"] + "\r\n\r\n").encode() + raw)
            fields = {part.get_param("name", header="content-disposition"): part.get_payload(decode=True)
                      for part in message.iter_parts()}
            path = self.workspace_path(fields["path"].decode())
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(fields["file"])
            return self.reply({"success": True, "data": {
                "success": True, "file_path": str(path), "file_size": len(fields["file"]),
            }})
        return self.reply({"error": "unknown fixture route"}, 404)

    def model(self, data):
        messages = data["messages"]
        user_index = max((i for i, message in enumerate(messages) if message["role"] == "user"), default=-1)
        text = json.dumps(messages[user_index].get("content", "")) if user_index >= 0 else ""
        match = re.search(r"WANWU_TEST_COMMAND:([A-Za-z0-9+/=]+)", text)
        completed = any(message["role"] == "tool" for message in messages[user_index + 1:])
        tools = {tool["function"]["name"] for tool in data.get("tools", [])}
        if match and not completed and "bash" in tools:
            delta = {"role": "assistant", "tool_calls": [{
                "index": 0, "id": "call_fixture", "type": "function", "function": {
                    "name": "bash", "arguments": json.dumps({
                        "command": base64.b64decode(match[1]).decode(), "description": "Integration fixture",
                    }),
                },
            }]}
            reason = "tool_calls"
        else:
            delta = {"role": "assistant", "content": "WANWU_INTEGRATION_DONE" if match else "Fixture title"}
            reason = "stop"
        if not data.get("stream"):
            return self.reply({
                "id": "fixture", "object": "chat.completion", "created": 1, "model": "fixture",
                "choices": [{"index": 0, "message": delta, "finish_reason": reason}],
                "usage": {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
            })
        chunks = [
            {"id": "fixture", "object": "chat.completion.chunk", "created": 1, "model": "fixture",
             "choices": [{"index": 0, "delta": delta, "finish_reason": None}]},
            {"id": "fixture", "object": "chat.completion.chunk", "created": 1, "model": "fixture",
             "choices": [{"index": 0, "delta": {}, "finish_reason": reason}],
             "usage": {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2}},
        ]
        body = b"".join(b"data: " + json.dumps(chunk).encode() + b"\n\n" for chunk in chunks) + b"data: [DONE]\n\n"
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


if __name__ == "__main__":
    for port in (8000, 8080):
        server = ThreadingHTTPServer(("127.0.0.1", port), Fixture)
        threading.Thread(target=server.serve_forever, daemon=True).start()
    print("FIXTURE_SERVICES_READY", flush=True)
    threading.Event().wait()
