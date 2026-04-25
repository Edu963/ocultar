# Sombra | Agentic Data Gateway

Sombra is the privacy-preserving gateway for OCULTAR. It enables secure connections to external LLMs and data sources (Slack, SharePoint) while enforcing strict data sovereignty policies.

## Key Features

- **Multi-Model Routing**: Route queries to local SLMs, OpenAI, or Gemini.
- **Fail-Closed Redaction**: Automatically pseudonymize PII using the OCULTAR refinery.
- **Policy Enforcement**: Strip high-sensitivity data (SSN, Passwords) before they ever reach the AI.
- **Connector Ecosystem**: Native support for File uploads, Banking APIs, and Slack.

## Configuration

Sombra is configured via `configs/sombra.yaml`.

### Data Policies

Each connector can have a `policy` block:

```yaml
policy:
  strip_categories: ["SSN", "CREDENTIAL", "SECRET"]
  allowed_models: ["local-slm", "gpt-4o"]
```

- **`strip_categories`**: List of PII types to REMOVE from the data context.
- **`allowed_models`**: Restricts which AI models can process data from this connector.

## Getting Started

1.  **Configure environment**:
    ```bash
    cp ../../.env.example .env && source .env
    ```

2.  **Run with local entry point**:
    ```bash
    go run main.go
    ```

3.  **Test a query**:
    ```bash
    curl -X POST http://localhost:8081/query \
      -F "model=gemini-flash-latest" \
      -F "connector=slack-prod" \
      -F "prompt=Summarize my day." \
      -F "channel_id=YOUR_CHANNEL_ID"
    ```

## Documentation

- [PII Detection & Compliance](../../docs/TECH_DOCS.md)
- [Ecosystem Manifest](../../ecosystem.manifest.json)
