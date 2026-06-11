# Architecture

Ocultar is a local HTTP sidecar that intercepts text before it reaches a cloud AI model, masks every piece of personally identifiable information (PII) it finds, and returns the cleaned text to the caller. The originals are stored in an encrypted local vault. Authorized callers can restore them after the AI response is received.

The design follows one invariant: **if Ocultar cannot guarantee that PII has been removed, the request is blocked — it is never forwarded raw**.

---

## Detection tiers

Ocultar runs two detection passes in sequence before any text leaves the machine.

### Tier 1 — Deterministic (always active)

| Sub-tier | Name | What it catches |
|---|---|---|
| 0 | Dictionary shield | VIP names, org names, custom entity lists from `configs/protected_entities.json` |
| 0.1 | Evasion shield | Base64 and JWT blobs — decoded, scanned, and re-encoded |
| 0.5 | Entropy shield | High-entropy strings (API keys, secrets) via Shannon scoring |
| 1 | Rule engine | EMAIL, SSN, IBAN, credit cards, 50+ national ID formats (regex + checksum validation) |
| 1.1 | Phone shield | Phone numbers — all country formats via libphonenumber |
| 1.2 | Address shield | Street addresses — heuristic parser (EN/FR/ES/DE) |
| 1.5 | Contextual shield | Names in greetings, signatures, and interrogative sentences |

Tier 1 is pure Go, zero external dependencies, and runs in under 5 ms for typical prompt sizes.

### Tier 2 — Contextual AI (optional, configurable)

Sends text to a local Small Language Model (SLM) for named-entity recognition. The Tier 2 scanner is always initialized but produces no results unless a compatible sidecar is running.

Activate it by pointing `SLM_SIDECAR_URL` at a running NER sidecar:

```bash
SLM_SIDECAR_URL=http://localhost:8085 ./ocultar -serve 4141
```

Two adapter protocols are supported:
- `privacy-filter` (default) — compatible with [openai/privacy-filter](https://huggingface.co/openai/privacy-filter) models
- `openai-chat` — compatible with llama.cpp, Qwen, and any OpenAI-protocol endpoint

If the Tier 2 sidecar is unavailable or times out, the circuit breaker opens and Ocultar continues with Tier 1 only — it does not block the request.

---

## Vault

Every detected entity is replaced by a deterministic token (`[EMAIL_9c8f7a1b]`) and the original value is stored in a local DuckDB vault encrypted with AES-256-GCM.

**Key derivation:** The vault key is never stored on disk. At startup, `OCU_MASTER_KEY` and `OCU_SALT` are passed through HKDF-SHA256 to derive a per-deployment encryption key that lives only in memory.

**Token format:** `SHA-256(original_value)[:8]` as hex. The same input always produces the same token within a vault, which allows privacy-safe analytics (joins, counts) on tokenized data without de-tokenization.

**Reveal:** Authorized callers (bearer token = `OCU_AUDITOR_TOKEN`) can call `POST /api/reveal` to restore tokens to originals. Every reveal call is recorded in the audit log.

---

## Fail-closed design

```
Client request
     │
     ▼
POST /api/refine
     │
     ├─ Tier 1 scan (deterministic)
     │       │
     │       ▼
     ├─ Vault StoreToken() ──── fails? ──► 500 error, request blocked
     │       │
     │       ▼
     ├─ Tier 2 scan (if configured)
     │       │
     │       ▼
     └─ Return masked text to caller
```

If `StoreToken()` fails (vault offline, disk full, key missing), the engine returns a `500` error and stops. The upstream AI model is never called. There is no bypass mode.

---

## Audit log

Every security-relevant event is written to an NDJSON audit log with an Ed25519 signature over each entry. The signing key is stored in the OS keychain (never on disk alongside the log).

Events logged:
- Vault entry created (token + entity type, never the original value)
- `POST /api/reveal` call (actor, tokens requested, timestamp)
- NER engine tier switch (Tier 1 only ↔ Tier 1 + Tier 2)

---

## Module layout

```
services/refinery/   — PII detection engine and HTTP server
internal/pii/        — Entity registry, regex patterns, validators
services/vault/      — DuckDB vault: encrypt, store, retrieve
apps/sombra/         — Zero-egress gateway (Sombra)
apps/proxy/          — Lightweight HTTP proxy layer
pkg/gateway/         — Shared gateway interfaces
```

`services/refinery` is the entry point for the standalone binary (`./ocultar -serve 4141`).
`apps/sombra` is the gateway layer used when routing AI traffic through Ocultar as a proxy.

---

## Request flow (standalone mode)

```
Client → POST /api/refine → Refinery (Tier 1 + Tier 2) → Vault → masked text → Client
                                                                  ↑
                                              POST /api/reveal ───┘ (auditor only)
```

The client is responsible for substituting tokens back into the AI response after processing. The `POST /api/reveal` endpoint is available for this purpose.

---

## Security properties

| Property | Mechanism |
|---|---|
| PII never leaves the machine | Detection and tokenization are 100% local; no external API calls |
| Vault encrypted at rest | AES-256-GCM, key derived via HKDF-SHA256, never written to disk |
| Fail-closed on vault error | `StoreToken` failure → 500, upstream never called |
| Tamper-evident audit log | Ed25519-signed NDJSON entries, key in OS keychain |
| Consistent tokenization | `SHA-256(value)[:8]` — same input, same token |
| Evasion resistance | Base64/JWT decoded and scanned before masking |
