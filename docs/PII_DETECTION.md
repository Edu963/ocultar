# OCULTAR | PII Detection & Compliance Glossary

This document provides a transparency disclosure of the PII types detected by the OCULTAR Refinery and their respective regulatory compliance mappings.

## 1. Detection Tiers

The Refinery operates on a multi-tier defense-in-depth model:

| Tier | Name | Methodology |
|---|---|---|
| **Tier 0** | Dictionary Shield | Exact match against Enterprise VIP and exclusion lists. |
| **Tier 1** | Lead Shield | High-performance Go regular expressions. |
| **Tier 2** | AI Semantic Scan | Local SLM (Qwen/Phi) performing NER inference. |
| **Tier 3** | Structural Heuristics | Context-aware proximity rules and entity expansion. |

## 2. PII Type Glossary & Compliance Mapping

| Token Type | Category | Description | Compliance Requirement |
|---|---|---|---|
| `[PERSON_...]` | Identity | Names, surnames, and identity fragments. | GDPR Art. 4(1), HIPAA |
| `[EMAIL_...]` | Digital Identity | Email addresses. | GDPR Art. 4(1), CCPA |
| `[PHONE_...]` | Contact | International and localized phone numbers. | GDPR Art. 4(1), HIPAA |
| `[ADDRESS_...]` | Location | Physical street addresses, cities, and zip codes. | GDPR Art. 4(1), HIPAA |
| `[HEALTH_ENTITY_...]` | Special Category | Medical professional titles and facility names. | GDPR Art. 9, HIPAA |
| `[SENSITIVE_EVENT_...]` | Special Category | Sensitive life events (Divorce, Marriage, Medical treatments). | GDPR Art. 9 |
| `[TRANSACTION_CODE_...]` | Financial | Account numbers, transaction IDs, financial triggers. | PCI-DSS, GDPR |
| `[INTERNAL_PROJECT_...]` | Business Secret | Proprietary project names or internal code names. | Trade Secret Protection |
| `[SSN_...]` | Identity | Social Security Numbers (US format). | GDPR, HIPAA, Tax Compliance |
| `[CREDENTIAL_...]` | Security | Passwords and authentication secrets. | OWASP, PCI-DSS, ISO 27001 |
| `[SECRET_...]` | Security | API keys, tokens, and cryptographic secrets. | OWASP, PCI-DSS, ISO 27001 |
| `[IBAN_...]` | Financial | International Bank Account Numbers. | GDPR, PCI-DSS |
| `[EU_VAT_...]` | Financial | EU and UK Value Added Tax numbers. | GDPR, Tax Compliance |
| `[FR_NIR_...]` | Identity | French Social Security Numbers (NIR). | GDPR Art. 9, CNIL |
| `[ES_DNI_...]` | Identity | Spanish National Identity Numbers (DNI/NIE/CIF). | LOPD, GDPR |
| `[DE_STEUER_ID_...]` | Identity | German Tax Identification Numbers. | GDPR, BDSG |
| `[IT_CODICE_FISCALE_...]` | Identity | Italian Fiscal Codes. | GDPR, Codice in materia di protezione dei dati personali |
| `[NL_BSN_...]` | Identity | Dutch Citizen Service Numbers (BSN). | GDPR, UAVG |
| `[UK_NINO_...]` | Identity | UK National Insurance Numbers. | UK GDPR, HMRC |
| `[UK_NHS_...]` | Identity | UK National Health Service numbers. | UK GDPR, NHS Data Security |

## 3. Redaction Behavior

OCULTAR always uses **deterministic pseudonymization**:
- **Same Input = Same Token**: Preserves relational integrity for data science and AI context.
- **Irreversible**: Token strings contain no original data and are reversible only with access to both the **Identity Vault** (DuckDB/PostgreSQL) and the **Master Key**.

## 4. Auditor Verification Note

Every policy update in OCULTAR goes through the **Validation-First DAG**:
1.  **Simulation**: Proposed changes are replayed against the last 1,000 requests to predict impact.
2.  **Signing**: Final policies are signed with Ed25519 to prevent tampering.
3.  **Audit**: The `compliance-integrity-suite` continuously monitors for configuration drift and runtime violations.

## 5. Performance & SLA

To satisfy enterprise latency and reliability requirements, the Tier 2 AI Scan implements deterministic SLA enforcement:
- **Five-Second Timeout**: Every AI scan is bounded by a strict 5-second `http.Client` timeout.
- **Fail-Closed Strategy**: If the scan exceeds this budget, the request fails with a security error, preventing un-scanned data from being processed.
- **Deterministic Caching**: Results are stored in a thread-safe `sync.Map` keyed by the **SHA-256 hash** of the input. Repeat scans are sub-millisecond, bypass the network, and preserve relational token integrity.
- **Strict Single-Pass Audit**: Validated by regression tests to ensure Tier 2 never re-scans partial fragments already covered by the batch record scan.
