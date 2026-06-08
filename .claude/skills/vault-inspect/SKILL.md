---
name: vault-inspect
description: |
  Vault Inspection Tool. Safely queries the local DuckDB or PostgreSQL vault.
  Lists token counts, types, and encryption status.
  Use when: "check vault", "list tokens", "vault status", "inspect database".
allowed-tools:
  - Bash
  - Read
  - Grep
triggers:
  - check vault
  - inspect tokens
  - vault status
---

# /vault-inspect — Sovereign Vault Monitoring

You are the **Data Custodian** for Ocultar. You verify that the Vault is storing ciphertext correctly and that the token mapping is healthy.

## Instructions

### Step 1: Detect Vault Backend
Read `configs/config.yaml` to see if we are using DuckDB or PostgreSQL.

```bash
grep "backend:" configs/config.yaml
```

### Step 2: Query Vault (DuckDB)
If DuckDB is used, run a query to count tokens.

```bash
# Get the vault path from config
VAULT_PATH=$(grep "path:" configs/config.yaml | awk '{print $2}')
# Query token counts by type
duckdb "$VAULT_PATH" "SELECT pii_type, COUNT(*) as count FROM tokens GROUP BY pii_type ORDER BY count DESC;"
```

### Step 3: Verify Encryption
Check a sample record to ensure it is encrypted (not plaintext).

```bash
duckdb "$VAULT_PATH" "SELECT token_id, pii_type, length(ciphertext) as len FROM tokens LIMIT 5;"
```

## Report Format
Produce a **Vault Status Report**:
- **Backend**: DuckDB / PostgreSQL.
- **Token Inventory**: Table of types and counts.
- **Storage Health**: Total records, database file size.
- **Encryption Check**: Confirmation that ciphertext is being stored.
