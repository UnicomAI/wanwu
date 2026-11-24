"""Loopback-only forwarding gateway for real DeepSeek tests; never logs keys/bodies."""
import argparse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
from pathlib import Path
import threading
from urllib.error import HTTPError
from urllib.request import ProxyHandler, Request, build_opener

parser = argparse.ArgumentParser()
parser.add_argument("--key-env-file", required=True)
parser.add_argument("--proxy", default="http://host.docker.internal:1087")
parser.add_argument("--max-requests", type=int, default=12)
args = parser.parse_args()
values = {}
for raw in Path(args.key_env_file).read_text().splitlines():
    if "=" in raw and not raw.lstrip().startswith("#"):
        key, value = raw.removeprefix("export ").strip().split("=", 1)
        values[key] = value.strip().strip("\"'")
api_key = values["DEEPSEEK_API_KEY"]
opener = build_opener(ProxyHandler({"https": args.proxy}))
lock = threading.Lock()
forwarded = 0


class Gateway(BaseHTTPRequestHandler):
    def log_message(self, *_):
        pass

    def reply(self, status, data):
        body = json.dumps(data).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        if self.path == "/health":
            return self.reply(200, {"healthy": True, "forwarded": forwarded})
        return self.reply(404, {"error": "unknown test gateway route"})

    def do_POST(self):
        global forwarded
        if self.path != "/v1/chat/completions":
            return self.reply(404, {"error": "unknown test gateway route"})
        with lock:
            if forwarded >= args.max_requests:
                return self.reply(429, {"error": "live test request budget exhausted"})
            forwarded += 1
            request_number = forwarded
        size = int(self.headers.get("Content-Length", 0))
        if not 0 < size < 4 * 1024 * 1024:
            return self.reply(413, {"error": "invalid test request size"})
        body = json.loads(self.rfile.read(size))
        body["model"] = "deepseek-flash"
        body["max_tokens"] = min(int(body.get("max_tokens", 2048)), 2048)
        body["thinking"] = {"type": "disabled"}
        request = Request("https://api.deepseek.com/v1/chat/completions", json.dumps(body).encode(), {
            "Authorization": "Bearer " + api_key,
            "Content-Type": "application/json",
        })
        try:
            with opener.open(request, timeout=180) as upstream:
                self.send_response(upstream.status)
                self.send_header("Content-Type", upstream.headers.get("Content-Type", "application/json"))
                self.send_header("Connection", "close")
                self.end_headers()
                for line in upstream:
                    self.wfile.write(line)
                    self.wfile.flush()
                self.close_connection = True
                print(json.dumps({"live_request": request_number, "upstream_status": upstream.status,
                                  "model": "deepseek-flash"}), flush=True)
        except HTTPError as error:
            print(json.dumps({"live_request": request_number, "upstream_status": error.code}), flush=True)
            return self.reply(error.code, {"error": "official provider rejected the test request"})
        except Exception as error:
            print(json.dumps({"live_request": request_number, "error_type": type(error).__name__}), flush=True)
            return self.reply(502, {"error": "official provider connection failed"})


print("LIVE_DEEPSEEK_GATEWAY_READY", flush=True)
ThreadingHTTPServer(("127.0.0.1", 8000), Gateway).serve_forever()
