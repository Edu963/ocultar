# Ocultar PII Refinery — Mistral Le Chat MCP Extension

Zero-egress PII protection for Mistral Le Chat workflows.
Runs entirely in your infrastructure — no data ever leaves your environment.

Optimised for French and EU compliance: SIRET/SIREN, IBAN, RGPD Article 25, CNIL requirements.

## Tools

| Tool | Description |
|------|-------------|
| `refine_text` | Redacts PII / DCP before sending text to Le Chat. Returns clean text + token map. |
| `reveal_tokens` | De-tokenizes tokens back to plaintext (auditor-only, requires `OCULTAR_AUDITOR_TOKEN`). |

## Prerequisites

- Ocultar Refinery running locally:
  ```bash
  docker compose -f docker-compose.community.yml up
  ```
- Python 3.10+

## Installation

```bash
pip install ocultar-mistral-mcp
```

Or with `uvx` (no install needed):
```bash
uvx ocultar-mistral-mcp
```

## Mistral Le Chat Configuration

In Mistral Le Chat, open **Settings → Tools → MCP Servers** and add:

```json
{
  "mcpServers": {
    "ocultar-pii": {
      "command": "ocultar-mistral-mcp",
      "env": {
        "OCULTAR_URL": "http://localhost:8080",
        "OCULTAR_API_KEY": "your-api-key"
      }
    }
  }
}
```

Or with `uvx` (no prior install required):

```json
{
  "mcpServers": {
    "ocultar-pii": {
      "command": "uvx",
      "args": ["ocultar-mistral-mcp"],
      "env": {
        "OCULTAR_URL": "http://localhost:8080",
        "OCULTAR_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `OCULTAR_URL` | Yes | URL of your local Ocultar Refinery (default: `http://localhost:8080`) |
| `OCULTAR_API_KEY` | No | Bearer token for Refinery auth |
| `OCULTAR_AUDITOR_TOKEN` | No | Enables `reveal_tokens` — must match `OCU_AUDITOR_TOKEN` on the server |

## Usage

Once connected, Le Chat will automatically call `refine_text` before processing sensitive data. You can also invoke it explicitly:

> "Avant de traiter ce texte, filtre les DCP : Jean Dupont, jean.dupont@banque.fr, IBAN FR76 3000 6000 0112 3456 7890 189"

Le Chat returns:
```json
{
  "cleanText": "[NAME_a1b2c3d4], [EMAIL_9c8f7a1b], IBAN [IBAN_7f3e9a2b]",
  "tokenMap": {
    "[NAME_a1b2c3d4]": "NAME",
    "[EMAIL_9c8f7a1b]": "EMAIL",
    "[IBAN_7f3e9a2b]": "IBAN"
  }
}
```

For authorized workflows that need to restore PII after AI processing:

> "Reveal these tokens: [EMAIL_9c8f7a1b], [IBAN_7f3e9a2b]"

This call is recorded in the immutable Ed25519-signed audit log.

## Why This Matters for French Enterprises

Sending customer data to any external AI API — including Mistral's cloud — without redaction constitutes a RGPD violation under Article 25 (Privacy by Design). The CNIL has issued enforcement guidance specifically targeting AI pipeline data flows.

Ocultar ensures that:
- No raw PII ever reaches Le Chat's API endpoint
- SIRET, SIREN, IBAN, and French address formats are detected and tokenized
- Every vault access is logged in a tamper-evident, Ed25519-signed audit trail
- You remain the data controller — Ocultar is a local processor under your full control

## Security Model

- `refine_text` is safe to expose to any Le Chat session
- `reveal_tokens` requires `OCULTAR_AUDITOR_TOKEN` and every call is logged with actor, timestamp, and Ed25519 signature
- The Refinery vault uses AES-256-GCM with HKDF-SHA256 key derivation — tokens are useless without the master key
- If the Refinery is unreachable, both tools fail closed — raw PII is never forwarded

## License

Apache 2.0 — see [LICENSE](../../LICENSE)
