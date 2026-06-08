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
The SharePoint connector ingests documents from Microsoft SharePoint via the Microsoft Graph API using OAuth2 Client Credentials. It authenticates as your Azure AD application, polls the drive for new and changed files using delta queries (incremental sync), extracts their text content, and passes everything through the Zero-Egress Refinery before any downstream use.

> [!IMPORTANT]
> This connector requires a **Pro/Enterprise License** with the SharePoint bitmask (Bit 1) enabled.

**Azure AD setup (one-time):**
1. Register an app in Azure AD → **App registrations → New registration**.
2. Under **Certificates & secrets**, create a client secret. Note the value.
3. Under **API permissions**, add `Sites.Read.All` (Microsoft Graph, Application type) and grant admin consent.
4. Note the **Application (client) ID** and **Directory (tenant) ID**.

**Configuration:**
```yaml
connectors:
  - id: sharepoint-prod
    type: sharepoint-graph
    config:
      tenant_id: "your-tenant-id"      # Azure AD Directory ID
      client_id: "your-client-id"      # Azure AD Application ID
      client_secret: "your-client-secret"
      site_id: "your-site-id"          # SharePoint site ID (optional; required for Fetch)
```

**Supported file formats:**

| Format | Extension | Extraction method |
|---|---|---|
| Plain text | `.txt`, `.csv`, `.json`, `.md`, `.log`, `.xml`, `.html`, `.eml` | Raw content |
| Word documents | `.docx` | ZIP/XML — extracts all `<w:t>` run text |
| Excel workbooks | `.xlsx` | ZIP/XML — extracts shared strings table |
| PowerPoint | `.pptx` | ZIP/XML — extracts all slide `<a:t>` runs |
| PDF | `.pdf` | Byte-scan — extracts literal string operands from text blocks |

> [!NOTE]
> PDF extraction covers standard Latin-character PDFs. Fully-encrypted PDFs and CID-keyed fonts (common in scanned/image-only PDFs) will produce empty output and are skipped with a log warning.

**How it works:**
- On startup, a background goroutine polls every 30 seconds using `GET /sites/{id}/drive/root/delta`. The first run performs a full enumeration; subsequent runs fetch only changed files (delta links are persisted in memory).
- Files whose extension is not in the supported list are silently skipped.
- All extracted text is passed to `refinery.ProcessInterface` before any logging or forwarding — no raw PII ever leaves the connector.

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
