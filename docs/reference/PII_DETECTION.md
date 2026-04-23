# OCULTAR | PII Detection & Compliance Glossary

This document provides a transparency disclosure of the PII types detected by the OCULTAR Refinery and their respective regulatory compliance mappings.

> [!IMPORTANT]
> **EU Sovereign Detection Pack (v1)** is now active. This pack provides deterministic coverage and checksum-backed validation for core EU and UK identifiers, satisfying GDPR Art. 9 and local sovereignty requirements (CNIL, AEPD, BDSG).

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
| `[SSN_...]` | Identity | Social Security Numbers. Supports both hyphenated (`XXX-XX-XXXX`) and raw 9-digit (`XXXXXXXXX`) formats. Utilizes contextual triggers to ensure high-fidelity detection. | GDPR, HIPAA, Tax Compliance |
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
| `[PL_PESEL_...]` | Identity | Polish National Identification Numbers (PESEL). | GDPR |
| `[FI_HETU_...]` | Identity | Finnish Personal Identity Codes. | GDPR |
| `[SE_PIN_...]` | Identity | Swedish Personal Identity Numbers. | GDPR |
| `[DK_CPR_...]` | Identity | Danish Personal Identification Numbers. | GDPR |
| `[NO_FNR_...]` | Identity | Norwegian Birth Numbers (FNR). | GDPR |
| `[BR_CPF_...]` | Identity | Brazilian Individual Taxpayer Registry (CPF). | LGPD |
| `[CL_RUT_...]` | Identity | Chilean National ID (RUT). | LPDP |
| `[INDIA_AADHAAR_...]` | Identity | Indian Aadhaar Numbers (12-digit). | Digital Personal Data Protection Act |
| `[SINGAPORE_ID_...]` | Identity | Singapore National ID (NRIC/FIN). | PDPA |
| `[US_PASSPORT_...]` | Identity | US Passport Numbers. | Privacy Act of 1974 |
| `[US_DL_...]` | Identity | US Driver's License Numbers. | Driver's Privacy Protection Act |
| `[AWS_KEY_...]` | Security | AWS Access Key IDs. | SOC2, PCI-DSS |
| `[AWS_SECRET_...]` | Security | AWS Secret Access Keys. | SOC2, PCI-DSS |
| `[GCP_SERVICE_ACCOUNT_...]` | Security | GCP Service Account Emails. | SOC2 |
| `[IP_ADDRESS_...]` | Digital Identity | IPv4 addresses. | GDPR, CCPA |

## 3. Canonical InfoType Mapping (Google Cloud DLP)

For enterprise compliance parity, OCULTAR internal types are mapped to **Google Cloud InfoTypes**. This registry is visible in the Operational Dashboard and available via `/api/config/mapping`.

| Ocultar ID | Google InfoType Equivalent |
|---|---|
| `EMAIL` | `EMAIL_ADDRESS` |
| `SSN` | `US_SOCIAL_SECURITY_NUMBER` |
| `IBAN` | `IBAN_CODE` |
| `CREDIT_CARD` | `CREDIT_CARD_NUMBER` |
| `IP_ADDRESS` | `IP_ADDRESS` |
| `LOCATION` | `LOCATION` |
| ... | (Full list of 30+ mappings in Registry) |

## 4. Redaction Behavior

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
