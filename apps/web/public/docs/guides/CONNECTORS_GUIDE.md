# OCULTAR | Connectors Guide

> **Audience:** DevOps engineers and developers who need to ingest data from external platforms (Slack, SharePoint, etc.) into OCULTAR.

---

## 1. Overview

OCULTAR Connectors are modular ingestion components that fetch or receive data from external sources and feed it into the Zero-Egress Refinery. This ensures that data from your enterprise collaboration tools is sanitized before it reaches any LLM or is stored in your vault.

## 2. Supported Connectors

### 2.1 Slack Workspace (Enterprise ✦)
The Slack connector allows you to ingest channel history and listen for message events.

> [!IMPORTANT]
> This connector requires an **Enterprise License** with the Slack capability (Bit 0) enabled.

**Configuration:**
```yaml
connectors:
  - id: slack-main
    type: slack
    config:
      workspace_id: T12345678
      token: "xoxb-your-slack-bot-token"
```

### 2.2 Microsoft SharePoint & Teams (Pro ✦)
The SharePoint connector allows you to ingest documents and messages from Microsoft Graph API. It ensures all data is tokenized before being processed by any AI models.

> [!IMPORTANT]
> This connector requires a **Pro/Enterprise License** with the SharePoint bitmask (Bit 1) enabled.

**Configuration:**
```yaml
connectors:
  - id: sharepoint-prod
    type: sharepoint-graph
    config:
      tenant_id: "your-tenant-id"
      client_id: "your-client-id"
      client_secret: "your-client-secret"
      site_id: "your-site-id" # Optional
```

### 2.3 Dynamic Plugins (Enterprise ✦)
Enterprise users can load custom connectors as Go plugins (`.so` files).

**Example:**
```yaml
connectors:
  - id: custom-source
    type: plugin
    path: "/path/to/your/connector.so"
```

## 3. Configuration

Connectors are configured via environment variables (for basic usage) or via a `connectors` section in the enterprise configuration file.

### Environment Variables
- `SLACK_TOKEN`: The API token for your Slack bot.
- `SLACK_WORKSPACE_ID`: Your Slack Workspace ID.

## 4. Zero-Egress Implementation

Every connector follows the **Refinery-First** principle:
1. Data is fetched from the source (e.g., Slack API).
2. Data is immediately passed to `pkg/refinery.ProcessInterface`.
3. Only the **refined** (redacted) data is logged or forwarded.
4. All secrets (API keys, tokens) stay within the secure OCULTAR environment.

## 5. Development

To build a new connector, implement the `Connector` interface in `services/refinery/pkg/connector`:

```go
type Connector interface {
    ID() string
    Type() string
    Init(config map[string]interface{}, eng *refinery.Refinery) error
    Start() error
    Stop() error
    Fetch(ctx context.Context, params map[string]interface{}) ([]byte, error)
}
```

Register your connector in its `init()` function:
```go
func init() {
    connector.Register("my-connector", func() connector.Connector {
        return &MyConnector{}
    })
}
```
