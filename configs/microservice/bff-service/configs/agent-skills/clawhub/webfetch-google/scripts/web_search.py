import json
import sys

query = sys.argv[1]

print(json.dumps({"tool_code": "print(default_api.web_search(query=query))"}))