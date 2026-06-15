## Ocultar PII Refinery — Goose Extension

Zero-egress PII protection for Goose AI workflows.
Runs entirely in your infrastructure — no data leaves your environment.

### Prerequisites
- Ocultar running locally (see Quick Start below)
- Python 3.10+

### Quick Start — Docker

```bash
docker run --rm -p 4141:4141 \
  -e OCU_MASTER_KEY=<64-hex-chars> \
  -e OCU_SALT=<32-hex-chars> \
  -e OCU_AUDITOR_TOKEN=<your-secret-token> \
  ghcr.io/ocultar-dev/ocultar:latest -serve 4141
```

### Installation

```bash
pip install ocultar-goose-mcp
```

### Configuration in Goose
1. Open Goose settings
2. Add Extension → Command-line Extension
3. Name: `ocultar-pii`
4. Command: `ocultar-goose-mcp`
5. Environment:
   ```
   OCULTAR_URL=http://localhost:4141
   ```

### Usage
Ask Goose: `Refine this text before processing: [text with PII]`

Goose will call `ocultar-pii` which redacts PII locally before any further processing.
The tool returns the cleaned text with each PII value replaced by a deterministic token
(e.g. `[EMAIL_9c8f7a1b]`).

### Available tools

| Tool | Description |
|---|---|
| `refine_text` | Redacts PII from text before AI processing |

> **Note:** `reveal_tokens` (de-tokenization) is intentionally omitted from this extension.
> Token reveal is an auditor-only operation and is not suitable for automated agent workflows.
> Use the Claude MCP extension if you need reveal access.

### Why local-only?
The zero-egress design means your sensitive data never leaves your infrastructure.
The MCP server runs on stdio — no network server, no remote calls, no supply chain attack surface.
If Ocultar is unreachable, the extension returns an error and withholds the raw text — it never
forwards unmasked data as a fallback.
