# OCULTAR | API Reference

> **Audience:** Go developers embedding OCULTAR as a library, and operators calling the HTTP API.
> All types come from `github.com/Edu963/ocultar`.

---

## Table of Contents

1. [Environment Variables](#1-environment-variables)
2. [Refinery Package (`internal/pii`)](#2. Refinery Package (`internal/pii`)
   - [Types & Interfaces](#21-types--interfaces)
   - [Constructor](#22-constructor)
   - [Methods](#23-methods)
   - [Cryptography Helpers](#24-cryptography-helpers)
3. [Vault Package (`services/vault`)](#3-vault-package-servicesvault)
4. [Config Package (`services/refinery/pkg/config`)](#4-config-package-servicesrefinerypkgconfig)
5. [Proxy Package (`services/refinery/pkg/proxy`)](#5-proxy-package-servicesrefinerypkgproxy)
6. [HTTP Endpoints](#6-http-endpoints)
   - [POST /api/refine](#61-post-apirefine)
   - [POST /api/refine/file](#62-post-apirefinefile)
7. [HTTP Proxy Mode](#7-http-proxy-mode)
8. [Error Reference](#8-error-reference)

---

## 1. Environment Variables

All environment variables are read at startup by the `main` entrypoint. No variable is optional unless marked as such.

| Variable | Required | Description | Example |
|---|---|---|---|
| `OCU_MASTER_KEY` | ✅ | AES-256 master key — any UTF-8 string; SHA-256-hashed internally to 32 bytes. **Never commit this.** | `openssl rand -hex 32` |
| `OCU_LICENSE_KEY` | Enterprise only | Activates Enterprise tier features (Tier 2 AI scan, PostgreSQL vault, SIEM audit log). Absent → Community tier. | `ENTER-PRISE-KEY-HERE` |
| `OCU_PROXY_TARGET` | Proxy mode | Default upstream URL the proxy forwards sanitised requests to. | `https://api.openai.com` |
| `OCU_PROXY_PORT` | Proxy mode | Port the proxy listener binds to. Defaults to `8080`. | `8080` |
| `OCU_VAULT_PATH` | Optional | Override the DuckDB vault file path. Defaults to `vault.db`. Use `:memory:` for ephemeral operation. | `/data/vault.db` |
| `OCU_PILOT_MODE` | Optional | Set to `true` to enforce a hard vault entry cap (pilot/trial limitation). | `true` |
| `OCU_SALT` | Optional | Extra entropy injected into key derivation in the Sombra gateway (`github.com/Edu963/sombra`). Not used by the core refinery. | `random-string` |

---

## 2. Refinery Package (`internal/pii`)

### 2.1 Types & Interfaces

#### `Refinery`

```go
type Refinery struct {
    Vault       vault.Provider
    MasterKey   []byte
    DryRun      bool
    Report      bool
    Serve       string
    PilotMode   bool
    VaultCount  *atomic.Int64
    AuditLogger AuditLogger
    AIScanner   AIScanner
}
```

| Field | Description |
|---|---|
| `Vault` | Storage backend (DuckDB or PostgreSQL). Must not be `nil`. |
| `MasterKey` | 32-byte AES key (after SHA-256 hashing of the raw env value). |
| `DryRun` | When `true`, PII is detected and reported but no tokens are written to the vault. |
| `Report` | When `true`, per-request PII hit metadata is accumulated for `GenerateReport`. |
| `Serve` | Non-empty string when running in HTTP serve mode (e.g. `"9090"`). Activates hit tracking. |
| `PilotMode` | Enforces a vault entry cap for trial/pilot use. |
| `VaultCount` | Atomic counter of vault entries — shared with the dashboard for live metrics. |
| `AuditLogger` | SIEM audit logger. Defaults to `NoopAuditLogger` (Community). |
| `AIScanner` | Local SLM scanner for Tier 2 NER. Defaults to `NoopAIScanner` (Community). |

---

#### `AuditLogger` interface

```go
type AuditLogger interface {
    Init(filePath string) error
    Log(user, action, result string)
    Close()
}
```

| Method | Description |
|---|---|
| `Init(filePath string) error` | Opens or creates the SIEM log file at `filePath`. Must be called before `Log`. |
| `Log(user, action, result string)` | Appends a structured JSON line to the log. `action` is `"vaulted"` or `"matched"`. |
| `Close()` | Flushes and closes the log file. |

Community default: `NoopAuditLogger` — all methods are no-ops.

---

#### `AIScanner` interface

```go
type AIScanner interface {
    ScanForPII(text string) (map[string][]string, error)
    CheckHealth(host string)
    IsAvailable() bool
}
```

| Method | Description |
|---|---|
| `ScanForPII(text string) (map[string][]string, error)` | Sends `text` to the local SLM and returns a map of `PII_TYPE → []detected_string`. Returns `nil, nil` when no PII is found. |
| `CheckHealth(host string)` | Pings the SLM host. Sets internal availability flag. |
| `IsAvailable() bool` | Returns `true` only when the SLM is healthy and responding. Always `false` in Community. |

Community default: `NoopAIScanner` — `IsAvailable()` always returns `false`.

---

#### `DryRunReport`

```go
type DryRunReport struct {
    Mode       string              `json:"mode"`
    FilesIn    int                 `json:"files_scanned"`
    Hits       map[string][]string `json:"pii_hits"`
    TotalCount int                 `json:"total_pii_count"`
    Blocking   bool                `json:"blocking"`
}
```

| Field | Description |
|---|---|
| `Mode` | `"dry-run"`, `"report"`, or `"serve"` depending on refinery configuration. |
| `FilesIn` | Number of files/payloads processed in the session. |
| `Hits` | Map of PII type → list of truncated hashes (e.g. `{"EMAIL": ["af2101fb…"]}`). |
| `TotalCount` | Sum of all PII hits across all types. |
| `Blocking` | `true` when at least one PII entity was detected. Useful as a CI/CD gate. |

---

### 2.2 Constructor

#### `NewRefinery`

```go
func NewRefinery() *Refinery
```

Creates a `Refinery` using the default `pii.Registry`.

**Returns:** `*Refinery` — ready to use; never `nil`.

---

### 2.3 Methods

#### `Scan`

```go
func (e *Refinery) Scan(input string) []DetectionResult
```

Core detection function. Performs exhaustive deterministic sweep with validation.

#### `Redact`

```go
func (e *Refinery) Redact(input string, tokenFunc func(DetectionResult) (string, error)) (string, error)
```

Uses `Scan` to find PII and calls the `tokenFunc` to replace PII with tokens.

**Pipeline order:**

| Tier | Shield | Source |
|---|---|---|
| 0.1 | Embedded Base64 Evasion | Inline in `RefineString` |
| 0.5 | Dictionary Shield | `config.Global.Dictionaries` |
| 1 | Dynamic Regex Shields | `config.Global.Regexes` (email, URL, custom) |
| 1.1 | Phone Shield | `pkg/refinery/phone_parser.go` (`libphonenumber`) |
| 1.2 | Address Shield | `pkg/refinery/address_parser.go` |
| 1.5 | Greeting & Signature Shield | `greetingRegex` |
| 2 | AI NER Scan | `AIScanner` (Enterprise only) |

**Parameters:**
- `input` — raw string to redact.
- `actor` — identifier of the originating client (IP or request ID) for audit logging.
- `preScanMap` — pre-computed SLM results from a parent `ProcessInterface` call; pass `nil` for standalone use.

**Returns:** `(redacted string, error)` — error only on vault failure or SLM inference error.

> **Note:** `RefineString` returns the *best-effort* output on non-fatal errors. A returned `error` is always fatal — the caller must not persist the output.

---

#### `ProcessInterface`

```go
func (e *Refinery) ProcessInterface(data interface{}, actor string) (interface{}, error)
```

Recursively traverses a JSON-decoded Go value (`map[string]interface{}`, `[]interface{}`, or `string`). Handles:
- **Base64-encoded strings** — decoded, refined, re-encoded.
- **URL-encoded strings** — decoded, refined, re-encoded with `url.QueryEscape`.
- **Nested JSON strings** — un-marshalled, refined recursively, re-marshalled.
- **SLM batch optimisation** — for complex objects, marshals the entire record to a flat string and runs one SLM scan before recursing.

**Parameters:**
- `data` — the decoded JSON value (root or node). Pass the output of `json.Unmarshal`.
- `actor` — see `RefineString`.

**Returns:** `(interface{}, error)` — the refined value (same type as input), or an error.

---

#### `GenerateReport`

```go
func (e *Refinery) GenerateReport(filesScanned int) DryRunReport
```

Aggregates the in-session PII hit map (accumulated by `getOrSetSecureToken`) into a `DryRunReport`. Thread-safe.

**Parameters:**
- `filesScanned` — number of files or payloads processed (caller-supplied, for the `files_scanned` field).

---

#### `ResetHits`

```go
func (e *Refinery) ResetHits()
```

Clears the in-memory hit map. Call between requests in serve mode to prevent cross-contamination of reports. Thread-safe.

---

### 2.4 Cryptography Helpers

#### `Encrypt`

```go
func Encrypt(plaintext, key []byte) (string, error)
```

Encrypts `plaintext` with **AES-256-GCM**. Returns a hex-encoded string structured as:

```
[12-byte nonce][GCM ciphertext + 16-byte auth tag]
```

All encoded as a single continuous hex string. The nonce is randomly generated per call using `crypto/rand`.

---

#### `Decrypt`

```go
func Decrypt(hexCiphertext string, key []byte) ([]byte, error)
```

Decrypts a hex-encoded AES-256-GCM ciphertext produced by `Encrypt`. Returns the original plaintext bytes.

**Error cases:**
- Malformed hex → `hex.DecodeString` error.
- Ciphertext too short (< nonce size) → `"ciphertext too short"`.
- GCM authentication failure (wrong key or tampered ciphertext) → `cipher.Open` error.

---

#### `DecryptToken`

```go
func DecryptToken(v vault.Provider, masterKey []byte, token string) (string, error)
```

Looks up a OCULTAR token (e.g. `[EMAIL_af2101fb]`) in the vault and returns the original PII string.

**Behaviour:**
- If the `vault.Provider` implements `GetEncryptedByToken(token string) (string, bool)` (i.e. is a `*duckdbProvider`), performs a reverse lookup and decrypts.
- If the token is not found in the vault → returns the token unchanged (safe fallback — no data loss).
- If decryption fails (e.g. key rotation) → logs the error and returns the token unchanged (**fail-safe**, not fail-closed).

Used internally by the proxy re-hydration layer. Not intended for direct use.

---

## 3. Vault Package (`services/vault`)

### `Provider` interface

```go
type Provider interface {
    StoreToken(hash, token, encryptedPII string) (bool, error)
    GetToken(hash string) (string, bool)
    CountAll() int64
    Close() error
}
```

| Method | Description |
|---|---|
| `StoreToken(hash, token, encryptedPII string) (bool, error)` | Persists a PII mapping. Returns `(true, nil)` on new insertion, `(false, nil)` if the hash already exists (idempotent). |
| `GetToken(hash string) (string, bool)` | Looks up the token for a given SHA-256 PII hash. Returns `("", false)` on miss. |
| `CountAll() int64` | Returns the total number of vault entries. Used for pilot mode cap enforcement. |
| `Close() error` | Releases database connections. Always defer `Close()` after obtaining a provider. |

### `New` factory

```go
func New(cfg config.Settings, vaultPath string) (Provider, error)
```

Selects and constructs the correct backend:

| `cfg.VaultBackend` | Selected backend | License required |
|---|---|---|
| `""` or `"duckdb"` | `duckdbProvider` — embedded local file | Community |
| `"postgres"` | `postgresProvider` — external PostgreSQL cluster | Enterprise |

**Parameters:**
- `cfg` — loaded `config.Settings` (from `config.Global`).
- `vaultPath` — file path for DuckDB (ignored for postgres). Pass `":memory:"` for ephemeral in-process vaulting.

---

## 4. Config Package (`services/refinery/pkg/config`)

### `Settings` struct

```go
type Settings struct {
    Regexes            []RegexRule `yaml:"regexes"`
    Dictionaries       []DictRule  `yaml:"dictionaries"`
    PresidioConfidence float64     `yaml:"presidio_confidence"`
    VaultBackend       string      `yaml:"vault_backend"`
    PostgresDSN        string      `yaml:"postgres_dsn"`
}
```

### `RegexRule`

```go
type RegexRule struct {
    Type     string         `yaml:"type"`
    Pattern  string         `yaml:"pattern"`
    Compiled *regexp.Regexp // set automatically by CompileRegexes()
}
```

| Field | Description |
|---|---|
| `Type` | Token label prefix in `UPPER_SNAKE_CASE` (e.g. `EMAIL`, `INTERNAL_ID`). |
| `Pattern` | Go `regexp/syntax` compatible regex. Prefix with `(?i)` for case-insensitive. Does **not** support look-ahead/look-behind. |

### `DictRule`

```go
type DictRule struct {
    Type  string   `yaml:"type"`
    Terms []string `yaml:"terms"`
}
```

| Field | Description |
|---|---|
| `Type` | Token label prefix. |
| `Terms` | Exact strings to redact. Matching is case-insensitive (handled by `applyReplacement`). |

### Functions

| Function | Description |
|---|---|
| `config.Load()` | Initialises `config.Global` with defaults + loads `configs/protected_entities.json` + compiles regexes. Called once at startup. **Fatal** if `protected_entities.json` is missing or empty (fail-closed). |
| `config.InitDefaults()` | Alias for `Load()`, used in tests. |
| `config.CompileRegexes()` | Compiles all `RegexRule.Pattern` fields. Called automatically by `Load()`. |

### Built-in Regex Rules (Community defaults)

| Type | Pattern | Matches |
|---|---|---|
| `EMAIL` | `(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}` | Standard email addresses |
| `URL` | `(?i)https?://[^\s"<>{}\[\]\\]+\|www\.…` | HTTP/S URLs and bare `www.` domains |
| `SSN` | `\b\d{3}-\d{2}-\d{4}\b` | US Social Security Numbers |
| `CREDENTIAL` | `(?i)\bpassword\s*[:=]\s*[^\s,]+` | In-line passwords (e.g., "password: mypass") |
| `SECRET` | `(?i)\b(?:secret\|key\|token)\s*[:=]\s*[^\s,]+` | In-line secrets/keys |

---

## 5. Proxy Package (`services/refinery/pkg/proxy`)

### `Handler`

```go
type Handler struct { /* unexported fields */ }
```

Implements `http.Handler`. Constructed via `NewHandler`.

### `NewHandler`

```go
func NewHandler(eng *refinery.Refinery, v vault.Provider, masterKey []byte, targetURL string) (*Handler, error)
```

**Parameters:**
- `eng` — shared refinery instance.
- `v` — vault provider for re-hydration lookups.
- `masterKey` — decryption key.
- `targetURL` — default upstream URL (value of `OCU_PROXY_TARGET`).

**Returns:** `(*Handler, error)` — error if `targetURL` cannot be parsed.

**Internals:**
- Concurrency semaphore capped at **15** concurrent requests (matches PostgreSQL connection pool limit).
- `ResponseHeaderTimeout` set to **120 seconds** for upstream calls.
- Body reads capped at **5 MB** per request.

---

## 6. HTTP Endpoints

The dashboard/API server (`--serve <port>`) exposes the following endpoints.

### 6.1 `POST /api/refine`

Refines a raw text or JSON payload.

**Request:**

```http
POST /api/refine HTTP/1.1
Content-Type: application/json

{"messages": [{"role": "user", "content": "Email john@example.com for details."}]}
```

Or plain text:

```http
POST /api/refine HTTP/1.1
Content-Type: text/plain

Call Sarah at +33 6 12 34 56 78 or email sarah@example.com
```

**Response `200 OK`:**

```json
{
  "refined": "{\"messages\":[{\"role\":\"user\",\"content\":\"Email [EMAIL_9c8f7a1b] for details.\"}]}",
  "report": {
    "mode": "serve",
    "files_scanned": 1,
    "pii_hits": {
      "EMAIL": ["9c8f7a1b…"]
    },
    "total_pii_count": 1,
    "blocking": true
  }
}
```

**Size limit:** 5 MB. Payloads exceeding this are rejected with `413 Request Entity Too Large`.

---

### 6.2 `POST /api/refine/file`

Processes a file upload with streaming output. Designed for large datasets — never loads the entire file into RAM.

**Request:**

```http
POST /api/refine/file HTTP/1.1
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="data.csv"
Content-Type: text/csv

email,name
john@example.com,John Smith
...
```

**Response `200 OK`:** `application/octet-stream` — the refined file, streamed line-by-line.

```bash
# Usage example
curl -F "file=@my_data.csv" http://localhost:9090/api/refine/file > cleaned.csv
```

---

### 6.3 `GET /api/config`

Returns the full current configuration (`config.yaml` state).

### 6.4 `GET /api/config/regex`

Returns the list of active Regex rules, including their `canonical_mapping` (Google InfoType).

### 6.5 `POST /api/config/regex`

Adds or updates a Regex rule. Persists to `config.yaml`.
**Body**: `{"type": "CUSTOM_ID", "pattern": "\\b[0-9]{5}\\b"}`

### 6.6 `DELETE /api/config/regex`

Removes a Regex rule.
**Body**: `{"type": "CUSTOM_ID"}`

### 6.7 `GET /api/config/dictionary`

Returns all dictionary categories and their terms.

### 6.8 `POST /api/config/dictionary`

Adds a term to a dictionary category.
**Body**: `{"type": "VIP", "term": "Internal Project Name"}`

### 6.9 `DELETE /api/config/dictionary`

Removes a category (and all its terms).
**Body**: `{"type": "VIP"}`

### 6.10 `GET /api/config/mapping`

Returns the **Canonical Entity Registry**: a complete mapping of all OCULTAR internal identifiers to Google Cloud DLP InfoTypes.

### 6.11 `GET /api/config/system`

Returns system-level limits (`max_concurrency`, `queue_size`).

---

## 7. HTTP Proxy Mode

The proxy (`docker-compose.proxy.yml`) exposes port `8080` by default.

### Request Headers

| Header | Direction | Description |
|---|---|---|
| `Ocultar-Target` | Client → Proxy | Per-request upstream URL override. Validated against SSRF (blocks `localhost`, `127.x`, `10.x`, `192.168.x`, `169.254.x`). |
| `X-Forwarded-For` | Client → Proxy | Used as the `actor` identifier in audit logs. Falls back to `RemoteAddr`. |

### Response Headers Added by Proxy

| Header | Direction | Description |
|---|---|---|
| `X-Ocultar-Redacted` | Proxy → Upstream | Set to `"true"` when at least one PII entity was redacted from the request body. |

### Proxy Request Flow

```
Client POST ──► [5 MB cap] ──► obfuscation check
    ──► refinery.ProcessInterface (Tier 0→2 redaction)
        → on error → 403 / 500 (fail-closed, never forwards un-redacted data)
    ──► forward sanitised body to upstream
    ──► read upstream response
    ──► scan for [TYPE_token] patterns → DecryptToken → rehydrate
    ──► return final response to client
```

---

## 8. Error Reference

### Refinery Errors

| Error string | Cause | Resolution |
|---|---|---|
| `"encryption failed: …"` | `crypto/rand` or AES failure | System-level; check OS entropy source. |
| `"vault storage failed: …"` | DuckDB/PostgreSQL write error | Check disk space, DB connectivity, and `OCU_VAULT_PATH`. |
| `"SLM inference failed: …"` | Local SLM HTTP timeout or bad response | Check SLM container health: `docker compose logs ocultar-ai`. |
| `"trial limit reached"` | `OCU_PILOT_MODE=true` and vault cap exceeded | Upgrade to Enterprise or reset the vault. |

### Proxy HTTP Status Codes

| Code | Meaning |
|---|---|
| `200 OK` | Request processed and forwarded successfully. |
| `400 Bad Request` | Body could not be parsed (returned by upstream, passed through). |
| `403 Forbidden` | Refinery blocked the request (trial limit, SSRF attempt, obfuscated payload). |
| `413 Request Entity Too Large` | Payload exceeds 5 MB. |
| `429 Too Many Requests` | Concurrency semaphore full — retry after a short delay. |
| `500 Internal Server Error` | Refinery error during redaction. Un-redacted data was **not** forwarded. |
| `502 Bad Gateway` | Upstream returned an error or connection failed. |

### Config / Startup Fatal Errors

| Message | Cause |
|---|---|
| `[FATAL] Failed reading protected_entities.json!` | `configs/protected_entities.json` not found at startup. |
| `[FATAL] Failed parsing protected_entities.json!` | File contains malformed JSON. |
| `[FATAL] protected_entities.json parsed successfully but contains zero entries.` | File is a valid but empty JSON array `[]`. |
| `[vault] Active license tier … does not permit postgres …` | `vault_backend: postgres` without an Enterprise license. |
| `[vault] vault_backend is 'postgres' but postgres_dsn is not set` | Missing `postgres_dsn` in `config.yaml`. |
## 9. Sombra Gateway Configuration

The Sombra Gateway (`github.com/Edu963/sombra`) adds an extra layer of policy enforcement on top of the OCULTAR refinery.

### Connector Policy

Connectors in `sombra.yaml` support a `policy` block for fine-grained control:

```yaml
connectors:
  - name: slack-prod
    type: slack
    policy:
      strip_categories: ["SSN", "CREDENTIAL", "SECRET"]
      allowed_models: ["gemini-flash-latest"]
```

| Field | Description |
|---|---|
| `strip_categories` | List of PII types (e.g., `SSN`, `PERSON`) that should be **removed** (stripped) from the text before sending to the AI. This is stronger than redaction (which preserves relational tokens). |
| `allowed_models` | List of LLM models this connector is permitted to call. |

### Redaction vs. Stripping

- **Redaction**: Replaces PII with a token like `[PERSON_1234abcd]`. Useful for tasks where the AI needs to know *that* a person was mentioned without knowing *who*.
- **Stripping**: Replaces tokens with a generic placeholder like `[STRIPPED_SSN]`. Used for high-sensitivity data where the AI should not even see the pseudonymized token.
