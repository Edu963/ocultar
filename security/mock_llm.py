import http.server
import json
import sys

class MockLLMHandler(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        
        # Log the received data to a file for verification
        with open("/tmp/mock_received.log", "ab") as f:
            f.write(b"--- NEW REQUEST ---\n")
            f.write(post_data)
            f.write(b"\n")
            f.flush()
        
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        
        # Echo back the data but wrapped, to simulate an AI response that might contain tokens
        response = {
            "choices": [{
                "message": {
                    "role": "assistant",
                    "content": f"I received: {post_data.decode('utf-8')}"
                }
            }]
        }
        self.wfile.write(json.dumps(response).encode('utf-8'))

if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 9999
    server = http.server.HTTPServer(('127.0.0.1', port), MockLLMHandler)
    print(f"Mock LLM listening on port {port}...")
    server.serve_forever()
