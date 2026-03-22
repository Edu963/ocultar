---
name: privacy-risk-analyzer
description: Expert analyzer for dataset re-identifiability. Executes mathematical K-Anonymity and L-Diversity checks to ensure GDPR/HIPAA compliance for structured data.
---

# Privacy Risk Analyzer (v1.1)

## Purpose

The PRA evaluates the technical privacy guarantees of a dataset. It ensures that individuals cannot be uniquely identified through combinations of quasi-identifiers (ZIP, Birthday, Gender).

## Inputs / Outputs

### Inputs
- `dataset` (List[JSON]): Active records for analysis.
- `quasi_identifiers`: Fields that could lead to re-identification.
- `sensitive_attributes`: Fields containing private data.

### Outputs
- `risk_report` (Artifact): Mathematical summary (K, L).
- `mitigation_config` (JSON): Precise bucketization or masking rules.
- `verdict` (Enum): `SAFE` | `RISKY` | `UNSAFE`.

## Preconditions
- Dataset MUST have at least 10 records for meaningful K-Anonymity analysis.
- `quasi_identifiers` list MUST NOT be empty.

---

## Instructions

### 1. K-Anonymity Analysis
- Group records by `quasi_identifiers`.
- **Measurement**: Smallest group size = $K$.
- **Verdict**: If $K < 3$ (GDPR) or $K < 5$ (HIPAA), return `UNSAFE`.

### 2. L-Diversity Check
- Count unique values of `sensitive_attributes` within each $K$-group.
- **Goal**: At least $L$ distinct values (Prevent attribute inference).
- **Verdict**: If $L < 2$, return `RISKY`.

### 3. Automated Mitigation Logic
Construct a `mitigation_config` to reach target $K$:
- **Bucketization**: "Age: 25" -> "Age: 20-30".
- **Suppression**: Remove high-entropy fields (e.g., `Exact_Timestamp`).
- **Logic**: If `verdict` != `SAFE`, generate the required `REDACT` or `STRIP` rules for the Sombra Gateway.

## Failure Handling
- **Data Sparsity**: If the dataset is too small (<10), return `INSUFFICIENT_DATA` error.
- **Attribute Conflict**: If a field is both a QI and a Sensitive Attribute, signal `CONFIRMED_LEAK`.

## Postconditions
- `mitigation_config` MUST be compatible with `refinery-architecture-manager` input.
