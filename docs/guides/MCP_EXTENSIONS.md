# OCULTAR | MCP Extensions

OCULTAR ships three MCP (Model Context Protocol) extensions that plug directly into AI clients. Each extension exposes two tools — `refine_text` and `reveal_tokens` — and enforces the zero-egress guarantee at the protocol level: if the local Refinery is unreachable, both tools fail closed and refuse to forward any data.

---

## Supported Clients

| Extension | Client | Package |
|---|---|---|
| **ocultar-claude-mcp** | Claude Desktop · Claude Code CLI | `pip install ocultar-claude-mcp` |
| **ocultar-goose-mcp** | Goose AI | `pip install ocultar-goose-mcp` |
| **ocultar-mistral-mcp** | Mistral Le Chat | `pip install ocultar-mistral-mcp` |

All three packages require **Python 3.10+** and the OCULTAR Refinery running locally on port 8080.

---

## Prerequisites

Start the Refinery before connecting any MCP client:

```bash
docker compose -f docker-compose.community.yml up
```

Or manually:

```bash
go run ./services/refinery/cmd/main.go --serve 8080
```

---

## Tools

Both tools are available in all three extensions.

### `refine_text`

Redacts PII from text before it reaches the AI model. Returns the cleaned text and a map of every token to its PII type.

**Input**
```json
{ "text": "Jean Dupont, jean.dupont@banque.fr, IBAN FR76 3000 6000 0112 3456 7890 189" }
```

**Output**
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

Safe to expose to any AI session. No PII ever leaves your infrastructure.

### `reveal_tokens`

De-tokenizes specific tokens back to plaintext. Requires `OCULTAR_AUDITOR_TOKEN` — auditor-only. Every call is recorded in the immutable Ed25519-signed audit log with actor identity and timestamp.

**Input**
```json
{ "tokens": ["[EMAIL_9c8f7a1b]", "[IBAN_7f3e9a2b]"] }
```

---

## Claude Desktop

**Install:**
```bash
pip install ocultar-claude-mcp
```

**Configure** — add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "ocultar-pii": {
      "command": "ocultar-claude-mcp",
      "env": {
        "OCULTAR_URL": "http://localhost:8080",
        "OCULTAR_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Claude Code (CLI)

```bash
claude mcp add ocultar-pii -- ocultar-claude-mcp
```

Or add the same block above to `.claude/settings.json`.

---

## Goose AI

**Install:**
```bash
pip install ocultar-goose-mcp
```

**Configure in Goose:**
1. Open Goose Settings
2. Add Extension → Command-line Extension
3. Name: `ocultar-pii`
4. Command: `ocultar-goose-mcp`
5. Environment: `OCULTAR_URL=http://localhost:8080`

---

## Mistral Le Chat

Optimised for French and EU compliance: SIRET/SIREN, IBAN, French phone numbers, RGPD Article 25, CNIL requirements.

**Install:**
```bash
pip install ocultar-mistral-mcp
```

Or with `uvx` (no install needed):
```bash
uvx ocultar-mistral-mcp
```

**Configure** — in Mistral Le Chat, open **Settings → Tools → MCP Servers** and add:

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

With `uvx`:
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

### Why This Matters for French Enterprises

Sending customer data to any external AI API — including Mistral's cloud — without redaction constitutes an RGPD violation under Article 25 (Privacy by Design). The CNIL has issued enforcement guidance specifically targeting AI pipeline data flows.

OCULTAR ensures that:
- No raw PII ever reaches Le Chat's API endpoint
- SIRET, SIREN, IBAN, and French address formats are detected and tokenized
- Every vault access is logged in a tamper-evident, Ed25519-signed audit trail
- You remain the data controller — OCULTAR is a local processor under your full control

---

## Environment Variables

All three extensions share the same environment variables:

| Variable | Required | Default | Description |
|---|---|---|---|
| `OCULTAR_URL` | Yes | `http://localhost:8080` | URL of your local OCULTAR Refinery |
| `OCULTAR_API_KEY` | No | — | Bearer token for Refinery authentication |
| `OCULTAR_AUDITOR_TOKEN` | No | — | Enables `reveal_tokens` — must match `OCU_AUDITOR_TOKEN` on the server |

---

## Security Model

- `refine_text` is safe to expose to any AI session — it only sends text to the local Refinery, which runs on `localhost`. No telemetry, no remote calls.
- `reveal_tokens` requires `OCULTAR_AUDITOR_TOKEN`. Every call is logged with actor identity, timestamp, and Ed25519 signature in the tamper-proof audit trail.
- The Refinery vault uses AES-256-GCM with HKDF-SHA256 key derivation — tokens are useless without the master key.
- **Fail-closed guarantee:** if the Refinery is unreachable for any reason, both tools return an MCP error and refuse to forward raw data or vault contents to the caller.

---

## License

Apache 2.0
