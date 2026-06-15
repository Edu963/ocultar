# Deployment Guide

---

## Quick start — Docker

The fastest way to run Ocultar:

```bash
docker run --rm -p 4141:4141 \
  -e OCU_MASTER_KEY=$(openssl rand -hex 32) \
  -e OCU_SALT=$(openssl rand -hex 16) \
  -e OCU_AUDITOR_TOKEN=$(openssl rand -hex 24) \
  ghcr.io/ocultar-dev/ocultar:latest -serve 4141
```

Test it:

```bash
curl -s -X POST http://localhost:4141/api/refine \
  -H "Content-Type: text/plain" \
  --data "My name is Alice Martin and my email is alice@example.com"
```

---

## Production deployment

### 1. Generate and store secrets

Generate the three required secrets once and store them in your secret manager or environment:

```bash
OCU_MASTER_KEY=$(openssl rand -hex 32)   # 32-byte AES key material
OCU_SALT=$(openssl rand -hex 16)          # 16-byte HKDF salt
OCU_AUDITOR_TOKEN=$(openssl rand -hex 24) # auditor bearer token
```

These must be treated as credentials:
- `OCU_MASTER_KEY` encrypts the vault. Losing it means losing access to all stored tokens.
- Rotating `OCU_MASTER_KEY` invalidates the existing vault — all previously issued tokens become irrecoverable.
- `OCU_AUDITOR_TOKEN` grants access to `POST /api/reveal`. Restrict it to authorized principals only.

### 2. Persist the vault

The vault (`vault.db`) must be persisted across container restarts. Mount a volume:

```bash
docker run -d \
  -p 4141:4141 \
  -e OCU_MASTER_KEY=<your-key> \
  -e OCU_SALT=<your-salt> \
  -e OCU_AUDITOR_TOKEN=<your-token> \
  -e OCU_VAULT_PATH=/data/vault.db \
  -v /var/lib/ocultar:/data \
  ghcr.io/ocultar-dev/ocultar:latest -serve 4141
```

### 3. Verify

```bash
curl http://localhost:4141/api/health
```

Expected response:

```json
{ "status": "healthy", "version": "1.14", "vault": { "status": "online" } }
```

---

## Configuration reference

| Variable | Required | Default | Purpose |
|---|---|---|---|
| `OCU_MASTER_KEY` | Yes (production) | insecure dev key | 32+ byte hex key material for HKDF |
| `OCU_SALT` | Yes (production) | built-in default | Per-deployment HKDF salt |
| `OCU_AUDITOR_TOKEN` | Yes | — | Bearer token for `/api/reveal` and `/api/entities` |
| `OCU_VAULT_PATH` | No | `vault.db` | Path to DuckDB vault file |
| `SLM_SIDECAR_URL` | No | `http://localhost:8085` | Tier 2 NER sidecar endpoint |
| `SLM_ADAPTER` | No | `privacy-filter` | Sidecar protocol: `privacy-filter` or `openai-chat` |

If `OCU_MASTER_KEY` or `OCU_SALT` are not set, Ocultar starts with insecure development defaults and logs a warning. Never run in production without setting both.

---

## Enabling Tier 2 AI NER

Tier 2 sends text to a local SLM for higher-recall named-entity recognition. It is optional — Ocultar runs fully without it.

To activate:

```bash
docker run -d \
  -p 4141:4141 \
  -e OCU_MASTER_KEY=<key> \
  -e OCU_SALT=<salt> \
  -e OCU_AUDITOR_TOKEN=<token> \
  -e SLM_SIDECAR_URL=http://your-slm-host:8085 \
  -e SLM_ADAPTER=openai-chat \
  ghcr.io/ocultar-dev/ocultar:latest -serve 4141
```

Compatible sidecar endpoints:
- Any llama.cpp server with an OpenAI-compatible `/v1/chat/completions` endpoint (`SLM_ADAPTER=openai-chat`)
- A [privacy-filter](https://huggingface.co/openai/privacy-filter) model endpoint (`SLM_ADAPTER=privacy-filter`, default)

If the SLM sidecar is unavailable, the circuit breaker opens and Ocultar falls back to Tier 1 only — it does not block requests.

---

## Building from source

Requires Go 1.24+ with CGO enabled. A C compiler is required for DuckDB and libphonenumber.

```bash
# Install GCC (Ubuntu/Debian)
sudo apt-get install -y gcc

# Clone and build
git clone https://github.com/ocultar-dev/ocultar.git
cd ocultar
CGO_ENABLED=1 go build -o ocultar ./services/refinery/cmd
./ocultar -serve 4141
```

Run tests:

```bash
CGO_ENABLED=1 go test ./...
```

---

## Key rotation

Rotating `OCU_MASTER_KEY` invalidates the existing vault. Plan accordingly:

1. Complete any in-flight reveal operations before rotation.
2. Stop Ocultar.
3. Back up `vault.db`.
4. Delete `vault.db` (the old key cannot be migrated — a new vault starts clean).
5. Generate a new `OCU_MASTER_KEY` and `OCU_SALT`.
6. Restart Ocultar with the new secrets.
7. Re-seed the entity registry if you use pre-seeded entities (`POST /api/entities/seed`).

Re-seeding is not required for normal operation — the vault will repopulate as new text is refined.

---

## Health monitoring

Poll `GET /api/health` for liveness and readiness checks:

```bash
curl -sf http://localhost:4141/api/health | jq .status
```

The endpoint returns `"healthy"` when the vault is online and the engine is ready. It returns `"degraded"` if the Tier 2 SLM circuit breaker has opened (Tier 1 still active).

Recommended alerting thresholds:
- Alert if `vault.status` is not `"online"` — Ocultar will block all refine requests.
- Warn if `slm.circuit` is `"open"` — Tier 2 is down, recall may be lower.

---

## Docker Compose example

```yaml
services:
  ocultar:
    image: ghcr.io/ocultar-dev/ocultar:latest
    command: ["-serve", "4141"]
    ports:
      - "4141:4141"
    environment:
      OCU_MASTER_KEY: ${OCU_MASTER_KEY}
      OCU_SALT: ${OCU_SALT}
      OCU_AUDITOR_TOKEN: ${OCU_AUDITOR_TOKEN}
      OCU_VAULT_PATH: /data/vault.db
    volumes:
      - ocultar-vault:/data
    restart: unless-stopped

volumes:
  ocultar-vault:
```

Store `OCU_MASTER_KEY`, `OCU_SALT`, and `OCU_AUDITOR_TOKEN` in a `.env` file (never committed) or inject them from your secret manager.
