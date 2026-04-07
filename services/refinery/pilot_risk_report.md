# OCULTAR Data Risk Assessment Report

> **CONFIDENTIAL — For Authorised Recipients Only**
> This report constitutes a technical risk and privacy assessment based on automated analysis. It is informational in nature and does not constitute legal advice or a regulatory compliance determination. Distribution is restricted to named stakeholders.

---

## Report Metadata

| Field | Value |
| :--- | :--- |
| **Report ID** | OCU-86AE0323 |
| **Generated** | 07 April 2026, 18:03 UTC |
| **Dataset Scope** | `../../datasets/leaky_demo.json` |
| **Records Analysed** | 4 |
| **Methodology Version** | v3.1 |
| **Engine** | OCULTAR Enterprise v1.14 |

---

## Executive Risk Summary

> [!CAUTION]
> **Overall Risk Level: CRITICAL (9.8/10)**
> **Compliance Likelihood: ⚠️ High Non-Compliance Likelihood (External Processing Scenarios)**




The dataset identified in this report contains an estimated **4 records** that fall below commonly cited EU pseudonymization thresholds. In its current state, this data **presents elevated risk for use with external AI systems and LLM APIs** without prior sanitisation.

The estimated financial exposure associated with unauthorised disclosure of this dataset is in the range of **€15338 – €198900** (simulated estimate based on industry breach benchmarks). This range encompasses regulatory exposure modelling, operational incident response costs, and a risk multiplier derived from the dataset's anonymization profile. Actual impact may vary significantly based on enforcement context and organisational factors.

---

## Risk Scorecard

| Category | Score | Level | Business Implication |
| :--- | :---: | :---: | :--- |
| **Identifiability Risk** | 9.5/10 | CRITICAL | K=1: Each individual is uniquely identifiable from their quasi-attributes alone. This represents the highest re-identification risk level. Common industry benchmarks suggest K≥3–5 as a minimum for basic pseudonymization. |
| **Financial Sensitivity** | 10.0/10 | CRITICAL | An estimated 100% of assessed records contain high-sensitivity attributes (financial identifiers, personal names). Exposure of this dataset in an unprotected context would likely present significant regulatory and operational risk. Subject to contextual factors including jurisdiction, data controller agreements, and processing purpose. |
| **Re-identification Risk** | 10.0/10 | CRITICAL | Estimated re-identification attack surface: 10/10 (based on K-Anonymity=1 and L-Diversity=1). An adversary with partial auxiliary knowledge may be able to isolate individuals with elevated probability. This is a modelled estimate; actual risk depends on the adversary model and available external data. |
| **Compliance Readiness** | 10.0/10 | CRITICAL | An estimated 100% of records (4/4) fall below commonly cited minimum pseudonymization thresholds. This dataset presents a high likelihood of non-compliance in external processing scenarios. Regulatory exposure depends on jurisdiction, processing context, and applicable exemptions. |
| **Overall** | **9.8/10** | **CRITICAL** | Weighted composite score (Identifiability 35%, Financial 25%, Re-id 25%, Compliance 15%) |

---

## Technical Metrics — Interpreted

### K-Anonymity
**Raw Score:** 1

K-Anonymity = 1 — K=1: Each individual is uniquely identifiable from their quasi-attributes alone. This represents the highest re-identification risk level. Common industry benchmarks suggest K≥3–5 as a minimum for basic pseudonymization.

> **Industry Benchmark:** Common industry frameworks suggest a minimum K-score of 3–5 for basic pseudonymization. This is a technical benchmark, not a mandatory legal threshold—contextual factors, processing purpose, and applicable exemptions determine actual compliance obligations.

### L-Diversity
**Raw Score:** 1

L-Diversity = 1 — Sensitive attributes appear homogeneous within quasi-identifier groups. This increases the risk of homogeneity attacks, where an adversary may infer the sensitive value of any individual in a group.

> **Industry Benchmark:** An L-Diversity score of ≥2 is commonly recommended to mitigate homogeneity attacks, as referenced in ISO/IEC 29101 (Privacy Architecture Framework). This is an industry guideline; applicable legal thresholds depend on jurisdictional context.

---

## Financial Exposure Model

The **Value at Risk (VaR)** range below is computed using a three-component methodology anchored to industry breach cost benchmarks. All figures are **simulated estimates** and should not be interpreted as predicted fine amounts or contractual commitments.

### VaR Components

| Component | Methodology | Min Estimate | Max Estimate |
| :--- | :--- | ---: | ---: |
| **Regulatory Exposure** | Dataset Risk Score (0.98) × anchor range (€10,000–€100,000) | **€9825** | **€98250** |
| **Operational Cost** | €100–€300 × 4 records (industry benchmark range) | **€400** | **€1200** |
| **Risk Multiplier** | Derived from K=1, L=1 profile | **1.5×** | **2.0×** |
| | | | |
| **Total Value at Risk (Estimated)** | | **€15338** | **€198900** |

> **Assumptions & Methodology Note:**
> This estimate is based on industry breach cost benchmarks (€100–€300/record range from published incident studies) and a regulatory exposure simulation anchor (€10,000–€100,000 base, scaled by dataset risk score). The risk multiplier (range: 1.0×–2.0×) is derived from K-Anonymity and L-Diversity scores. Actual financial impact may vary significantly depending on enforcement context, data controller obligations, jurisdiction, organisational size, and mitigating controls in place. These figures are simulated estimates and do not constitute legal or financial advice.

---

## AI & LLM Exposure Assessment

### Decision: BLOCK

| Parameter | Assessment |
| :--- | :--- |
| **External LLM API Safety** | CRITICAL risk |
| **Internal Copilot Safety** | 🚫 Not recommended without sanitisation |
| **Vector DB / RAG Indexing** | 🚫 Not recommended without prior processing |

**RAG & Vector Database Guidance:**
Blocked for RAG indexing. Embedding raw PII (names, IBANs, emails) into a vector database creates queryable re-identification surfaces that are difficult to retroactively remove. This assessment is based on the detected K-Anonymity and L-Diversity levels.

**Recommended Action:**
Do not transmit this dataset to any external API or internal AI copilot system without first running the full OCULTAR redaction pipeline. This recommendation is based on technical risk assessment, not a legal determination.

---

## Before / After Simulation

This section demonstrates the modelled impact of the OCULTAR Enterprise pipeline on your dataset's risk profile. Figures are projected estimates based on typical processing outcomes.

| Metric | Scenario A — Current State (No Protection) | Scenario B — After OCULTAR Processing |
| :--- | :--- | :--- |
| **Risk Level** | 🔴 CRITICAL | 🟢 LOW |
| **Risk Score** | 9.8 / 10 | 0.5 – 1.5 / 10 (projected) |
| **Financial Exposure (VaR)** | €15338 – €198900 (estimated) | €307 – €1227 (projected residual) |
| **AI / LLM Status** | BLOCK | ALLOW |

**What changes:**
- **Before:** The raw dataset as-is, transmitted directly to an LLM API or stored in a vector database. All PII fields are exposed in plaintext.
- **After:** After OCULTAR tokenization and format-preserving encryption pipeline. Direct identifiers are removed.

---

## Assumptions

The following assumptions underpin all quantitative estimates in this report:

| Assumption | Value / Range | Basis |
| :--- | :--- | :--- |
| **Regulatory anchor (low)** | €10,000 | Simulation baseline |
| **Regulatory anchor (high)** | €100,000 | Simulation ceiling |
| **Operational cost per record** | €100–€300 | Industry study range |
| **Pseudonymization threshold** | K≥3, L≥2 | Common benchmark |

---

## Remediation Plan

Dataset presents elevated re-identification risk. Recommended remediation steps (in order of priority):
  1. **Tokenization** — Replace Names, IBANs, and Email addresses with OCULTAR vault tokens.
  2. **Generalization** — Replace precise Region values with broader geographic tiers to increase K-Anonymity group size.
  3. **Format-Preserving Encryption (FPE)** — Encrypt IBAN fields while preserving format for analytics.
  All steps above can be automated via the OCULTAR Enterprise pipeline.

---

## Appendix: Methodology & Standards

This report applies the following analytical frameworks:

- **K-Anonymity** (Sweeney, 2002)
- **L-Diversity** (Machanavajjhala et al., 2006)
- **GDPR Article 5(1)(f)**
- **ISO/IEC 29101**

> This report was generated automatically by OCULTAR Enterprise v1.14. technical assessment only.

---

*OCULTAR Enterprise v1.14 | Methodology v3.1 | Report ID: OCU-86AE0323*
*Generated: 07 April 2026, 18:03 UTC*
