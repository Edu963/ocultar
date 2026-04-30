# GDPR Article 25 — Privacy by Design and by Default

**Document classification:** Public — for vendor security review and DPA annex  
**Version:** 1.0  
**Last updated:** 2026-04-28  
**Contact:** security@ocultar.dev

---

## 1. Purpose

This document maps OCULTAR's technical architecture to the requirements of **GDPR Article 25** (Privacy by Design and by Default). It is intended for:

- Data Protection Officers (DPOs) evaluating OCULTAR as a technical measure
- CISOs conducting vendor security review
- Legal teams preparing Data Processing Agreements (DPAs)
- Enterprise customers documenting compliance with GDPR Art. 30 records of processing

---

## 2. GDPR Article 25 — Normative Text

### Article 25(1) — Privacy by Design

> *"The controller shall … implement appropriate technical and organisational measures, such as pseudonymisation, which are designed to implement data-protection principles, such as data minimisation, in an effective manner and to integrate the necessary safeguards into the processing in order to meet the requirements of this Regulation and protect the rights of data subjects."*

### Article 25(2) — Privacy by Default

> *"The controller shall implement appropriate technical and organisational measures for ensuring that, by default, only personal data which are necessary for each specific purpose of the processing are processed … In particular, such measures shall ensure that by default personal data are not made accessible without the individual's intervention to an indefinite number of persons."*

---

## 3. Architecture-to-Requirement Mapping

### 3.1 Pseudonymisation (Art. 25(1), Recital 78)

| Requirement | OCULTAR Implementation |
|---|---|
| Personal data must be pseudonymised where possible | Every PII value is replaced with a deterministic token: `[TYPE_token_id]` (e.g. `[EMAIL_9c8f7a1b]`). The token is a SHA-256 digest of the plaintext, truncated to 8 hex characters. The mapping between token and plaintext is stored only in the encrypted vault — never transmitted. |
| Pseudonymisation must be reversible only by authorised parties | Token re-hydration (reveal) requires `OCU_AUDITOR_TOKEN`, which must match the server-side secret. Every reveal call is logged in the Ed25519-signed audit trail with actor, timestamp, and payload hash. |
| Re-identification must require additional information | The vault master key (`OCU_MASTER_KEY`) is required to decrypt any stored ciphertext. Without it, tokens are opaque and non-invertible. Key is stored in AWS Secrets Manager or equivalent — never in the container image or environment plaintext. |

### 3.2 Data Minimisation (Art. 25(1), Art. 5(1)(c))

| Requirement | OCULTAR Implementation |
|---|---|
| Only necessary personal data reaches the processing purpose | The upstream AI provider (OpenAI, Gemini, etc.) receives only the tokenised payload. Raw PII is never included in any outbound request. |
| Minimisation must be enforced, not optional | The proxy pipeline is synchronous and blocking: `redactBody()` must return successfully before the upstream connection is opened. There is no configuration flag to disable redaction on a live instance. |
| Scope of data shared with third parties is limited | The only third-party call is to the AI provider configured in `OCU_PROXY_TARGET`. That call receives tokens, not PII. No telemetry, analytics, or side-channel egress exists. |

### 3.3 Fail-Closed Default (Art. 25(2))

Art. 25(2) requires that the *default* behavior protects data subjects, not that it can be configured to do so.

| Default behavior | OCULTAR enforcement |
|---|---|
| No PII forwarded by default | If `redactBody()` returns any error, the proxy responds HTTP 503 and closes the connection without opening the upstream socket. |
| Vault unavailability blocks the request | If the vault write fails (encrypted storage of token ↔ ciphertext mapping), the request is rejected. A token that cannot be stored cannot be safely issued. |
| AI scanner unavailability does not bypass detection | If the Tier 2 SLM sidecar times out or is unreachable, the proxy returns HTTP 500. It does not fall back to forwarding un-scanned text. |
| Queue saturation blocks new requests | If the internal processing queue is full, the proxy responds HTTP 429. It does not silently drop the redaction step. |

Every failure mode is fail-closed. There is no graceful degradation path that could result in PII reaching the upstream.

### 3.4 Encryption at Rest (Art. 25(1), Art. 32(1)(a))

| Requirement | OCULTAR Implementation |
|---|---|
| Personal data must be protected at rest | Vault stores only AES-256-GCM ciphertexts. Plaintext PII values are never written to disk, database, or log. |
| Encryption keys must be managed separately from data | The master key (`OCU_MASTER_KEY`) is provided via environment variable at runtime — injected from AWS Secrets Manager, Doppler, or equivalent. It is not stored alongside the vault data. |
| Key derivation must prevent cross-context key reuse | A per-deployment salt (`OCU_SALT`) is combined with `OCU_MASTER_KEY` via HKDF-SHA256 to derive the vault encryption key. The same master key with a different salt produces a cryptographically independent derived key, isolating environments. |

### 3.5 Accountability and Audit Trail (Art. 25(1), Art. 5(2), Art. 30)

| Requirement | OCULTAR Implementation |
|---|---|
| Controller must demonstrate compliance (accountability) | The Ed25519-signed audit log provides a tamper-evident record of every refine and reveal operation: actor, timestamp, payload hash, and digital signature over the full event. |
| Audit records must be verifiable | The log is hash-chained: each entry contains the SHA-256 of the previous entry. Any deletion, modification, or insertion breaks the chain and is detected by the built-in verifier (`--verify-audit`). |
| Records of processing activities (Art. 30) | Each log entry records the processing purpose (refine vs. reveal), token identifiers affected, and timestamp — sufficient to reconstruct the processing record for a given data subject. |

### 3.6 SSRF and Network Boundary Enforcement (Art. 25(1), Art. 32(1)(b))

OCULTAR's SSRF protection prevents the proxy from being used to exfiltrate tokenised or partially redacted data to internal network targets, which would undermine the zero-egress guarantee.

Blocked target classes:
- RFC 1918 private ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
- IPv4 loopback (127.0.0.0/8)
- IPv6 loopback (::1)
- IPv6 link-local (fe80::/10)
- Unspecified addresses (0.0.0.0, ::)
- DNS rebinding — the resolved IP is checked after DNS resolution, not just the hostname

---

## 4. Art. 25 Compliance Summary Matrix

| GDPR Obligation | Satisfied By | Verifiable Evidence |
|---|---|---|
| Pseudonymisation | SHA-256 deterministic token replacement | Token format `[TYPE_xxxxxxxx]` in all upstream payloads |
| Data minimisation | Zero-egress proxy: upstream receives tokens only | Network capture shows no raw PII in upstream requests |
| Privacy by default | Fail-closed on every error path | 9 automated fail-closed tests in `services/refinery/fail_closed_test.go` |
| Encryption at rest | AES-256-GCM vault with HKDF-SHA256 | Vault file contains no readable plaintext |
| Key separation | OCU_MASTER_KEY injected at runtime via secrets manager | CloudFormation template in `docs/distribution/AWS_MARKETPLACE.md` |
| Accountability | Hash-chained Ed25519 audit log | `--verify-audit` produces pass/fail with chain index |
| SSRF prevention | RFC 1918 / loopback / link-local blocklist | 14 automated SSRF tests in `services/refinery/ssrf_test.go` |

---

## 5. EU AI Act Alignment

Where OCULTAR is deployed as the data boundary layer before an AI system (e.g. with Sombra Gateway or an external LLM API), it contributes to compliance with the following EU AI Act (2024/1689) obligations:

| EU AI Act Article | Obligation | OCULTAR Contribution |
|---|---|---|
| Art. 10 — Data Governance | Training/inference data must be relevant, representative, and free of errors that could result in discrimination | OCULTAR ensures that PII is not transmitted to AI systems, reducing the risk that personal data is incorporated into model outputs or retained by the AI provider. |
| Art. 12 — Record-Keeping | High-risk AI systems must keep logs sufficient to identify causes of incidents | The Ed25519-signed, hash-chained audit log provides a verifiable processing record for every prompt that passed through OCULTAR before reaching the AI system. |
| Art. 13 — Transparency | Deployers of high-risk AI must inform natural persons that they interact with an AI system | OCULTAR does not address user-facing transparency; that obligation remains with the AI deployer. OCULTAR's audit trail supports post-hoc accountability if transparency obligations are contested. |
| Art. 26 — Deployer Obligations | Deployers must implement appropriate technical and organisational measures | OCULTAR's zero-egress architecture is precisely the type of technical measure Art. 26(2) contemplates: it restricts the categories of personal data processed by the AI system. |

---

## 6. Data Processing Agreement — Short-Form Template

The following template may be used as Annex 1 (Technical and Organisational Measures) to a DPA between the enterprise customer (Controller) and a third-party AI provider. It documents that OCULTAR satisfies the controller's Art. 25 obligations before data is transferred to the processor.

---

### Annex 1 — Technical and Organisational Measures (GDPR Art. 32 / Art. 28(3)(c))

**Data controller:** [Customer legal entity name]  
**Sub-processor receiving data:** [AI provider name, e.g. OpenAI Ireland Ltd.]  
**PII pre-processing layer:** OCULTAR Refinery v[X.Y.Z], operated by Controller within Controller's infrastructure

---

**Measure 1 — Pseudonymisation before transfer**

All personal data is pseudonymised by the OCULTAR Refinery before transmission to the sub-processor. The sub-processor receives only opaque tokens (e.g. `[EMAIL_9c8f7a1b]`). The pseudonymisation key (vault master key) is held exclusively by the Controller and is not shared with the sub-processor.

**Measure 2 — Zero-egress guarantee**

The OCULTAR proxy is configured fail-closed: any failure in the pseudonymisation step results in the request being rejected with HTTP 503 before the sub-processor connection is opened. No raw personal data is transmitted to the sub-processor under any failure condition.

**Measure 3 — Encryption at rest**

Pseudonymisation mappings (token ↔ encrypted plaintext) are stored in an AES-256-GCM encrypted vault under a HKDF-SHA256 derived key. The vault is hosted within the Controller's infrastructure and is not accessible to the sub-processor.

**Measure 4 — Audit trail**

Every pseudonymisation and de-pseudonymisation event is recorded in a SHA-256 hash-chained, Ed25519-signed audit log. Log integrity is verifiable via the built-in chain verification tool. The log constitutes the Controller's Art. 30 processing record for AI-assisted operations.

**Measure 5 — Access control for de-pseudonymisation**

Re-identification (token reveal) requires a separate auditor token (`OCU_AUDITOR_TOKEN`) not accessible to the sub-processor. All reveal operations are logged.

---

*This annex was prepared using the OCULTAR GDPR Article 25 Compliance Pack. The Controller is responsible for ensuring that the OCULTAR deployment described herein is maintained in accordance with this annex.*

---

## 7. Scope and Limitations

This compliance mapping covers OCULTAR's technical measures only. It does not constitute legal advice and does not replace:

- A full Data Protection Impact Assessment (DPIA) for high-risk processing under Art. 35
- Art. 30 Records of Processing Activities maintained by the Controller
- Controller-level policies for data subject rights (access, erasure, portability)
- Assessment of the upstream AI provider's own compliance posture

For EU sovereign deployments requiring national identifier validation (ES DNI, FR NIR, IT Codice Fiscale, DE Steuer-ID, NL BSN, PL PESEL, UK NINO/NHS), see `docs/reference/EU_SOVEREIGN_PACK_V1.md`.
