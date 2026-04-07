# OCULTAR Data Risk Assessment Report

> **CONFIDENTIAL — For Authorised Recipients Only**
> This report constitutes a formal risk and compliance assessment. Distribution is restricted to named stakeholders.

---

## Report Metadata

| Field | Value |
| :--- | :--- |
| **Report ID** | OCU-9A1DA4E1 |
| **Generated** | 07 April 2026, 03:31 UTC |
| **Dataset Scope** | `datasets/leaky_demo.json` |
| **Records Analysed** | 4 |
| **Methodology Version** | v3.0 |
| **Engine** | OCULTAR Enterprise v1.14 |

---

## Executive Risk Summary

> [!CAUTION]
> **Overall Risk Level: CRITICAL (9.8/10)**
> **Compliance Status: ❌ NON-COMPLIANT (GDPR Art. 5)**




The dataset identified in this report contains **4 records** that fail minimum EU anonymisation thresholds. In its current state, this data **cannot safely be used with external AI systems and LLM APIs** without a material risk of regulatory breach and personal data exposure.

The estimated financial exposure associated with unauthorised disclosure of this dataset is **€73370**, comprising regulatory fines, forensic response costs, and reputational damage.

---

## Risk Scorecard

| Category | Score | Level | Business Implication |
| :--- | :---: | :---: | :--- |
| **Identifiability Risk** | 9.5/10 | CRITICAL | K=1: Every individual is uniquely identifiable from their quasi-attributes. Maximum re-identification risk. |
| **Financial Sensitivity** | 10.0/10 | CRITICAL | 100% of records contain high-sensitivity attributes (financial identifiers, personal names). Exposure of this dataset would constitute a significant regulatory breach. |
| **Re-identification Risk** | 10.0/10 | CRITICAL | Combined K-Anonymity=1 and L-Diversity=1 yield a re-identification attack surface of 10/10. An adversary with partial knowledge could isolate individuals with high probability. |
| **Compliance Readiness** | 10.0/10 | CRITICAL | 100% of records (4/4) fail minimum anonymisation requirements. Regulatory reporting of this dataset in its current state would constitute a GDPR violation under Art. 5(1)(f). |
| **Overall** | **9.8/10** | **CRITICAL** | Weighted composite score (Identifiability 35%, Financial 25%, Re-id 25%, Compliance 15%) |

---

## Technical Metrics — Interpreted

### K-Anonymity
**Raw Score:** 1

K-Anonymity = 1 → K=1: Every individual is uniquely identifiable from their quasi-attributes. Maximum re-identification risk.

> **Regulatory Threshold:** A minimum K-score of 3 is required under EU statistical disclosure guidelines and is the recommended baseline for GDPR Article 89 research exemptions.

### L-Diversity
**Raw Score:** 1

L-Diversity = 1 → Sensitive attributes are homogeneous within groups. Homogeneity attacks are feasible. An adversary knows the sensitive value of any individual in a group.

> **Regulatory Threshold:** An L-Diversity score of ≥2 is required to mitigate homogeneity attacks as recommended under ISO/IEC 29101 (Privacy Architecture Framework).

---

## Financial Exposure Model

The **Value at Risk (VaR)** is computed using a Three-Pillar methodology aligned with Ponemon Institute (2024) cost benchmarks and GDPR Article 83 fine exposure estimates.

| Pillar | Formula | Estimated Exposure |
| :--- | :--- | ---: |
| **Regulatory Fines** | €10,000 × 4 violating records | **€40000** |
| **Operational Cost** | €250 × 4 total records (forensics, notification, legal) | **€1000** |
| **Brand Damage** | (Regulatory + Operational) × Risk-weighted multiplier × 40% | **€32370** |
| | | |
| **Total Value at Risk** | | **€73370** |

> **Methodology Note:** The Brand Damage multiplier scales with Overall Risk Score (range: 0.5x–2.0x) and represents projected customer churn, competitive disadvantage, and crisis management costs based on industry sector benchmarks.

---

## AI & LLM Exposure Assessment

### Decision: BLOCK

| Parameter | Assessment |
| :--- | :--- |
| **External LLM API Safety** | CRITICAL risk |
| **Internal Copilot Safety** | 🚫 Blocked — must not be used |
| **Vector DB / RAG Indexing** | 🚫 Blocked for indexing |

**RAG & Vector Database Guidance:**
BLOCKED for RAG indexing. Embedding raw PII (names, IBANs, emails) into a vector database creates permanent, queryable re-identification attack surfaces that cannot be retroactively purged.

**Recommended Action:**
Do not transmit this dataset to any external API or internal copilot system. Run full OCULTAR redaction pipeline before any AI usage.

---

## Before / After Simulation

This section demonstrates the measurable impact of the OCULTAR Enterprise pipeline on your dataset's risk profile.

| Metric | Scenario A — Current State (No Protection) | Scenario B — After OCULTAR Processing |
| :--- | :--- | :--- |
| **Risk Level** | 🔴 CRITICAL | 🟢 LOW |
| **Risk Score** | 9.8 / 10 | 0.5 / 10 (projected) |
| **Financial Exposure (VaR)** | €73370 | €1467 (projected residual) |
| **AI / LLM Status** | BLOCK | ALLOW |

**What changes:**
- **The raw dataset as-is, transmitted directly to an LLM API or stored in a vector database. All PII fields are exposed in plaintext.**
- **After OCULTAR&#39;s tokenization and format-preserving encryption pipeline. All PII replaced with reversible vault tokens. Dataset becomes K=∞ Anonymous.**

> The projected 95% risk reduction and 98% VaR reduction are based on OCULTAR's K=∞ Tokenization model, where all PII is replaced with reversible vault-encrypted references. Residual exposure represents standard operational overhead.

---

## Remediation Plan

Dataset is highly re-identifiable. Recommended remediation (in order of priority):
  1. **Tokenization** — Replace Names, IBANs, and Email addresses with OCULTAR vault tokens.
  2. **Generalization** — Replace precise Region values with broader geographic tiers.
  3. **Format-Preserving Encryption (FPE)** — Encrypt IBAN fields while preserving format for analytics.
  All steps above can be automated via the OCULTAR Enterprise pipeline.

---

## Appendix: Methodology & Standards

This report applies the following analytical frameworks:

- **K-Anonymity** (Sweeney, 2002): Minimum group size for quasi-identifier equivalence classes.
- **L-Diversity** (Machanavajjhala et al., 2006): Sensitive attribute diversity within equivalence classes.
- **GDPR Article 5(1)(f)**: Integrity and confidentiality principle.
- **GDPR Article 83**: Administrative fine schedule (up to 4% global annual turnover or €20M).
- **Ponemon Institute 2024 Cost of a Data Breach Report**: Forensic cost benchmarks.
- **ISO/IEC 29101**: Privacy architecture framework for ICT systems.

---

*OCULTAR Enterprise v1.14 | Methodology v3.0 | Report ID: OCU-9A1DA4E1*
*This report was generated automatically. Findings are based on the dataset provided at time of analysis.*
