## Ocultar PII Refinery — Goose Extension

Zero-egress PII protection for Goose AI workflows.
Runs entirely in your infrastructure — no data leaves your environment.

### Prerequisites
- Ocultar running locally (docker compose -f docker-compose.community.yml up)
- Python 3.10+

### Installation

```bash
pip install ocultar-goose-mcp
```

### Configuration in Goose
1. Open Goose settings
2. Add Extension → Command-line Extension
3. Name: ocultar-pii
4. Command: ocultar-goose-mcp
5. Environment:
   OCULTAR_URL=http://localhost:8080
   OCULTAR_API_KEY=your-key

### Usage
Ask Goose: 'Refine this text before processing: [text with PII]'
Goose will call ocultar-pii which redacts PII locally before
any further processing.

### Why local-only?
The zero-egress guarantee means your sensitive data never leaves
your infrastructure. The MCP server runs on stdio — no network
server, no remote calls, no supply chain attack surface.
