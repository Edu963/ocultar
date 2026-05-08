# OCULTAR | PII Detection & Compliance Glossary

This document provides a transparency disclosure of the PII types detected by the OCULTAR Refinery and their respective regulatory compliance mappings.

> [!IMPORTANT]
> **EU Sovereign Detection Pack (v1)** is now active. This pack provides deterministic coverage and checksum-backed validation for core EU and UK identifiers, satisfying GDPR Art. 9 and local sovereignty requirements (CNIL, AEPD, BDSG).

## 1. Detection Tiers

The Refinery operates on a multi-tier defense-in-depth model:

| Tier | Name | Methodology | `method` tag |
|---|---|---|---|
| **Tier 0.1** | Base64 Evasion Shield | Decodes and recursively scans every Base64 segment in the payload. PII hidden inside encoded blobs is caught and re-encoded after redaction — the payload structure is preserved, the data is not. | *(inherits inner tier's tag)* |
| **Tier 0** | Dictionary Shield | Exact match against Enterprise VIP and exclusion lists. | `"dictionary"` |
| **Tier 1** | Deterministic Registry | High-performance Go regular expressions with **checksum-backed validation** (Luhn mod-10 for credit cards, MOD97 for IBANs, and 12 national ID validators). Pattern matches that fail their checksum are discarded rather than flagged — eliminating false positives at the source. | `"regex"` or `["regex","checksum"]` |
| **Tier 1.1** | Phone Shield | libphonenumber validation for international and localized formats. Runs after Tier 1 to avoid misidentifying digit sequences already claimed by national IDs. | `"phone"` |
| **Tier 1.2** | Address Shield | Heuristic street address parser supporting EN/FR/ES/DE formats. | `"address"` |
| **Tier 1.5** | Greeting/Signature Shield | Extracts names from salutations (`"Regards, Jean"`) and self-introductions (`"My name is..."`). | `"greeting"` |
| **Tier 2** | AI Semantic Scan | Contextual NER via a pluggable Python sidecar. Default model: `openai/privacy-filter` (bidirectional token classifier, Apache 2.0). `piiranha-v1` is supported as a multilingual alternative for mixed-language corpora. Activated via `TIER2_ENGINE=privacy-filter` and `--profile ai` in Docker Compose. | `"ai-ner"` |
| **Tier 3** | Structural Heuristics | Context-aware proximity rules and entity expansion (professional titles, possessives, conjunction linkage). | `"structural"` |

The `method` tag appears in each `DetectionResult.method` field returned by `/api/refine` and displayed in the **Detection Attribution** panel of the dashboard. It tells you exactly which pipeline tier caught each entity.

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
| `[IBAN_...]` | Financial | International Bank Account Numbers. Validated with MOD97 checksum. | GDPR, PCI-DSS |
| `[CREDIT_CARD_...]` | Financial | Credit card numbers (Visa, Mastercard, Amex, Discover, JCB). Every candidate is validated with the **Luhn algorithm (mod-10 checksum)** before vaulting — sequences that fail are not redacted, eliminating false positives. | PCI-DSS, GDPR |
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
- **Same Input = Same Token**: The token for any PII value is computed as `[TYPE_sha256[:8]]`. The same email address, SSN, or card number always produces the same token — across requests, sessions, and documents. Relational integrity is fully preserved in tokenized datasets.
- **Irreversible without the Vault**: Token strings contain no original data. Reversal requires both the **Identity Vault** (DuckDB/PostgreSQL) and the **Master Key** — neither of which ever leaves your infrastructure.

### Privacy-Safe Analytics on Tokenized Data

Because tokens are deterministic, **you can perform aggregations, joins, and frequency analysis directly on tokenized data — without ever de-tokenizing it.**

A dataset where every email has been replaced with `[EMAIL_9c8f7a1b]` still supports:
- **Counting unique users** — distinct token values = distinct PII values
- **Joining across tables** — the same person appears with the same token in every table
- **Frequency analysis** — which customers appear most often, without exposing who they are
- **Anomaly detection** — token-level clustering and outlier detection on sensitive fields

Re-hydration to plaintext is only required when a human must read a specific value. All analytical workloads can remain in the tokenized domain. This property is a direct consequence of the SHA-256-based token design in `getOrSetSecureResult` (`services/refinery/pkg/refinery/refinery.go`) and requires no additional configuration.

## 5. Auditor Verification Note

Every policy update in OCULTAR goes through the **Validation-First DAG**:
1.  **Simulation**: Proposed changes are replayed against the last 1,000 requests to predict impact.
2.  **Signing**: Final policies are signed with Ed25519 to prevent tampering.
3.  **Audit**: The `compliance-integrity-suite` continuously monitors for configuration drift and runtime violations.

## 6. Performance & SLA

To satisfy enterprise latency and reliability requirements, the Tier 2 AI Scan implements deterministic SLA enforcement:
- **Thirty-Second Timeout**: Every AI scan is bounded by a strict 30-second `http.Client` timeout (hardcoded in `inference/remote.go`). Configurable at the application level via `inference_timeout` in `config.yaml` (currently advisory only).
- **Fail-Closed Strategy**: If the scan exceeds this budget or the sidecar is unreachable, the request fails with a security error, preventing un-scanned data from being processed.
- **Session Cache**: Results are stored in a thread-safe `sync.Map` keyed by the original input string. Repeat scans within the same request session are sub-millisecond and bypass the network.
- **Single-Pass Batch Scan**: For complex JSON records, the SLM is called once per record (not per string field) to reduce round-trips and preserve relational token integrity.
