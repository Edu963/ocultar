#!/bin/bash
set -e

echo ""
echo "================================================================"
echo "          OCULTAR Enterprise Proxy — Smoke Test               "
echo "================================================================"
echo ""

PROXY_URL=${1:-"http://localhost:8081"}

# Wait for the proxy to be ready
echo "[*] Waiting for proxy to be healthy at $PROXY_URL..."
for i in {1..30}; do
  if curl -s -f "$PROXY_URL/healthz" > /dev/null; then
    echo "[+] Proxy is healthy!"
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "[-] Proxy healthcheck timed out."
    exit 1
  fi
  sleep 2
done

# Perform a smoke test using the leaky dataset
echo "[*] Running smoke test with leaky_demo.json payload..."

# We simulate a POST containing PII. In a real proxy, it redirects to upstream. 
# Here, we point it to an endpoint that echoes the request back, or we just rely
# on the built-in batch processing endpoint for the smoke test if upstream isn't up.
# For a generic reverse proxy test, we'll POST /v1/chat/completions style payload 

PAYLOAD=$(cat << 'EOF'
{
  "messages": [
    {
      "role": "user",
      "content": "Hello, my phone number is +33 6 12 34 56 78 and my email is test.user@enterprise.com. Please analyze my profile."
    }
  ]
}
EOF
)

# Call the proxy
echo "[*] Sending raw PII payload:"
echo "$PAYLOAD"
echo ""

RESPONSE=$(curl -s -X POST "$PROXY_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD")

echo "[*] Upstream received:"
echo "$RESPONSE"

# Verify redaction occurred via proxy logs (since responses get rehydrated automatically)
if docker logs ocultar-proxy --tail 20 2>&1 | grep -q 'redacted: true'; then
  echo ""
  echo "[+] SUCCESS: PII successfully intercepted and redacted!"
  exit 0
fi

# Fallback check incase the proxy returned the literal response unmodified 
# or if it was blocked entirely due to upstream connection refused
if echo "$RESPONSE" | grep -q 'test.user@enterprise.com'; then
  echo ""
  echo "[-] FAILURE: PII leaked through proxy!"
  exit 1
fi

echo ""
echo "[-] FAILURE: Analysis couldn't verify redaction tokens. Check proxy logs."
exit 1
