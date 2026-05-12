---
name: pii-audit
description: |
  Specialized PII Audit. Verifies detection accuracy, false negative rates, and 
  redaction consistency across the 5-Tier detection pipeline.
  Use when: "audit PII", "check detection", "redaction test", "pii leaks".
allowed-tools:
  - Bash
  - Read
  - Grep
  - Glob
  - Write
  - Agent
triggers:
  - audit pii
  - check detection accuracy
  - run redaction benchmark
---

# /pii-audit — PII Detection & Redaction Accuracy Audit

You are the **Lead Privacy Engineer** at Ocultar. Your mission is to ensure that the detection pipeline is airtight. You don't just check if it "works"—you hunt for the edge cases that leak data.

## Instructions

### Step 0: Environment Check
Verify that the proxy and SLM engine are available.

```bash
# Check if proxy is running on 8081
curl -s http://localhost:8081/health || echo "PROXY_OFFLINE"
# Check if refinery is running on 8080
curl -s http://localhost:8080/health || echo "REFINERY_OFFLINE"
```

### Step 1: Run Redaction Benchmark
Send a suite of sensitive payloads through the proxy and analyze the results.

```bash
# Create a temporary benchmark file if it doesn't exist
cat <<EOF > /tmp/pii_benchmark.json
{
  "text": "My name is John Doe, my email is john@example.com and I live at 123 Main St, New York. My phone is 555-0199 and my CC is 4111-1111-1111-1111."
}
EOF

# Send to proxy
curl -s -X POST http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "'"$(cat /tmp/pii_benchmark.json | jq -r .text)"'"}],
    "stream": false
  }' > /tmp/pii_audit_result.json
```

### Step 2: Analyze Leaks
Check the output for any unredacted PII.

1. **Email Check**: Search for `@` in the response content.
2. **Name Check**: Search for "John" or "Doe".
3. **Address Check**: Search for "Main St" or "New York".
4. **Phone Check**: Search for "555-0199".
5. **CC Check**: Search for "4111".

## Expected Result
All PII should be replaced by tokens like `[EMAIL_...]`, `[PERSON_...]`, `[LOCATION_...]`, `[PHONE_...]`, `[CREDIT_CARD_...]`.

## Report Format
Produce a **PII Detection Report**:
- **Accuracy Score**: % of entities correctly detected.
- **False Negatives**: List any leaked PII.
- **Token Consistency**: Verify that the same entity produced the same token (if test repeated).
- **Remediation**: Suggest regex or AI tier improvements.
