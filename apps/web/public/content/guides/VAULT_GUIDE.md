# Vault — Storage Layer Guide

> **Audience:** Developers integrating OCULTAR into a Go application, and operators deploying it in production.

The vault is OCULTAR's encrypted token store. Every time the refinery detects PII, it writes one row to the vault and replaces the PII in the payload with a deterministic token. This document explains how that works, when to care about it, and how to configure it.

---

## Table of Contents

1. [What the Vault Does](#1-what-the-vault-does)
2. [Token Format](#2-token-format)
3. [The Tokenization Flow](#3-the-tokenization-flow)
4. [Rehydration](#4-rehydration)
5. [Backends: DuckDB vs PostgreSQL](#5-backends-duckdb-vs-postgresql)
6. [Using the Vault in Go Code](#6-using-the-vault-in-go-code)
7. [In-Memory Vault for Tests](#7-in-memory-vault-for-tests)
8. [Security Properties](#8-security-properties)
9. [Entity Registry (Path 3)](#9-entity-registry-path-3)

---

## 1. What the Vault Does

The vault is a local encrypted lookup table. Its job is simple:

- **Store**: given a PII string, encrypt it and record the mapping `token_id → (token, ciphertext)`.
- **Retrieve**: given a `token_id`, return the original PII string (after AES decryption).

No PII ever leaves the vault unencrypted. The upstream API (OpenAI, Gemini, etc.) sees only tokens like `[EMAIL_9c8f7a1b]`.

---

## 2. Token Format

Every token looks like `[TYPE_xxxxxxxx]` where:

| Part | Example | Meaning |
|---|---|---|
| `TYPE` | `EMAIL`, `PERSON`, `SSN` | The detected PII category |
| `xxxxxxxx` | `9c8f7a1b` | First 8 hex chars of SHA-256(plaintext) |

Example: `alice@example.com` → `[EMAIL_9c8f7a1b]`

The hash suffix is **deterministic**: the same input always produces the same token. This means the vault is naturally idempotent — re-processing a document that has already been redacted produces identical tokens with no duplicate rows.

---

## 3. The Tokenization Flow

For each detected PII match, the refinery runs:

```
1. token_id  = hex(SHA-256(plaintext))[0:8]
2. token     = "[TYPE_" + token_id + "]"
3. ciphertext = AES-256-GCM(plaintext, HKDF(OCU_MASTER_KEY, OCU_SALT))
4. vault.StoreToken(token_id, token, ciphertext)
   → returns (inserted=true) if new, (inserted=false) if already cached
5. Replace PII in payload with token
```

`StoreToken` is idempotent: if `token_id` already exists, the existing row is returned and no write occurs. This makes the entire pipeline safe to retry.

---

## 4. Rehydration

The proxy re-hydration layer scans upstream responses for `[TYPE_xxxxxxxx]` patterns and replaces them with the original PII:

```
[EMAIL_9c8f7a1b] → vault.GetToken("9c8f7a1b") → ciphertext
                  → AES-256-GCM.Decrypt(ciphertext, derivedKey)
                  → "alice@example.com"
```

Rehydration is **optional and per-client**. The proxy only rehydrates when the caller is authorized. If rehydration is disabled, tokens are returned as-is.

---

## 5. Backends: DuckDB vs PostgreSQL

| | DuckDB (default) | PostgreSQL |
|---|---|---|
| Setup | Zero-config — a single `.db` file | Requires `postgres_dsn` in `config.yaml` |
| Concurrency | Single-writer (embedded) | Multi-process HA |
| Use case | Dev, single-node production | Multi-replica, cloud-native |
| Config key | `vault_backend: duckdb` (or empty) | `vault_backend: postgres` |

Set `OCU_VAULT_PATH` to control where the DuckDB file is written (default: `vault.db` next to the binary). For PostgreSQL, set `postgres_dsn` in `configs/config.yaml`.

---

## 6. Using the Vault in Go Code

```go
import (
    "github.com/ocultar-dev/ocultar/pkg/config"
    "github.com/ocultar-dev/ocultar/services/vault"
)

func main() {
    config.InitDefaults()

    v, err := vault.New(config.Global, "/data/vault.db")
    if err != nil {
        log.Fatal(err)
    }
    defer v.Close()

    // Store a token manually (the refinery does this automatically)
    inserted, err := v.StoreToken("9c8f7a1b", "[EMAIL_9c8f7a1b]", encryptedBytes)

    // Look up an existing token
    token, found := v.GetToken("9c8f7a1b")

    // Count total vault entries (used for pilot-mode cap)
    count := v.CountAll()
}
```

In practice you never call `StoreToken` directly — the refinery does it during `RefineString`. You do call `DecryptToken` in the re-hydration layer.

---

## 7. In-Memory Vault for Tests

Pass an empty string as `vaultPath` to get an ephemeral in-process DuckDB instance. It is destroyed when `Close()` is called. No `.db` files are written.

```go
func TestMyFeature(t *testing.T) {
    v, _ := vault.New(config.Settings{VaultBackend: "duckdb"}, "")
    t.Cleanup(func() { v.Close() })
    // ...
}
```

This is the required pattern for all tests in `services/refinery`. See `CLAUDE.md` → Testing Conventions.

---

## 8. Security Properties

| Property | Implementation |
|---|---|
| **Confidentiality** | AES-256-GCM with a per-deployment key derived via HKDF(OCU_MASTER_KEY, OCU_SALT) |
| **Integrity** | GCM authentication tag — any ciphertext modification causes decryption failure |
| **Determinism** | SHA-256(plaintext) → token_id — same input, same token, across all requests |
| **Key rotation** | Change OCU_MASTER_KEY + OCU_SALT → all existing tokens become un-decryptable (requires vault reset) |
| **Zero-egress** | Vault is always local — no PII is sent to any remote store |

Fail-closed invariant: if `StoreToken` returns an error, the refinery blocks the request and returns HTTP 500. Un-redacted data is never forwarded upstream.

---

## 9. Entity Registry (Path 3)

The vault also hosts the **Entity Registry** — a secondary table that maps name variants (`"John"`, `"Doe"`, `"J. Doe"`) to a single canonical token (`[PERSON_1]`). This solves identity fragmentation across documents.

Entity tokens use a numeric suffix (`_1`, `_2`) instead of a hash suffix. They bypass AES encryption — `GetEntityByToken("[PERSON_1]")` returns the canonical name string directly.

For a full walkthrough, see [Entity Registry Guide](./ENTITY_REGISTRY_GUIDE.md).
