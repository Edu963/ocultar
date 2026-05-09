# OCULTAR | EU Sovereign AI Deployment — GDPR & CNIL Compliance Narrative
## French Financial Sector Edition

**Document type:** Data Protection Officer (DPO) Technical Evidence  
**Regulatory scope:** GDPR (EU) 2016/679 · CNIL deliberation n°2020-091 · PCI-DSS v4.0  
**Classification:** Internal — Legal & Compliance  

---

## 1. Executive Summary

OCULTAR is a zero-egress PII detection and pseudonymization proxy. It is deployed between your internal systems and any AI model API (whether hosted internally or externally). Its function is to ensure that **no personal data as defined by GDPR Art. 4(1) ever leaves your controlled infrastructure in identifiable form**.

This document is intended for the Data Protection Officer (DPO), legal counsel, and the CISO. It maps OCULTAR's technical architecture to the specific obligations of your organization under GDPR, CNIL guidance, and PCI-DSS as they apply to the use of AI language models in a French financial context.

---

## 2. What OCULTAR Does — Plain Language

When a user or application sends a request to an AI model (e.g., "Analyze this contract: Jean Dupont, IBAN FR76 3000 6000 0112 3456 7890 189..."), OCULTAR intercepts the request **before it reaches the model**. It:

1. **Detects** every piece of personal data in the payload using a multi-tier engine (deterministic regex, Luhn/MOD97 checksum validation, libphonenumber, heuristics, and optional AI NER).
2. **Replaces** each detected value with a deterministic pseudonymous token, e.g., `[IBAN_12ab34cd]`.
3. **Stores** the encrypted original value in a local vault (AES-256-GCM). The vault never leaves your infrastructure.
4. **Forwards** the tokenized payload to the AI model. The model receives no personal data.
5. **Optionally re-hydrates** the AI response — replacing tokens back with original values — for authorized callers only.

The AI model — whether OpenAI, Mistral, or an internal model — is treated as an **untrusted external processor**. It receives only pseudonymous tokens.

---

## 3. GDPR Compliance Mapping

### Art. 4(1) — Definition of Personal Data

OCULTAR detects and pseudonymizes all categories of personal data relevant to the French financial sector:

| Data Category | OCULTAR Token | Regulatory Basis |
|---|---|---|
| Full name | `[PERSON_...]` | GDPR Art. 4(1) |
| Email address | `[EMAIL_...]` | GDPR Art. 4(1) |
| Phone number (FR +33) | `[PHONE_...]` | GDPR Art. 4(1) |
| IBAN (MOD97 validated) | `[IBAN_...]` | GDPR Art. 4(1), PCI-DSS |
| French SSN / NIR | `[FR_NIR_...]` | GDPR Art. 9, CNIL |
| SIRET / SIREN (Luhn validated) | `[FRANCE_SIRET_NUMBER_...]` / `[FRANCE_SIREN_NUMBER_...]` | GDPR Art. 4(1) |
| EU VAT number | `[EU_VAT_...]` | GDPR Art. 4(1) |
| BIC / SWIFT code | `[BIC_...]` | GDPR Art. 4(1), PCI-DSS |
| Credit / debit card (Luhn validated) | `[CREDIT_CARD_...]` | PCI-DSS, GDPR |
| IP address | `[IP_ADDRESS_...]` | GDPR Art. 4(1), CJEU Breyer judgment |
| API keys / credentials | **Blocked (HTTP 403)** | OWASP, PCI-DSS |
| Health / sensitive life events | **Blocked (HTTP 403)** | GDPR Art. 9 |

Detection uses **checksum-backed validation**: IBANs that fail MOD97 and credit card numbers that fail the Luhn algorithm are not flagged, eliminating false positives at the source.

---

### Art. 25 — Data Protection by Design and by Default

GDPR Art. 25 requires that personal data protection be built into the processing architecture from the outset, not added as an afterthought.

**OCULTAR satisfies Art. 25 through:**

| Requirement | OCULTAR implementation |
|---|---|
| Pseudonymization at the earliest possible stage | PII is tokenized **before** the payload reaches any AI model. The model never sees personal data. |
| Minimal data exposure | Only the pseudonymous token is forwarded. The original value remains in the local vault. |
| Default-protected configuration | `OCU_PILOT_MODE=0` requires explicit opt-in for Tier 2 AI. Community mode operates entirely on deterministic rules — no model inference. |
| Data minimization | Only entities classified as PII are replaced. Structural and non-personal data passes through unchanged. |
| Purpose limitation enforcement | The policy-as-code engine enforces explicit rules (e.g., block credentials, block Art. 9 special categories) that cannot be bypassed at runtime. |

---

### Art. 32 — Security of Processing

GDPR Art. 32 requires appropriate technical measures to ensure a level of security appropriate to the risk.

| Measure | OCULTAR implementation |
|---|---|
| **Encryption at rest** | AES-256-GCM. Each PII value is encrypted with a key derived via HKDF-SHA256 from `OCU_MASTER_KEY` and `OCU_SALT`. |
| **Pseudonymization** | Token format: `[TYPE_sha256[:8]]`. The token contains no original data; reversal requires the vault and the master key. |
| **Integrity of the audit log** | Every vault operation and policy block is signed with Ed25519. Log entries cannot be tampered with without detection. |
| **Fail-closed design** | If the AI sidecar is unavailable, the request fails with a security error. Data is never processed un-scanned. |
| **Transport security** | All inter-service communication uses TLS. The proxy validates upstream certificates. |
| **Access control** | Re-hydration (token → plaintext) requires the vault and the master key. Neither is exposed to external parties. |
| **Concurrency limits** | Configurable semaphore (`max_concurrency`) prevents resource exhaustion under load. |

---

### Art. 9 — Special Categories of Personal Data

GDPR Art. 9 imposes stricter obligations on health data, biometric data, and data revealing sensitive personal characteristics.

OCULTAR's French Finance policy configuration includes an explicit **block** rule for Art. 9 categories:

```yaml
- name: block-special-category-data
  when:
    entity: [HEALTH_ENTITY, SENSITIVE_EVENT]
    min_confidence: 0.8
  action: block
```

A request containing medical professional titles, facility names, or sensitive life events (divorce, medical treatments) is **rejected with HTTP 403** before any AI model processes it. The rejection is logged with action `POLICY_BLOCK` in the immutable audit trail.

This approach treats Art. 9 compliance as a hard technical gate, not a procedural control.

---

## 4. CNIL-Specific Requirements

### Délibération n°2020-091 — AI Systems and Personal Data

The CNIL has established that AI systems processing personal data must implement:

| CNIL requirement | OCULTAR response |
|---|---|
| Transparency of processing | The compliance evidence endpoint (`GET /api/compliance/evidence`) provides a machine-readable audit snapshot: active policies, vault entry count, and recent audit log entries. |
| Minimization before AI processing | Handled by the proxy architecture: the AI model receives only pseudonymous tokens. |
| Local processing priority | OCULTAR operates entirely on-premises. No data transits external infrastructure during tokenization or storage. |
| Accountability | The Ed25519-signed audit log provides cryptographic proof of what was processed, when, and by whom. |

### French NIR (Numéro d'Inscription au Répertoire)

The NIR (French Social Security Number) is a special-category identifier under CNIL guidance. OCULTAR detects all NIR formats including spaced variants (e.g., `1 85 06 75 115 423 18`) using the pattern registered in the EU Sovereign Detection Pack. Detection uses format validation (key digit verification). All detected NIRs are tokenized to `[FR_NIR_...]` and stored encrypted in the local vault.

---

## 5. PCI-DSS v4.0 Alignment

| PCI-DSS Requirement | OCULTAR implementation |
|---|---|
| Req. 3.4 — Render PAN unreadable | Credit card numbers are detected via Luhn-validated regex and replaced with `[CREDIT_CARD_...]` tokens before reaching any AI model or logging system. |
| Req. 3.5 — Protect stored data | Vault entries are AES-256-GCM encrypted. The encryption key is derived via HKDF — it is never stored in the database. |
| Req. 10 — Audit trails | Every tokenization event is logged with timestamp, actor (IP/user), entity type, and token ID. Logs are append-only and Ed25519-signed. |
| Req. 12.3.2 — Risk assessment | The compliance evidence endpoint provides a machine-readable control snapshot for quarterly PCI assessments. |

---

## 6. Data Subject Rights (GDPR Chapter III)

| Right | How OCULTAR supports it |
|---|---|
| **Art. 15 — Right of access** | The vault maps each token to an encrypted ciphertext. De-tokenization requires the master key and vault access — both controlled by the data controller. |
| **Art. 17 — Right to erasure** | Deleting a vault entry removes the only link between the token and the original value. The token becomes permanently irreversible. |
| **Art. 20 — Data portability** | Tokenized datasets can be exported. The original values can be re-hydrated on demand by authorized callers with vault access. |

---

## 7. The Sovereignty Guarantee

"Zero-egress" means the following is guaranteed by architecture, not by policy:

1. The vault (encrypted PII storage) runs on your infrastructure. It is a DuckDB or PostgreSQL database under your control.
2. The master key (`OCU_MASTER_KEY`) never leaves your environment. It is loaded from your secrets manager or environment variable at startup and never written to disk.
3. The AI model (external or internal) receives only pseudonymous tokens. It cannot reconstruct any personal data from a token without the vault and the master key.
4. OCULTAR makes no outbound connections except to the explicitly configured upstream AI endpoint.

This architecture satisfies the "appropriate safeguards" requirement of GDPR Art. 46 for international data transfers: the data that reaches the AI model (whether hosted in the US, EU, or elsewhere) is not personal data as defined by Art. 4(1).

---

## 8. DPO Sign-Off Checklist

- [ ] `OCU_MASTER_KEY` is set to a high-entropy value (≥ 32 bytes) and stored in a secrets manager (not in `.env` committed to source control)
- [ ] `OCU_SALT` is set to a unique per-deployment value
- [ ] `OCU_AUDIT_PRIVATE_KEY` is set and the audit log path is write-protected
- [ ] Active policies reviewed and approved: `configs/config.french-finance.yaml`
- [ ] `GET /api/compliance/evidence` integrated with your SOC 2 / ISO 27001 audit tool
- [ ] Vault backup procedure defined (DuckDB file or PostgreSQL dump)
- [ ] Data Processing Agreement (DPA) completed with your AI model provider noting that no personal data is transmitted (only pseudonymous tokens)

---

*This document reflects OCULTAR v1.14. For the latest compliance mapping, re-run `GET /api/compliance/evidence` against your deployed instance.*
