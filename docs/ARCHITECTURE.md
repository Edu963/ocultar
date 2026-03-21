# OCULTAR | Architecture Reference

> **Audience:** Solutions architects, security auditors, and developers who need a precise understanding of how OCULTAR processes data, where trust boundaries lie, and how components interact.

---

## Table of Contents

1. [System Context (C4 Level 1)](#1-system-context-c4-level-1)
2. [Container Diagram (C4 Level 2)](#2-container-diagram-c4-level-2)
3. [Refinery Pipeline — Data Flow](#3-refinery-pipeline--data-flow)
4. [AI-Driven Governance & Orchestration](#4-ai-driven-governance--orchestration)
5. [Security Trust Boundaries](#5-security-trust-boundaries)
6. [Package Dependency Graph](#6-package-dependency-graph)
7. [Cryptographic Design](#7-cryptographic-design)
8. [Vault Schema](#8-vault-schema)
9. [Fail-Closed Guarantees](#9-fail-closed-guarantees)
10. [Scalability & Concurrency Model](#10-scalability--concurrency-model)

---

## 1. System Context (C4 Level 1)

Who uses OCULTAR, and what external systems does it touch?

```mermaid
graph TD
    Operator["👤 Operator\n(SOC / DevOps)"]
    Client["👤 End-User / Application / Server"]
    OCULTAR["🔒 OCULTAR\n(Global Data Refinery)"]
    Upstream["☁️ Upstream API / SIEM\n(OpenAI, Splunk, etc.)"]
    Vault[("🗄️ Identity Vault\nDuckDB / PostgreSQL")]

    Client -- "HTTP / gRPC / Syslog" --> OCULTAR
    OCULTAR -- "Sanitised Data" --> Upstream
    OCULTAR -- "Re-hydrated response" --> Client
    OCULTAR -- "AES-256-GCM encrypted PII" --> Vault
    Operator -- "RBAC Controlled Access" --> OCULTAR

    style OCULTAR fill:#1a1a2e,color:#e2c27d,stroke:#e2c27d
    style Vault fill:#16213e,color:#e2c27d,stroke:#e2c27d
```

**Zero-Egress guarantee:** PII never flows to `Upstream API` — only tokens do. The vault lives entirely on-premise.

---

## 2. Container Diagram (C4 Level 2)

Internal services and how they communicate:

```mermaid
graph TD
    subgraph "Client Network"
        App["Application / curl"]
    end

    subgraph "OCULTAR Host (on-premise)"
        Proxy["ocultar-proxy\n(Go HTTP/gRPC/Syslog)\npkg/proxy"]
        Engine["OCULTAR Engine\n(Go library)\npkg/engine\nRefineBatch ✦"]
        Config["Config Loader\npkg/config\n configs/config.yaml"]
        Vault_svc["Identity Vault\npkg/vault"]
        DDB[("DuckDB\nvault.db")]
        PG[("PostgreSQL ✦\nexternal HA cluster")]
        SLM["Specialized SLM ✦\n(Clinical / General)\nllama.cpp / Qwen"]
        Dashboard["Dashboard\nindex.html\n:9090"]
        AuditLog["Audit System\npkg/audit\nTamper-proof logs"]
        Connectors["Sombra Pro Connectors ✦\npkg/connector\nSlack / SharePoint"]
    end

    subgraph "External"
        Upstream["Upstream LLM API\nhttps://api.openai.com"]
    end

    App -->|"POST /v1/...\nraw JSON"| Proxy
    Proxy -->|"ProcessInterface"| Engine
    Engine -->|"reads rules"| Config
    Engine -->|"StoreToken / GetToken"| Vault_svc
    Engine -->|"ScanForPII ✦"| SLM
    Engine -->|"Log ✦"| AuditLog
    Connectors -->|"RefineBatch / ProcessInterface"| Engine
    Vault_svc -->|"default"| DDB
    Vault_svc -->|"vault_backend: postgres ✦"| PG
    Proxy -->|"sanitised payload"| Upstream
    Upstream -->|"response with tokens"| Proxy
    Proxy -->|"DecryptToken → rehydrate"| Vault_svc
    Dashboard -->|"POST /api/refine"| Engine

    style SLM fill:#f96,stroke:#c33
    style PG fill:#f96,stroke:#c33
    style AuditLog fill:#f96,stroke:#c33
```
*✦ = Enterprise only*

---

## 3. Refinery Pipeline — Data Flow

A single request through the OCULTAR proxy, showing every processing step:

```mermaid
sequenceDiagram
    participant C as Client
    participant P as Proxy Handler
    participant E as Engine
    participant V as Vault
    participant U as Upstream API

    C->>P: POST /v1/chat/completions<br/>{messages:[{content:"Email john@example.com"}]}

    Note over P: 1. Enforce 5 MB cap
    P->>P: Obfuscation check<br/>(Base64/JWT prefix, URL-encoded JSON)

    P->>E: ProcessInterface(body, actor)

    Note over E: Tier 0.1 — Base64 Evasion Shield
    Note over E: Tier 0.5 — Dictionary Shield (protected_entities.json)
    Note over E: Tier 1 — Regex Shields (EMAIL, URL, custom)
    Note over E: Tier 1.1 — Phone Shield (libphonenumber)
    Note over E: Tier 1.2 — Address Shield (heuristic parser)
    Note over E: Tier 1.5 — Greeting & Signature Shield
    Note over E: Tier 2 — AI NER Scan ✦ (SLM / Clinical NER)
    Note over E: Tier 3 — Structural Heuristics (Zero-Egress Hardening)

    E->>V: GetToken(sha256("john@example.com"))
    V-->>E: cache miss
    E->>E: HKDF(keyMaterial, salt) → derivedKey
    E->>E: Encrypt("john@example.com", derivedKey)
    E->>V: StoreToken(hash, "[EMAIL_9c8f7a1b]", ciphertext)
    V-->>E: (true, nil) — inserted

    E-->>P: {messages:[{content:"Email [EMAIL_9c8f7a1b]"}]}

    P->>P: Set X-Ocultar-Redacted: true
    P->>U: POST /v1/chat/completions<br/>{messages:[{content:"Email [EMAIL_9c8f7a1b]"}]}
    U-->>P: {choices:[{message:{content:"I'll email [EMAIL_9c8f7a1b]"}}]}

    Note over P: Response re-hydration (RBAC Checked)
    P->>V: GetEncryptedByToken("[EMAIL_9c8f7a1b]")
    V-->>P: ciphertext
    P->>P: Decrypt(ciphertext, masterKey) → "john@example.com"
    P->>P: Log to Tamper-proof Audit Trail

    P-->>C: {choices:[{message:{content:"I'll email john@example.com"}}]}
```

---

## 4. AI-Driven Governance & Orchestration

Ocultar uses a decentralized network of **Specialized Agent Skills** to maintain security, compliance, and product integrity. These are orchestrated by the `Continuous AI Orchestrator`.

### Governance Tiers

| Tier | Agent Skills | Focus |
|---|---|---|
| **Security & Privacy** | `pii-detection-updater`, `dictionary-shield-manager`, `security-sanitizer`, `audit-log-validator` | Heart of the system; prevents data egress and ensures log integrity. |
| **Development & Code** | `refinery-rule-generator`, `sombra-gateway-policy-enforcer`, `enterprise-dashboard-checker` | Automates architectural compliance and UI consistency. |
| **Intelligence & Audit** | `change-impact-visualizer`, `pii-security-docs-generator`, `red-team-evasion-scanner` | Provides transparency and proactive stress-testing. |
| **Business & Ops** | `roi-cost-efficiency-accountant`, `industry-snapshot-generator`, `sombra-performance-benchmarker` | Optimizes for scale, cost, and industry-specific needs. |

### The 16-Step Ocultar Protocol
The orchestrator triggers a standard 16-step sequence for every meaningful change:
1. **Docs** → 2. **Security Docs** → 3. **Packaging** → 4. **PII Verification** → 5. **Rule Gen** → 6. **Shield Sync** → 7. **Log Validation** → 8. **Policy Enforcement** → 9. **Dashboard Check** → 10. **Sanitization** → 11. **Release Builder** → 12. **Impact Visualization** → 13. **Red-Teaming** → 14. **Industry Snapshot** → 15. **ROI Accounting** → 16. **Performance Benchmarking**.

---

## 5. Security Trust Boundaries

```mermaid
graph LR
    subgraph "UNTRUSTED ZONE\n(client network, upstream APIs)"
        Client["Client\nApplication"]
        Upstream["Upstream\nLLM API"]
    end

    subgraph "TRUSTED ZONE\n(OCULTAR process boundary)"
        direction TB
        Proxy["proxy.Handler"]
        Engine["engine.Engine"]
        Vault["vault.Provider"]
        Key["MasterKey\n(in-process memory only)"]
    end

    subgraph "PERSISTENT STORAGE\n(on-premise, encrypted at rest)"
        DB[("vault.db / PostgreSQL\nAES-256-GCM ciphertext only")]
    end

    Client -- "raw PII crosses boundary →" --> Proxy
    Proxy -- "tokens only cross boundary →" --> Upstream
    Engine -- "writes encrypted PII" --> DB
    Key -- "never leaves process" --> Engine

    style Key fill:#c33,color:#fff,stroke:#900
```

**What NEVER leaves the trusted zone:**
- Plain-text PII
- The `OCU_MASTER_KEY` value
- Decrypted vault contents

**What MAY leave the trusted zone:**
- Token strings (e.g. `[EMAIL_9c8f7a1b]`) — these are safe; reversible only with the vault + master key.
- Structured audit log entries (token strings only, never PII).

---

## 5. Package Dependency Graph

```mermaid
graph TD
    Config["pkg/config\n(zero internal deps)"]
    License["pkg/license"]
    Vault["pkg/vault"]
    Engine["pkg/engine"]
    Proxy["pkg/proxy"]
    Reporter["pkg/reporter"]
    DistCommunity["dist/community\n(main)"]
    DistEnterprise["dist/enterprise\n(main)"]
    Connectors["pkg/connector"]

    Vault --> Config
    Vault --> License
    Engine --> Config
    Engine --> Vault
    Engine --> License
    Proxy --> Engine
    Proxy --> Vault
    Reporter --> Engine
    DistCommunity --> Engine
    DistCommunity --> Vault
    DistCommunity --> Config
    DistEnterprise --> Engine
    DistEnterprise --> Vault
    DistEnterprise --> Config
    DistEnterprise --> Reporter
    Connectors --> Engine
    Connectors --> Config
```

**Key architecture rules:**
1. `pkg/config` has zero internal dependencies — it is the root of the dependency tree.
2. `pkg/engine` does **not** import `pkg/proxy` — the engine knows nothing about HTTP.
3. `pkg/vault` does **not** import `pkg/engine` — storage is decoupled from redaction logic.

---

## 6. Cryptographic Design

### Key Derivation

```
OCU_MASTER_KEY (env var, arbitrary string)
        │
        ▼
   HKDF-SHA256(keyMaterial, salt, info)
        │
        ▼
   32-byte AES key  ──► used for Encrypt() / Decrypt()
```

The Sombra gateway uses **HKDF** (SHA-256, fixed salt) instead of raw SHA-256 for stronger key separation.

### License Enforcement (Entitlement Bitmask)
Enterprise licenses (signed via Ed25519) include a 64-bit `Capabilities` mask. This provides granular control over which "Pro" features are enabled for a specific customer.
- **Bit 0 (1)**: Slack Connector
- **Bit 1 (2)**: SharePoint Connector
- **Bitmask 0**: Grants all (legacy/default compatibility)

### Token Generation

```
original PII: "john@example.com"
        │
        ▼
  SHA-256("john@example.com")
  = 9c8f7a1b3f2e4d6a...  (full 64-char hex)
        │
        ├──► Token: [EMAIL_9c8f7a1b]   ← first 8 hex chars of the hash
        │           (stored as plain text — safe to expose)
        │
        └──► Hash (full 64 chars): lookup key in vault
```

**Determinism:** The same PII always produces the same token. This is intentional — it preserves relational integrity across records (two rows with the same email get the same token).

### AES-256-GCM Ciphertext Format

```
hex_encoded(
    random_nonce (12 bytes, crypto/rand)
    ||
    gcm.Seal(nonce, nonce, plaintext, nil)  ← 16-byte auth tag appended by GCM
)
```

Total overhead per PII value: 12 bytes (nonce) + 16 bytes (auth tag) = 28 bytes of overhead, hex-encoded to 56 extra characters on top of the plaintext length.

### Storage Layout

| Column | Type | Content |
|---|---|---|
| `pii_hash` | TEXT (PK) | Full SHA-256 hex of original PII — lookup key |
| `token` | TEXT | `[TYPE_XXXXXXXX]` string — safe to expose |
| `encrypted_pii` | TEXT | Hex-encoded AES-256-GCM ciphertext |

---

## 7. Vault Schema

### DuckDB (default)

```sql
CREATE TABLE IF NOT EXISTS vault (
    pii_hash      TEXT PRIMARY KEY,
    token         TEXT NOT NULL,
    encrypted_pii TEXT NOT NULL
);
```

Single file at `vault.db`. Zero external dependencies. Supports concurrent reads; writes are serialised via DuckDB's MVCC.

### PostgreSQL (Enterprise ✦)

Same schema created at startup via `CREATE TABLE IF NOT EXISTS`. Connection string via `postgres_dsn` in `config.yaml`. Connection pool capped at 15 (`sync.Semaphore` in `pkg/proxy`) to prevent exhaustion.

---

## 8. Fail-Closed Guarantees

OCULTAR enforces fail-closed at every critical junction:

| Failure scenario | OCULTAR behaviour |
|---|---|
| `protected_entities.json` missing or empty | `log.Fatal` — process refuses to start |
| `configs/config.yaml` contains invalid regex | `regexp.MustCompile` panics at startup — process refuses to start |
| Engine error during redaction (any tier) | Proxy returns `500` — un-redacted body is **never forwarded** |
| Trial limit reached (`OCU_PILOT_MODE`) | Proxy returns `403` — request is blocked |
| Obfuscated payload detected (Base64/JWT prefix, URL-encoded JSON) | Engine returns error → proxy returns `403` |
| SSRF attempt in `Ocultar-Target` header | `resolveTarget` returns error → proxy returns `403` |
| Payload exceeds 5 MB | `MaxBytesReader` triggers → proxy returns `413` |
| Concurrency limit exceeded (>15 concurrent) | Semaphore blocks for 5 seconds → `429` if not acquired |
| Token re-hydration key mismatch (key rotation) | Logs error, returns **token unchanged** (fail-safe — no data loss) |
| SLM inference fails (Enterprise) | Engine returns error → proxy returns `500` |
| Pro Connector initialized without Bitmask license | `log.Printf([WARN])` — connector remains dormant (Fail-Closed) |

> **Re-hydration exception:** Unlike request-side processing which is strictly fail-closed, response-side re-hydration is **fail-safe** — if a token cannot be decrypted (e.g. after a key rotation), the token is returned as-is rather than returning a 500 error. This prevents permanent proxy unavailability while preserving security (the token string itself contains no PII).

---

## 9. Scalability & Concurrency Model

### Concurrency limits

| Layer | Mechanism | Limit |
|---|---|---|
| Proxy (per instance) | `chan struct{}` semaphore | 15 concurrent requests |
| DuckDB | File-level lock | Single-process only |
| PostgreSQL (Enterprise) | Connection pool (15 max) | Horizontally scalable |

### Horizontal scaling

| Mode | Scalable? | Notes |
|---|---|---|
| CLI / Dashboard | ✅ (per-process) | Each process uses its own `vault.db` — sharding by node. |
| Proxy (DuckDB) | ❌ | Single-node only; multiple proxy instances would use separate vaults. |
| Proxy (PostgreSQL ✦) | ✅ | Multiple proxy instances can share a single PostgreSQL vault. Load-balance freely. |

### Memory model

- **Request bodies** are read fully into RAM before processing (5 MB cap enforced via `io.MaxBytesReader`).
- **File uploads** (`/api/refine/file`) use `bufio.Scanner` — line-by-line streaming. Memory overhead is bounded to a single line, not the file size.
- **SLM batch scan** marshals an entire JSON record to a string once per document — for deeply nested objects, worst-case memory is 2× the JSON document size.
- **Vault hits map** (`engine.Hits`) is accumulated per-request and protected by `sync.Mutex`. It is reset between requests via `ResetHits()`.
