# 🔍 Troubleshooting & Diagnostics

This guide helps you identify and resolve common issues encountered while running OCULTAR.

## 🛠️ The "Big Three" Diagnostics

Before diving deep, check these three points:
1.  **Docker Health**: `docker ps` — Are all containers running?
2.  **Health Endpoint**: `curl localhost:4141/api/health` — Is the Refinery responsive?
3.  **Logs**: `docker compose logs -f sombra` — Any error messages?

---

## 🚩 Common Issues

### 1. Dashboard is Blank / "Connection Refused"
- **Cause**: Sombra hasn't finished booting or the local AI model is still downloading.
- **Fix**: Wait 60 seconds. Check Sombra logs for "Sombra Gateway running on...".

### 2. "PII Detected" but the request failed with a 500 error
- **Cause**: The Vault is likely full or the disk is read-only.
- **Fix**: Check `OCU_VAULT_PATH` permissions and available disk space. If using DuckDB, ensure no other process is locking the file.

### 3. "Deep AI Scrub" is disabled in the Dashboard
- **Cause**: The `slm-engine` is not reachable.
- **Fix**: Verify `SLM_SIDECAR_URL` is correct. Check if the AI container is running: `docker ps | grep slm-engine`.

---

## 📈 Performance Bottlenecks

### Slow Refinement
- **Diagnosis**: Check the logs for `REF_SCAN` duration.
- **Reason**: Large files or complex AI NER scans (Tier 2).
- **Optimization**: Upgrade hardware (more CPUs/RAM) or disable Tier 2 if not strictly required for the use case.

### High Memory Usage
- **Diagnosis**: `docker stats`.
- **Reason**: The `slm-engine` loads a large model into memory (~1.5GB).
- **Optimization**: Ensure the host has enough RAM. If running on a laptop, close other memory-intensive apps.

---

## 🧪 Advanced Debugging

### Inspecting the Vault (DuckDB)
You can query the vault directly to see what tokens have been generated (encrypted).
```bash
duckdb /path/to/vault.db "SELECT * FROM tokens LIMIT 10;"
```

### Forcing a Rule Match
To verify if a specific regex is working, send a payload directly to the Refinery:
```bash
curl -s -X POST http://localhost:4141/api/refine \
  -H "Content-Type: application/json" \
  -d '{"text": "My email is test@example.com", "actor": "debug"}'
```
Check if the `refined` field in the response contains `[EMAIL_...]`.
