# Vault Architecture

The Vault is OCULTAR's secure, immutable storage engine for PII and sensitive tokens. It is designed to be **decryption-resistant** while enabling authorized data rehydration.

## Core Design Principles

### 1. Zero-Egress Storage
The Vault is strictly local. OCULTAR does not provide a cloud-based storage option for PII. Data resides solely within your infrastructure (on-premise or within your VPC).

### 2. Immutable Cryptographic Proof
Every write operation to the Vault generates a hash that is signed by the **Sombra Private Key**. This creates a tamper-evident audit log that is defensible during compliance audits.

### 3. Distributed Enterprise Vaulting (Phase 4)
In Enterprise deployments, the Vault can be distributed across multiple regional clusters with automated synchronization and RBAC (Role-Based Access Control) rehydration rules.

## Storage Backends

OCULTAR supports multiple storage engines depending on your performance and durability needs:

- **DuckDB (Default)**: Best for single-tenant local pilots. High-performance, low-overhead analytics engine.
- **PostgreSQL**: Recommended for enterprise production. Supports distributed transactions and mature backup/restore procedures.

### Configuration
```yaml
settings:
  vault_backend: "postgres"
  postgres_dsn: "host=db.corp.local port=5432 user=ocultar dbname=ocultar_vault sslmode=require"
```

---

## The Rehydration Pattern

"Rehydration" is the process of replacing a token (e.g., `[PERSON_1]`) back with its original value (`Héctor Eduardo`) when an authorized user is viewing the response from the AI.

1. **Detection**: Sombra identifies a token in the AI's response matching a pattern in the current session's vault.
2. **Authorize**: Sombra checks the user's `Identity Token` against the Vault's RBAC policy.
3. **Lookup**: If authorized, Sombra retrieves the raw value from the backend.
4. **Rehydrate**: The token is swapped for the raw value before being sent to the client.

> [!SECURITY]
> The Vault uses **AES-256-GCM** encryption for all PII at rest. The encryption key is rotated every 90 days by default, or can be managed via external KMS (Key Management Systems).
