# Ocultar

[![Apache 2.0](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go 1.25+](https://img.shields.io/badge/go-1.25%2B-00ADD8.svg)](https://go.dev)

Ocultar is an open-source local PII/PHI masking engine for AI workflows.

It runs as a local HTTP sidecar. Send it text before it reaches a cloud LLM; it returns the
same text with every piece of personal data replaced by a deterministic, reversible token
(`[EMAIL_9c8f7a1b]`, `[PERSON_3a12b4cd]`, …). Originals are encrypted and stored in a
local vault. Callers with the auditor token can restore them.

No PII ever reaches the upstream model.

**Built on Ocultar:** [Ki!](https://getki.ai) — desktop app for private AI workflows.

---

## Quick start — Docker

```bash
docker run --rm -p 4141:4141 \
  -e OCU_MASTER_KEY=<64-hex-chars> \
  -e OCU_SALT=<32-hex-chars> \
  -e OCU_AUDITOR_TOKEN=<your-secret-token> \
  ghcr.io/edu963/ocultar:latest -serve 4141
```

## Quick start — prebuilt binary

```bash
OCU_MASTER_KEY=<64-hex> OCU_SALT=<32-hex> OCU_AUDITOR_TOKEN=<token> ./ocultar -serve 4141
```

Generate the key and salt once and store them securely:

```bash
OCU_MASTER_KEY=$(openssl rand -hex 32)
OCU_SALT=$(openssl rand -hex 16)
OCU_AUDITOR_TOKEN=$(openssl rand -hex 24)
```

---

## API reference

### `GET /api/health`

Returns engine status. No authentication required.

```json
{
  "status": "healthy",
  "version": "1.14",
  "vault": { "status": "online" },
  "slm":   { "status": "online", "circuit": "closed" }
}
```

---

### `POST /api/refine`

Mask PII in text or JSON. No authentication required.

**Request body**: raw text string or any JSON value.

**Response**:

```json
{
  "refined": "{\"message\":\"Hello [PERSON_3a12b4cd], your order [EMAIL_9c8f7a1b] is ready.\"}",
  "report": {
    "hits": 2,
    "types": ["PERSON", "EMAIL"]
  }
}
```

> `refined` is a JSON-encoded string — parse it once to get the masked payload.

---

### `POST /api/reveal`

Restore vault tokens back to originals.

**Authentication**: `Authorization: Bearer <OCU_AUDITOR_TOKEN>` header required.
Returns `403` if `OCU_AUDITOR_TOKEN` is not set on the server.

**Request body**:

```json
{ "tokens": ["[PERSON_3a12b4cd]", "[EMAIL_9c8f7a1b]"] }
```

**Response**:

```json
{
  "results": {
    "[PERSON_3a12b4cd]": "Alice Martin",
    "[EMAIL_9c8f7a1b]": "alice@example.com"
  }
}
```

---

### `GET /api/entities` · `POST /api/entities` · `POST /api/entities/seed`

Manage the persistent entity registry (pre-seed canonical names so all variants map to the
same token). Requires `Authorization: Bearer <OCU_AUDITOR_TOKEN>`.

---

## Architecture

Ocultar runs two detection tiers before any text leaves the machine:

### Tier 1 — Deterministic regex / heuristics (fast, zero-egress)

| Sub-tier | Shield | What it catches |
|----------|--------|-----------------|
| 0 | Dictionary | VIP names, org names from `configs/protected_entities.json` |
| 0.5 | Pattern + Entropy | High-entropy strings (API keys, secrets) via Shannon scoring |
| 1 | Rule Engine | EMAIL, SSN, IBAN, credit cards, 50+ national ID formats |
| 1.1 | Phone Shield | libphonenumber validation |
| 1.2 | Address Shield | Heuristic street address parser (EN/FR/ES/DE) |
| 1.5 | Contextual | Names in greetings, signatures, interrogative sentences |

### Tier 2 — SLM-based NER (higher recall, configurable endpoint)

Sends text to a local AI sidecar for named-entity recognition. The scanner is always
initialized but produces no results unless a compatible sidecar is running at `SLM_SIDECAR_URL`.
Point it at a [privacy-filter](https://huggingface.co/openai/privacy-filter) or llama.cpp instance to activate NER.

```bash
SLM_SIDECAR_URL=http://localhost:8085 ./ocultar -serve 4141
```

Use `SLM_ADAPTER=openai-chat` for a llama.cpp / Qwen endpoint, or leave unset for the
privacy-filter protocol (default).

---

## Configuration

| Variable | Required | Default | Purpose |
|---|---|---|---|
| `OCU_MASTER_KEY` | Yes (production) | insecure dev key | 32+ byte AES key material for HKDF |
| `OCU_SALT` | Yes (production) | built-in default | Per-deployment HKDF salt |
| `OCU_AUDITOR_TOKEN` | Yes | — | Bearer token for `/api/reveal` and `/api/entities` |
| `OCU_VAULT_PATH` | No | `vault.db` | DuckDB vault file path |
| `SLM_SIDECAR_URL` | No | `http://localhost:8085` | Tier 2 NER sidecar endpoint |
| `SLM_ADAPTER` | No | `privacy-filter` | Sidecar protocol: `privacy-filter` or `openai-chat` |

---

## Building from source

Requires Go 1.25+ with CGO enabled (DuckDB and libphonenumber need a C compiler).

```bash
git clone https://github.com/Edu963/ocultar.git
cd ocultar
make build
```

Run tests:

```bash
CGO_ENABLED=1 go test ./...
```

---

## License

Apache 2.0 — see [LICENSE](LICENSE).

Ki! (the desktop app) is a separate commercial product built on Ocultar: [getki.ai](https://getki.ai).
