---
name: privacy-risk-analyzer
description: Expert analyzer for dataset re-identifiability. Executes mathematical K-Anonymity and L-Diversity checks to ensure GDPR/HIPAA compliance for structured data.
---

# Privacy Risk Analyzer (v1.0)

## Purpose

The Privacy Risk Analyzer evaluates the mathematical privacy guarantees of a dataset. It ensures that data remains anonymous even against "linkage attacks" by verifying that individuals cannot be uniquely identified through combinations of quasi-identifiers.

## Prerequisites

- **Dataset**: A structured list of JSON records (e.g., from an audit log or database export).
- **Core Engine Access**: Access to `services/engine/pkg/audit/risk.go` for the underlying logic references.

## Inputs / Outputs

### Inputs:
- `dataset` (List[JSON]): The collection of records to analyze.
- `quasi_identifiers` (List[String]): Fields that could lead to re-identification (e.g., `ZIP`, `Age`, `Gender`).
- `sensitive_attributes` (List[String]): Fields containing private data (e.g., `Diagnosis`, `Salary`).

### Outputs:
- `risk_report` (Artifact):
    - `k_anonymity`: The minimum size of an anonymity set.
    - `l_diversity`: The minimum number of distinct sensitive values per anonymity set.
    - `is_compliant`: Boolean based on threshold (Default: K>=3, L>=2).
    - `recommendation`: Guidance on further bucketization or encryption.

---

## Instructions

### Step 1 – Quasi-Identifier Grouping
- Group all records in the `dataset` by the combination of their `quasi_identifiers`.
- Create "Buckets" where all records have identical (or wildcarded) quasi-identifier values.

### Step 2 – K-Anonymity Calculation
- Measure the size of each bucket.
- **Goal**: Every record must exist in a bucket of size $K$.
- **Validation**: Identify the smallest bucket. This value is the `k_anonymity` score.
- **Threshold**: If $K < 3$, flag as **High Re-identification Risk**.

### Step 3 – L-Diversity Evaluation (Homogeneity Attack Prevention)
- For each bucket, count the number of unique values in the `sensitive_attributes`.
- **Goal**: Each anonymity set must have at least $L$ distinct sensitive values to prevent attackers from inferring the actual value even if they know the group.
- **Validation**: Identify the minimum diversity across all buckets. This is the `l_diversity` score.
- **Threshold**: If $L < 2$, flag as **Homogeneity Risk**.

### Step 4 – GDPR/HIPAA Alignment
- Cross-reference the scores against standard thresholds:
    - **GDPR Statistical Safety**: $K \ge 3$.
    - **HIPAA Safe Harbor**: $K \ge 5$ (Recommended) or expert determination.
- Finalize the `is_compliant` status.

### Step 5 – Mitigation Recommendations
- If non-compliant, suggest:
    - **Generalization**: Converting specific values to ranges (e.g., `Age 25` -> `Age 20-30`).
    - **Suppression**: Removing specific fields from the output.
    - **Format-Preserving Encryption (FPE)**: Pseudonymizing identifiers while maintaining data structure.

---

## Failure Handling

- **Empty Dataset**: Halt analysis and return a null report.
- **Missing Quasi-Identifiers**: If no quasi-identifiers are provided, the analysis cannot proceed; request field selection.

---

## Examples

### Example 1: Compliant Dataset
- **Input**: 1000 records; QI: `["City", "Year_of_Birth"]`.
- **Finding**: Smallest bucket size is 12 (K=12); each bucket has 5 distinct diagnoses (L=5).
- **Result**: `is_compliant: true`.

### Example 2: Re-identification Risk
- **Input**: 10 records; QI: `["ZIP", "Exact_Age"]`.
- **Finding**: A bucket contains only 1 record (K=1).
- **Result**: `is_compliant: false`. Recommendation: "Bucketize Age into 5-year ranges."
