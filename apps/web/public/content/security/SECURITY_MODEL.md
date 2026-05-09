# OCULTAR Security Model

**Document classification:** Public — for vendor security review  
**Version:** 1.2  
**Last updated:** 2026-04-28  
**Contact:** security@ocultar.dev

---

## 1. Executive Summary

OCULTAR is a zero-egress PII redaction proxy: raw personally identifiable information never leaves the deployment boundary because it is tokenized in-place before any upstream call is made. Every failure mode — AI scanner unavailability, vault write errors, queue saturation — is fail-closed, meaning the request is blocked and no PII is forwarded rather than degrading gracefully with data exposure. Sensitive data at rest is protected with AES-256-GCM encryption under a per-deployment key derived via HKDF-SHA256, and each vault write operation is covered by a SHA-256 hash-chained, Ed25519-signed audit log entry fsynced to disk.

---

## 2. Zero-Egress Guarantee

### What it means technically

"Zero-egress" means no raw PII value is transmitted to the upstream API under any operating condition. The redaction step in `services/refinery/pkg/proxy/proxy.go` is synchronous and blocking: the proxy constructs the upstream request only after `redactBody()` returns a sanitised payload with all PII replaced by opaque tokens (e.g. `[EMAIL_9c8f7a1b]`). If `redactBody()` returns an error for any reason, the request is rejected with HTTP 500 before the upstream connection is opened.

### What the guarantee covers

- All PII detected by Tiers 0–1.5 (regex, dictionary, entropy, phone, address, greeting) in the request body.
- All PII detected by the Tier 2 AI scanner (enterprise tier) when the SLM sidecar is active.
- JSON request bodies, streaming newline-delimited JSON, and plain-text bodies.

### What it explicitly does NOT cover

| Out of Scope | Reason |
|---|---|
| Request metadata (URL path, query parameters, HTTP headers) | Redaction applies to the request body only; PII in headers or URL parameters is not inspected. |
| Data in transit (client ↔ proxy, proxy ↔ upstream) | TLS is the caller's responsibility. OCULTAR does not terminate or upgrade TLS. |
| Upstream API provider security | Once tokenized data reaches the upstream, its handling is governed by the upstream provider's security posture. |
| Prompt injection attacks | OCULTAR detects and redacts PII; it does not analyse or block adversarial prompt content. |
| Re-hydration authorization | Token re-hydration (response-path replacement of tokens with plaintext) is available to any caller that receives the response. Callers requiring differential re-hydration must implement access control outside OCULTAR. |

### Architectural enforcement mechanism

The proxy pipeline is sequential and non-bypassable:

```
Client Request
    │
    ▼
TierMiddleware          ← enforces subscription tier, monthly caps
    │
    ▼
proxy.Handler.ServeHTTP
    │
    ├─ redactBody()     ← BLOCKING: no upstream call until this returns clean
    │       │
    │       ├─ Tier 0–1.5 detection (regex, dictionary, entropy, phone, address)
    │       └─ Tier 2 AI NER (enterprise, if scanner available)
    │
    ├─ resolveTarget()  ← SSRF check: rejects private/internal targets
    │
    └─ RoundTrip()      ← upstream call with sanitised body only
```

---

## 3. Fail-Closed Behavior

In every failure scenario tested, the proxy returns an error response and withholds both the upstream call and the `X-Ocultar-Redacted` response header. No partial or unredacted body is returned.

| Failure Scenario | HTTP Response | PII Exposure Risk | Test Reference |
|---|---|---|---|
| AI scanner (SLM sidecar) returns error | 500 Internal Server Error | None | `TestFailClosed_SLMUnreachable` |
| Vault write failure | 500 Internal Server Error | None | `TestFailClosed_VaultWriteFailure` |
| Vault timeout (15 s artificial delay, 3 s client timeout) | Client deadline exceeded — no response body | None | `TestFailClosed_SlowVault` |
| Proxy wait queue saturated (concurrency limit) | 429 Too Many Requests | None | `TestFailClosed_QueueFull` |
| Protected-entity dictionary empty at boot | `log.Fatalf` — process does not start | N/A | `TestFailClosed_ProtectedEntitiesBootGuard` |

**Test location:** `services/refinery/pkg/refinery/failclosed_test.go`

**Assertion method:** Each test calls `assertNoPII(t, body)` which scans the response body for three literal PII values: `"Jean-Pierre"`, `"Dumont"`, and `"jean.pierre@societe.fr"`. Any match is a test failure.

---

## 4. Encryption Model

### Algorithm

AES-256-GCM with a random 12-byte nonce prepended to the ciphertext. The authenticated encryption tag provides both confidentiality and integrity. A tampered ciphertext will be rejected at decrypt time with a GCM authentication failure.

### Key derivation

```
OCU_MASTER_KEY  (operator-supplied, high-entropy string)
OCU_SALT        (per-deployment, operator-supplied)
    │
    ▼
HKDF-SHA256(ikm=OCU_MASTER_KEY, salt=OCU_SALT, info="ocultar-aes-key")
    │
    ▼
32-byte AES-256 derived key
```

`OCU_MASTER_KEY` and `OCU_SALT` must differ between deployments. Their combination via HKDF ensures that vault databases from one deployment cannot be decrypted by another deployment's key material.

**Source:** `apps/proxy/main.go`, `getMasterKey()`.

### Vault storage format

Each vault entry stores:
- `token_id`: first 8 hex characters of `SHA-256(plaintext_PII)` — deterministic, no secrets
- `token_string`: `[TYPE_<token_id>]` — used for replacement in the payload
- `ciphertext`: `hex(nonce || AES-256-GCM(plaintext_PII, derived_key))` — plaintext never stored

The vault database contains no plaintext PII. An attacker with read access to the vault file cannot recover PII without the master key.

### Determinism property

Because `token_id = SHA-256(plaintext)[:8]`, the same PII value always produces the same token across requests and vault sessions. This is a deliberate design property: it enables cross-document relational integrity (e.g. the same person appears as `[PERSON_181bc039]` in all documents) without storing a lookup table of plaintext values.

**Implication:** token IDs are not secret. The confidentiality guarantee rests entirely on the ciphertext.

### Key management

Secrets (`OCU_MASTER_KEY`, `OCU_SALT`) are injected at runtime via Doppler and are never present in the codebase, Docker images, or container layers. The `docker-compose.community.yml` file references `${OCU_MASTER_KEY}` as an environment placeholder; the actual value is populated by `doppler run --`.

---

## 5. Audit Trail

### Audit logging by deployment context

Audit logging behavior differs by binary and tier. The `AuditLogger` interface (`services/refinery/pkg/refinery/refinery.go:27`) is the injection point; the implementation varies:

| Deployment | Logger | What is recorded | Cryptographic protection |
|---|---|---|---|
| Proxy — Community tier | `NoopAuditLogger` | Nothing — silent no-op | None |
| Proxy — Enterprise tier, `OCU_AUDIT_PRIVATE_KEY` set | `ImmutableLogger` (via `auditAdapter`) | Full event set — see below | SHA-256 hash chain + Ed25519 signatures + fsync |
| Proxy — Enterprise tier, `OCU_AUDIT_PRIVATE_KEY` unset | `NoopAuditLogger` | Nothing — warning logged at startup | None |
| Refinery CLI / HTTP server | `BasicFileLogger` | Timestamp, user, action, result, compliance mapping | None — plain JSON, append-only |

**Activation:** Set `OCU_AUDIT_PRIVATE_KEY` (hex-encoded 32-byte Ed25519 seed) in Doppler and restart. Generate with `openssl rand -hex 32`. The same seed must be retained to verify historical signatures after restarts — loss of the seed does not affect the hash chain, but makes signature verification of prior entries impossible.

### ImmutableLogger — specification (active in enterprise deployments)

The `ImmutableLogger` in `services/refinery/pkg/audit/immutable.go` is fully implemented with the following properties, verifiable in code:

**Per-event fields:**

| Field | Value |
|---|---|
| `timestamp` | RFC 3339 with nanoseconds, UTC |
| `actor` | Source IP / `X-Forwarded-For` value from the request |
| `action` | Verb (e.g. `REDACT`, `RESOLVE`, `REJECT`) |
| `resource` | Target endpoint or document identifier |
| `status` | `ALLOW` / `BLOCK` / `ERROR` |
| `details` | Contextual detail string (optional) |
| `prev_hash` | SHA-256 of the previous event's canonical payload |
| `signature` | Ed25519 signature over the current event payload |

**Hash chain:** Each event's `prev_hash` field contains `SHA-256(timestamp|actor|action|resource|status|details|prev_hash)` of the preceding entry. Deletion or reordering of any entry breaks the chain and is detectable.

**Signatures:** Each event is signed with Ed25519 over the full canonical payload including `prev_hash`, binding event authenticity to chain position.

**Disk durability:** Every `Log()` call invokes `logFile.Sync()` (fsync) before returning. Entries survive process crashes.

**Output format:** Newline-delimited JSON (`application/x-ndjson`). Compatible with Filebeat, Fluent Bit, and any SIEM that supports structured log ingestion.

### Persistent signing key

The Ed25519 signing key is operator-supplied via `OCU_AUDIT_PRIVATE_KEY` (hex-encoded 32-byte seed, injected through Doppler). The same seed is used across restarts, so signatures on all entries in the log are verifiable with a single stable public key retrievable via `PublicKeyHex()`. Cross-session signature verification is fully supported.

**Key generation:** `openssl rand -hex 32`. Store the output in Doppler and never commit it to the codebase.

**Key loss:** If `OCU_AUDIT_PRIVATE_KEY` is lost, prior log entries cannot be signature-verified. The hash chain integrity guarantee is unaffected — it does not depend on the signing key. Operators should back up the key in a durable secrets manager alongside `OCU_MASTER_KEY`.

---

## 6. Network Security

### SSRF Protection

OCULTAR enforces server-side request forgery protection on the `Ocultar-Target` override header. When a caller provides this header, the target URL is resolved and every resulting IP address is validated before the upstream connection is opened.

| Protection | Mechanism | Test Reference |
|---|---|---|
| RFC 1918 IPv4 blocking (10.0.0.0/8) | Explicit range check in `isPrivateIP()` | `TestSSRF_BlockedAddresses/rfc1918_10_lower`, `/rfc1918_10_upper` |
| RFC 1918 IPv4 blocking (172.16.0.0/12) | Explicit range check in `isPrivateIP()` | `TestSSRF_BlockedAddresses/rfc1918_172_lower`, `/rfc1918_172_upper` |
| RFC 1918 IPv4 blocking (192.168.0.0/16) | Explicit range check in `isPrivateIP()` | `TestSSRF_BlockedAddresses/rfc1918_192_lower`, `/rfc1918_192_upper` |
| Loopback blocking (127.0.0.0/8) | `ip.IsLoopback()` | `TestSSRF_BlockedAddresses/loopback_127_0_0_1`, `/loopback_127_upper` |
| DNS rebinding (localhost → 127.0.0.1) | Post-resolution IP validation | `TestSSRF_BlockedAddresses/dns_rebind_localhost` |
| AWS IMDS endpoint (169.254.169.254) | `ip.IsLinkLocalUnicast()` covers 169.254.0.0/16 | `TestSSRF_BlockedAddresses/link_local_imds` |
| Link-local range (169.254.0.0/16) | `ip.IsLinkLocalUnicast()` | `TestSSRF_BlockedAddresses/link_local_lower` |
| IPv6 loopback (::1) | `ip.IsLoopback()` after bracket-stripping | `TestSSRF_BlockedAddresses/ipv6_loopback` |
| IPv4-mapped IPv6 loopback (::ffff:127.0.0.1) | `ip.IsLoopback()` via Go's `net.IP` normalisation | `TestSSRF_BlockedAddresses/ipv4_mapped_ipv6_loopback` |
| Unspecified address (0.0.0.0, ::) | `ip.IsUnspecified()` — see note below | `TestSSRF_BlockedAddresses/unspecified_ipv4` |

**Test location:** `apps/proxy/ssrf_test.go`

### Vulnerabilities found and fixed during adversarial testing

The following two defects were discovered during the development of the SSRF test suite and fixed before any release. They are documented here because transparency about discovered defects is a more reliable trust signal than an absence of documented findings.

**Defect 1 — IPv6 addresses bypassed SSRF check (fixed)**

The original implementation used `net.SplitHostPort(parsed.Host)` to extract the hostname from a URL. For IPv6 addresses supplied as `http://[::1]/`, `parsed.Host` is `[::1]` (with brackets). `net.SplitHostPort` requires the format `[::1]:port` and fails without a port, returning the full bracketed string. `net.ParseIP("[::1]")` returns `nil` because it does not accept brackets. The result was that IPv6 loopback addresses produced a "failed to resolve" error path rather than a "SSRF blocked" path — technically still blocking the request, but not for the intended reason and potentially bypassable with valid-looking IPv6 addresses.

**Fix:** Changed to `host := parsed.Hostname()`, which is documented to strip brackets. `net.ParseIP("::1")` then correctly identifies the loopback address and `isPrivateIP` blocks it.

**Defect 2 — 0.0.0.0 not explicitly blocked (fixed)**

The unspecified address `0.0.0.0` (and its IPv6 equivalent `::`) is not a loopback, link-local, or RFC 1918 address. On most Linux kernels, a TCP connection to `0.0.0.0` is equivalent to a connection to localhost. The original `isPrivateIP` function did not check for unspecified addresses.

**Fix:** Added `ip.IsUnspecified()` as the first check in `isPrivateIP`.

---

## 7. Key Management

### Master key handling

`OCU_MASTER_KEY` is read from the process environment at startup. It is passed through HKDF-SHA256 (described in Section 4) and is never stored, logged, or written to disk. The derived key exists only in process memory.

In development mode (`--dev` flag), a hardcoded insecure default is used with an explicit log warning. Production deployments must supply `OCU_MASTER_KEY`; the proxy calls `log.Fatalf` and refuses to start without it when not in dev mode.

### Key rotation

No automated rotation procedure is currently implemented. Rotation requires:

1. Generating new `OCU_MASTER_KEY` and `OCU_SALT` values.
2. Re-encrypting all vault entries with the new derived key.
3. Updating the environment secrets in Doppler and restarting the deployment.

A helper script for step 2 is planned but not yet available. Operators performing rotation must handle vault re-encryption manually.

### Consequence of key loss

**Key loss is unrecoverable.** There is no key escrow, no backup KMS copy, and no recovery path. If `OCU_MASTER_KEY` is lost, all ciphertexts in the vault are permanently unreadable. The tokens stored in processed documents will remain valid as redaction placeholders but cannot be re-hydrated to plaintext.

Operators are strongly advised to store `OCU_MASTER_KEY` in a durable secrets manager with backup (e.g. Doppler with export, AWS Secrets Manager with cross-region replication).

### Codebase hygiene

`OCU_MASTER_KEY`, `OCU_SALT`, and all other secrets are absent from the codebase. The `.env` file (local development only) is listed in `.gitignore`. The `docker-compose.community.yml` file references secrets as environment variable placeholders (`${OCU_MASTER_KEY}`) injected at runtime by `doppler run --`.

---

## 8. Vault Persistence

### Storage mechanism

The vault uses DuckDB, an embedded analytical database. In the default community configuration, the DuckDB file is located at the path specified by `OCU_VAULT_PATH` (default: `/data/vault.db`).

### Container persistence

The vault file is mounted via a named Docker volume (`vault-data:/data`) defined in `docker-compose.community.yml`. Named volumes survive container recreation, image upgrades, and `docker compose down`.

### Production defect found and fixed

An anonymous volume configuration was identified during development in which `vault.db` resided in the container's writable layer rather than in a named volume. Every execution of `docker compose up --force-recreate` silently destroyed the vault file and all stored tokens — with no error message. This defect was fixed by adding the `vault-data` named volume to `docker-compose.community.yml` before any production use.

**Operators who deployed using the earlier anonymous-volume configuration should verify their current deployment uses the named volume** by running `docker volume ls` and confirming `ocultar-proxy-net_vault-data` (or equivalent) is present.

### Persistence verification

Three integration tests in `services/vault/persistence_test.go` verify vault correctness properties:

| Test | Property verified |
|---|---|
| `TestPersistence_TokenSurvivesRestart` | Tokens written in session A are fully recoverable after vault close + reopen |
| `TestPersistence_Determinism` | The same plaintext produces the same token in every session |
| `TestPersistence_KeyIsolation` | AES-256-GCM authentication rejects decryption under a wrong key; no corrupt plaintext is returned |

---

## 9. Compliance Mappings

The following mappings reflect OCULTAR's technical properties as of the current version. They are provided for reference during vendor assessment. They do not constitute a legal compliance determination.

| Regulation / Standard | Article / Requirement | OCULTAR's Position |
|---|---|---|
| GDPR | Article 25 — Data protection by design and by default | Zero-egress architecture ensures PII is tokenized before reaching any third-party processor. Vault stores only ciphertext. |
| GDPR | Article 32 — Security of processing | AES-256-GCM encryption at rest, HKDF key derivation, fail-closed request handling, Ed25519-signed audit log. |
| GDPR | Article 30 — Records of processing activities | Audit log records actor, action, resource, and timestamp for every redaction event. |
| EU AI Act | Data minimization requirement for AI training and inference | Tokenization prevents raw PII from entering AI inference pipelines. Upstream LLMs receive only token placeholders. |
| NIS2 Directive | Incident detection and logging | Hash-chained audit log with fsync durability supports post-incident forensics. |
| SOC 2 Type II | Trust Service Criteria | **In preparation.** Controls inventory in progress; formal audit not yet initiated. |
| ANSSI SecNumCloud | French national cloud security qualification | **Not yet.** On the product roadmap. No current ANSSI qualification claim is made. |
| PCI-DSS | Requirement 3 — Protect stored account data | Payment card numbers are detected (Luhn validation in Tier 1.1) and tokenized. The vault stores only AES-256-GCM ciphertext. PCI-DSS applicability depends on deployment scope and assessor determination. |

---

## 10. What OCULTAR Does Not Claim

The following limitations are explicit and intentional. They are stated here because a security document that omits known boundaries is less useful than one that names them directly.

**OCULTAR does not guarantee upstream AI provider security.**
Tokenized data is forwarded to the upstream API. The security posture of that API — including its model training practices, data retention policies, and employee access controls — is outside OCULTAR's scope and governed by the operator's agreement with the upstream provider.

**OCULTAR does not prevent prompt injection attacks.**
OCULTAR inspects content for PII patterns. It does not analyse whether a request contains instructions intended to manipulate an AI model's behavior. Prompt injection defense is a separate concern.

**OCULTAR does not encrypt data in transit.**
TLS between the client and the OCULTAR proxy, and between the proxy and the upstream, is the caller's and operator's responsibility. OCULTAR does not terminate, upgrade, or enforce TLS.

**OCULTAR does not provide access control beyond tier enforcement.**
API key validation and per-tier monthly rate limits are enforced at the `TierMiddleware` layer (Section 2). There is no per-user or per-document authorization model. All callers with a valid API key at a given tier can access the same proxy functionality.

**OCULTAR does not guarantee detection of all PII.**
The detection pipeline (Tiers 0–1.5 in community, Tier 2 in enterprise) operates on statistical and pattern-based methods. Novel PII formats, obfuscated values, or application-specific sensitive data not covered by the configured rules may not be detected. Operators are responsible for validating detection coverage against their specific data formats.

**The cryptographic audit trail is an enterprise feature gated on `OCU_AUDIT_PRIVATE_KEY`.**
Community tier uses `NoopAuditLogger` — no events are recorded. Enterprise deployments without `OCU_AUDIT_PRIVATE_KEY` also produce no audit log (a warning is emitted at startup). The `ImmutableLogger` (SHA-256 hash chain + Ed25519 signatures + fsync) activates only when `OCU_AUDIT_PRIVATE_KEY` is set. The Refinery CLI server always uses `BasicFileLogger` (plain JSON, no cryptographic protection). Operators requiring an audit trail must be on the enterprise tier with the key configured.

**The cryptographic audit trail requires operator key management.**
When `ImmutableLogger` is active, the Ed25519 signing key is loaded from `OCU_AUDIT_PRIVATE_KEY` at startup. Cross-session signature verification is fully supported as long as the same key is retained. Loss of `OCU_AUDIT_PRIVATE_KEY` makes prior signatures unverifiable; the hash chain integrity guarantee is unaffected.

---

*This document reflects the state of the codebase as of version 1.2.0. All claims are derived from the implementation in this repository. No claim in this document is aspirational; where a capability is planned rather than implemented, it is explicitly marked as such.*
