# API Reference

Ocultar exposes a local HTTP API on the port you specify at startup (default `4141`).
All endpoints accept and return JSON unless noted otherwise.

---

## Authentication

| Endpoint | Auth required |
|---|---|
| `GET /api/health` | None |
| `POST /api/refine` | None |
| `POST /api/reveal` | Bearer token (`OCU_AUDITOR_TOKEN`) |
| `GET /api/entities` | Bearer token (`OCU_AUDITOR_TOKEN`) |
| `POST /api/entities` | Bearer token (`OCU_AUDITOR_TOKEN`) |
| `POST /api/entities/seed` | Bearer token (`OCU_AUDITOR_TOKEN`) |

Pass the auditor token in the `Authorization` header:

```
Authorization: Bearer <OCU_AUDITOR_TOKEN>
```

If `OCU_AUDITOR_TOKEN` is not set on the server, all authenticated endpoints return `403 Forbidden`.

---

## `GET /api/health`

Returns engine status. No authentication required.

**Response `200 OK`:**

```json
{
  "status": "healthy",
  "version": "1.14",
  "vault": { "status": "online" },
  "slm":   { "status": "online", "circuit": "closed" }
}
```

| Field | Description |
|---|---|
| `status` | `"healthy"` or `"degraded"` |
| `version` | Engine version string |
| `vault.status` | `"online"` or `"offline"` |
| `slm.status` | `"online"`, `"offline"`, or `"disabled"` |
| `slm.circuit` | `"closed"` (normal) or `"open"` (Tier 2 disabled after repeated failures) |

---

## `POST /api/refine`

Detect and mask PII in text or JSON. No authentication required.

**Request body:** Raw UTF-8 text string, or any JSON value.

**Response `200 OK`:**

```json
{
  "refined": "{\"message\":\"Hello [PERSON_3a12b4cd], your balance is [FINANCIAL_9c8f7a1b].\"}",
  "report": {
    "hits": 2,
    "types": ["PERSON", "FINANCIAL"]
  }
}
```

| Field | Description |
|---|---|
| `refined` | JSON-encoded string containing the masked payload. Parse it once to get the masked text. |
| `report.hits` | Number of PII entities detected and masked |
| `report.types` | List of entity type labels found |

> `refined` is always a JSON-encoded string, even when the input was plain text. This is by design — it ensures the response is safe to embed in any JSON context without escaping issues.

**Token format:** `[TYPE_xxxxxxxx]` where `TYPE` is the entity category (e.g. `PERSON`, `EMAIL`, `SSN`, `IBAN`, `PHONE`) and `xxxxxxxx` is an 8-character hex digest derived from the original value. The same input always produces the same token within a vault instance.

**Error responses:**

| Status | Meaning |
|---|---|
| `500` | Vault unavailable or refinery error — request blocked (fail-closed) |
| `413` | Request body exceeds the 10 MB size limit |

---

## `POST /api/reveal`

Restore one or more vault tokens back to their original plaintext values.

**Authentication:** `Authorization: Bearer <OCU_AUDITOR_TOKEN>` required.

Every call to this endpoint is recorded in the immutable Ed25519-signed audit log.

**Request body:**

```json
{
  "tokens": ["[PERSON_3a12b4cd]", "[EMAIL_9c8f7a1b]"]
}
```

**Response `200 OK`:**

```json
{
  "results": {
    "[PERSON_3a12b4cd]": "Alice Martin",
    "[EMAIL_9c8f7a1b]": "alice@example.com"
  }
}
```

Tokens not found in the vault return the string `"ERR_NOT_FOUND"` as their value in the `results` map.

**Error responses:**

| Status | Meaning |
|---|---|
| `403` | Missing or invalid auditor token |
| `400` | Malformed request body |

---

## `GET /api/entities`

List all entries in the persistent entity registry.

**Authentication:** `Authorization: Bearer <OCU_AUDITOR_TOKEN>` required.

**Response `200 OK`:**

```json
[
  {
    "id": "PERSON_1",
    "entity_type": "PERSON",
    "canonical_name": "Alice Martin",
    "variants": ["Alice", "A. Martin", "alice martin"]
  }
]
```

---

## `POST /api/entities`

Add a new entry to the persistent entity registry.

Pre-seeding entities ensures that all variants of a name or identifier map to the same token consistently across sessions.

**Authentication:** `Authorization: Bearer <OCU_AUDITOR_TOKEN>` required.

**Request body:**

```json
{
  "entity_type": "PERSON",
  "canonical_name": "Alice Martin",
  "variants": ["Alice", "A. Martin"]
}
```

**Response `200 OK`:**

```json
{
  "canonical_token": "[PERSON_3a12b4cd]"
}
```

---

## `POST /api/entities/seed`

Bulk-insert multiple entity registry entries in one call.

**Authentication:** `Authorization: Bearer <OCU_AUDITOR_TOKEN>` required.

**Request body:** JSON array of entity objects, or `{"entities": [...]}` wrapper — same schema as `POST /api/entities`.

**Response `200 OK`:**

```json
{
  "seeded": 12,
  "tokens": ["[PERSON_3a12b4cd]", "[PERSON_7e2f1a9b]"]
}
```

---

## Entity type labels

| Label | Examples |
|---|---|
| `PERSON` | Full names, given names, surnames |
| `EMAIL` | Email addresses |
| `PHONE` | Phone numbers (all country formats via libphonenumber) |
| `SSN` | US Social Security Numbers |
| `IBAN` | International Bank Account Numbers |
| `CREDIT_CARD` | Payment card numbers (Luhn-validated) |
| `ADDRESS` | Street addresses (EN/FR/ES/DE heuristics) |
| `DATE_OF_BIRTH` | Birth dates |
| `PASSPORT` | Passport numbers |
| `NATIONAL_ID` | National ID numbers (50+ country formats) |
| `IP_ADDRESS` | IPv4 and IPv6 addresses |
| `FINANCIAL` | Account numbers, routing numbers, financial identifiers |

Full list of 117 supported types: `data/pii_coverage.json` in the repository.

---

## Size limits

| Limit | Value |
|---|---|
| Maximum request body | 10 MB |
| Maximum extracted text (documents) | 500 KB |
| Request timeout | 30 s |
