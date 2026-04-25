#!/usr/bin/env python3
"""
Mock AI server for end-to-end testing of SOMBRA.
Echoes back the prompt it receives so we can verify:
  1. The prompt has NO real PII (it was redacted before reaching us)
  2. SOMBRA re-hydrates any vault tokens we echo back in our response
"""
import json
from http.server import BaseHTTPRequestHandler, HTTPServer

class MockAI(BaseHTTPRequestHandler):
    def do_POST(self):
        # Accept any path — handles /v1/chat/completions and similar
        length = int(self.headers.get('Content-Length', 0))
        raw = self.rfile.read(length)
        try:
            body = json.loads(raw)
            msgs = body.get("messages", [])
            prompt = msgs[-1].get("content", "") if msgs else ""
        except Exception:
            prompt = raw.decode()

        print("\n── MOCK AI RECEIVED ──────────────────")
        print(prompt[:1000])
        print("──────────────────────────────────────\n")

        # Echo the prompt back so SOMBRA can re-hydrate any vault tokens.
        response = {
            "choices": [{
                "message": {
                    "role": "assistant",
                    "content": f"Summary of transactions:\n{prompt}"
                }
            }]
        }
        body = json.dumps(response).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, fmt, *args):
        pass  # silence access logs

print("Mock AI server listening on :9090")
HTTPServer(("", 9090), MockAI).serve_forever()
