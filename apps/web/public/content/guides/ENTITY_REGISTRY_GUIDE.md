# Entity Registry — Persistent Identity Resolution Guide

> **Audience:** Developers and operators who need consistent PII masking across name fragments, sessions, and documents for known identities (patients, employees, customers).

---

## Table of Contents

1. [The Problem: Token Fragmentation](#1-the-problem-token-fragmentation)
2. [The Solution: Entity Registry](#2-the-solution-entity-registry)
3. [How It Works](#3-how-it-works)
4. [Registering Entities via the Sombra API](#4-registering-entities-via-the-sombra-api)
5. [Bulk Seeding from a CRM or Patient Roster](#5-bulk-seeding-from-a-crm-or-patient-roster)
6. [Listing Registered Entities](#6-listing-registered-entities)
7. [How Rehydration Works for Entity Tokens](#7-how-rehydration-works-for-entity-tokens)
8. [Token Format: Hash vs. Entity](#8-token-format-hash-vs-entity)
9. [Coverage: Which Entity Types Use the Registry](#9-coverage-which-entity-types-use-the-registry)
10. [Database Schema](#10-database-schema)
11. [Use Cases](#11-use-cases)
12. [Security Notes](#12-security-notes)

---

## 1. The Problem: Token Fragmentation

By default, OCULTAR's vault is a string-to-token map. Every unique string produces an independent token based on its SHA-256 hash:

```
"John"     → [PERSON_8d9c1b15]
"Doe"      → [PERSON_7f83b1aa]
"John Doe" → [PERSON_91ba89fc]
```

This causes two critical failures when processing real-world documents:

**PII Leakage** — If a document contains "John" in one sentence and "John Doe" in another, and only "John Doe" is detected by the NER engine, "John" leaks as plaintext.

**Context Fragmentation** — Even when all variants are masked, the downstream LLM receives three different tokens and has no way to know they refer to the same person. Clinical summaries, CRM analysis, and HR workflows produce incoherent results.

---

## 2. The Solution: Entity Registry

The Entity Registry is a persistent, session-spanning identity layer built into the vault. It maps all known variants of an identity — first name, last name, full name, initials, aliases — to a single **canonical token**:

```
"John"     → [PERSON_1]   ← same token
"Doe"      → [PERSON_1]   ←
"John Doe" → [PERSON_1]   ←
"J. Doe"   → [PERSON_1]   ←
```

Canonical tokens use a **numeric suffix** (`[PERSON_1]`, `[PERSON_2]`) — distinct from the 8-char hex suffix of hash-based tokens (`[PERSON_8d9c1b15]`). This distinction lets the system route rehydration correctly.

---

## 3. How It Works

The entity registry is checked **before** the SHA-256 hash path in `getOrSetSecureResult`. For any PERSON-class entity type, the refinery calls `vault.LookupVariant(value)` first:

```
Refinery detects "John" (via SLM, Tier-1.5, or Tier-0 dictionary)
         │
         ▼
entityRegistryTypes["PERSON"] = true?
         │
         ├─ YES → LookupVariant("John")
         │              │
         │         ┌────┴────────────────────────┐
         │         │ Found → return [PERSON_1]   │  ← entity token (numeric)
         │         │ Miss  → fall through        │
         │         └─────────────────────────────┘
         │
         └─ NO  → SHA256("John")[:8] → [PERSON_8d9c1b15]   ← hash token
```

Entity tokens are stored in `canonical_entities` (not the `vault` table) and rehydrate directly to the `canonical_name` without AES decryption.

---

## 4. Registering Entities via the Sombra API

### `POST /v1/entities`

Register a single canonical entity with its known name variants.

**Request body:**
```json
{
  "entity_type":    "PERSON",
  "canonical_name": "John Doe",
  "variants":       ["John", "Doe", "J. Doe", "Mr. Doe"]
}
```

**Response:**
```json
{
  "canonical_token": "[PERSON_1]"
}
```

**Example:**
```bash
curl -X POST http://localhost:8086/v1/entities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <actor-token>" \
  -d '{
    "entity_type":    "PERSON",
    "canonical_name": "John Doe",
    "variants":       ["John", "Doe", "J. Doe"]
  }'
```

**Registration is idempotent.** Registering the same `canonical_name` again merges new variants and returns the existing token — it does not create a duplicate entity or change the token.

**Supported entity types:** `PERSON`, `ORGANIZATION`. Additional types follow the same pattern — the `entity_type` becomes the token prefix (e.g., `ORGANIZATION_1`).

---

## 5. Bulk Seeding from a CRM or Patient Roster

### `POST /v1/entities/seed`

Seed multiple entities in a single call. Accepts either a flat JSON array or a `{"entities": [...]}` wrapper.

**Request body (array):**
```json
[
  {
    "entity_type":    "PERSON",
    "canonical_name": "Jane Smith",
    "variants":       ["Jane", "Smith", "J. Smith", "Dr. Smith"]
  },
  {
    "entity_type":    "PERSON",
    "canonical_name": "Robert Johnson",
    "variants":       ["Robert", "Rob", "Johnson", "R. Johnson"]
  },
  {
    "entity_type":    "ORGANIZATION",
    "canonical_name": "Acme Corporation",
    "variants":       ["Acme", "Acme Corp", "Acme Co."]
  }
]
```

**Response:**
```json
{
  "seeded": 3,
  "tokens": ["[PERSON_1]", "[PERSON_2]", "[ORGANIZATION_1]"]
}
```

**Example — load a patient roster from a file:**
```bash
curl -X POST http://localhost:8086/v1/entities/seed \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d @patient_roster.json
```

**Seeding is safe to re-run.** Each entity is inserted idempotently — running the same seed twice produces identical tokens and does not create duplicate records.

---

## 6. Listing Registered Entities

### `GET /v1/entities`

Returns all registered canonical entities and their variant lists. Useful for audit trails, admin dashboards, and verifying roster coverage.

**Response:**
```json
[
  {
    "id":             "PERSON_1",
    "entity_type":    "PERSON",
    "canonical_name": "John Doe",
    "variants":       ["Doe", "J. Doe", "John", "Mr. Doe"]
  },
  {
    "id":             "PERSON_2",
    "entity_type":    "PERSON",
    "canonical_name": "Jane Smith",
    "variants":       ["Dr. Smith", "J. Smith", "Jane", "Smith"]
  },
  {
    "id":             "ORGANIZATION_1",
    "entity_type":    "ORGANIZATION",
    "canonical_name": "Acme Corporation",
    "variants":       ["Acme", "Acme Co.", "Acme Corp"]
  }
]
```

Variants are returned alphabetically. Entities are ordered by type then by sequential ID.

---

## 7. How Rehydration Works for Entity Tokens

`DecryptToken` checks the token format before looking up the vault:

```
Token: "[PERSON_1]"
         │
         ├─ Matches numeric suffix regex? → YES
         │
         ▼
    GetEntityByToken("[PERSON_1]")
         │
         ├─ Found → return canonical_name ("John Doe")
         └─ Miss  → fall through to AES vault lookup (safe)
```

**Entity tokens do not require AES decryption.** The `canonical_name` is stored as plaintext in `canonical_entities` — it is management-plane data seeded by an operator, not user-submitted PII captured at runtime.

---

## 8. Token Format: Hash vs. Entity

Two token formats now exist in the system:

| Format | Example | Source | Rehydration |
|---|---|---|---|
| Hash token | `[PERSON_8d9c1b15]` | SHA-256 of original value, 8-char hex | AES-256-GCM decrypt from `vault` table |
| Entity token | `[PERSON_1]` | Sequential integer, operator-registered | Direct lookup from `canonical_entities` table |

Both formats are valid OCULTAR tokens. They coexist in the same document — a document may contain both `[PERSON_1]` (a registered patient) and `[EMAIL_9c8f7a1b]` (an email detected at runtime).

The rehydration layer (`DecryptToken`) checks the numeric suffix regex first and routes accordingly. No configuration is required — the routing is automatic.

---

## 9. Coverage: Which Entity Types Use the Registry

The registry lookup is activated for these entity types:

| Type | Description |
|---|---|
| `PERSON` | Names detected by any tier |
| `PERSON_VIP` | Names from the Tier-0 protected dictionary |
| `HEALTH_ENTITY` | Clinical names (patient, provider) |
| `PROTECTED_ENTITY` | CRM-synced identities |
| `ORGANIZATION` | Company and institution names |

Types not in this list (e.g., `EMAIL`, `PHONE`, `IBAN`, `SSN`) always use the hash-based path — they are structural identifiers where canonical cross-session linking is not required.

---

## 10. Database Schema

The entity registry adds two tables to the vault database (DuckDB or PostgreSQL). Both are created automatically on startup.

```sql
-- One row per canonical identity
CREATE TABLE IF NOT EXISTS canonical_entities (
    id             VARCHAR PRIMARY KEY,       -- "PERSON_1", "ORGANIZATION_2"
    entity_type    VARCHAR NOT NULL,          -- "PERSON", "ORGANIZATION"
    canonical_name VARCHAR NOT NULL UNIQUE,   -- "John Doe"
    created_at     TIMESTAMP DEFAULT current_timestamp
);

-- One row per known variant; many variants map to one canonical entity
CREATE TABLE IF NOT EXISTS entity_variants (
    variant_name VARCHAR PRIMARY KEY,   -- "John", "Doe", "J. Doe"
    canonical_id VARCHAR NOT NULL,      -- references canonical_entities.id
    created_at   TIMESTAMP DEFAULT current_timestamp
);

-- Fast O(1) lookup during stream parsing (case-insensitive LOWER() comparison)
CREATE INDEX IF NOT EXISTS idx_ev_canonical_id ON entity_variants(canonical_id);
```

`LookupVariant` uses `WHERE LOWER(variant_name) = LOWER(?)` — the lookup is **case-insensitive**. "john", "John", and "JOHN" all resolve to the same entity.

---

## 11. Use Cases

### Healthcare — Patient Roster

Seed your patient list at startup. Every mention of a patient — first name, last name, full name, "Dr." prefix — collapses to a single token. The LLM receives a coherent, anonymized record.

```json
[
  {
    "entity_type":    "PERSON",
    "canonical_name": "Maria Garcia",
    "variants":       ["Maria", "Garcia", "M. Garcia", "Ms. Garcia", "Patient Garcia"]
  }
]
```

Clinical notes like "Maria was seen today... Garcia's lab results... the patient" now all tokenize as `[PERSON_1]` when the SLM identifies them.

### Legal — Case Participants

Register all parties before processing depositions or contracts. Opposing counsel, witnesses, and defendants all get stable canonical tokens — enabling privacy-safe cross-document analysis.

### HR — Employee Records

Seed the employee directory. HR documents, performance reviews, and org charts all reference the same person with the same token, enabling analytics without exposing identities.

### CRM — Account Contacts

Pair with the CRM Sync worker (`services/refinery/pkg/identities/sync.go`) to automatically ingest contact names from your CRM on a schedule. The entity registry provides the stable cross-session token layer on top.

---

## 12. Security Notes

**Plaintext storage of canonical names** — `canonical_names` are stored unencrypted in `canonical_entities`. This is intentional: these are management-plane identities seeded by an authorized operator, not raw PII captured from user prompts. The vault's AES-256-GCM layer protects hash-based tokens; entity names are analogous to an allowlist configuration.

If your threat model requires canonical names to be encrypted at rest, use filesystem-level encryption for `vault.db` (e.g., LUKS or BitLocker) or encrypt the entire PostgreSQL tablespace.

**Idempotency is not a security boundary** — the seed API accepts repeated registrations of the same entity. An attacker with write access to `/v1/entities/seed` could overwrite variants (though not canonical names, which are unique-constrained). Protect the Sombra admin endpoints with the `OCU_JWT_SECRET` bearer token in production.

**No cross-entity collisions** — `variant_name` has a `UNIQUE` constraint in both DuckDB and PostgreSQL. A variant string can only map to one canonical entity. Attempting to assign the same variant to two different entities returns a conflict error from the database layer.
