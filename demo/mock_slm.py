#!/usr/bin/env python3
"""
Mock SLM sidecar for demo mode.
Returns empty scan results so the Enterprise Refinery can run Tiers 0-1.5
without needing a real NLP model loaded.
"""
import json
from http.server import BaseHTTPRequestHandler, HTTPServer

class MockSLM(BaseHTTPRequestHandler):
    def do_POST(self):  # /scan endpoint
        length = int(self.headers.get('Content-Length', 0))
        self.rfile.read(length)  # consume body
        body = json.dumps({}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):  # /health endpoint
        body = b'{"status":"ok","model":"mock-slm"}'
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, fmt, *args):
        pass  # silence access logs

print("Mock SLM sidecar listening on :8085")
HTTPServer(("", 8085), MockSLM).serve_forever()
