# 🧪 Multi-Tier Refinery Pipeline

The Refinery Pipeline is the heart of OCULTAR. It uses a defense-in-depth approach to identify and tokenize PII before it reaches any external provider.

## Pipeline Tiers

The pipeline is organized into tiers of increasing complexity and "intelligence."

### Tier 0.1: Evasion Defense (Base64)
- **What it does**: Scans for Base64 or JWT blobs within the text.
- **Why**: Attackers or developers might accidentally/intentionally hide PII inside encoded segments to bypass simple filters.
- **Action**: Decodes the segment, runs the full refinery pipeline on the contents, and re-encodes it.

### Tier 0-0.5: Fast Deterministic Sweep
- **Tier 0 (Dictionary)**: High-speed protection for fixed strings like VIP names, internal project codes (`Project Phoenix`), or sensitive organization names.
- **Tier 0.5 (Entropy)**: Shannon scoring for high-entropy strings, catching API keys and database connection strings.

### Tier 1-1.5: Rule-Based Logic
- **Tier 1 (Registry)**: 50+ national ID types, EMAILS, SSNs, IBANs, and Credit Cards.
- **Tier 1.1 (Phone Shield)**: Uses `libphonenumber` to validate digit sequences and reduce false positives.
- **Tier 1.2 (Address Shield)**: Heuristic street address parser for multiple languages (EN/FR/ES/DE).
- **Tier 1.5 (Greetings/Signatures)**: Contextual detection of names in salutations ("Best regards, Jean") and intro sentences.

### Tier 2: AI NER (Small Language Model)
- **Model**: OpenAI Privacy Filter (1.5B parameters).
- **Inference**: Local execution via the `slm-engine` (Python/FastAPI).
- **Optimization**: Specifically fine-tuned for high-risk domains like French Finance.
- **Schema**: Normalizes model outputs (e.g., `private_person`) to canonical OCULTAR labels (`PERSON`).

### Tier 3: Structural Heuristics
- **Proximity Expansion**: If the system detects `[TOKEN]` followed by `Dupont`, it merges them into a single `PERSON` entity to prevent leakage through partial tokenization.

---

## The "Truth" vs. The Plan

Based on the current codebase:

| Feature | Status | Implementation Detail |
|---------|--------|-----------------------|
| **Registry (Tier 1)** | ✅ Implemented | Defined in `internal/pii/registry.go`. Supports 50+ patterns. |
| **Checksums** | ✅ Implemented | Luhn (Credit Cards), Mod97 (IBAN) in `internal/pii/validators.go`. |
| **SLM Engine (Tier 2)** | ✅ Implemented | `apps/slm-engine` provides a FastAPI wrapper around Transformers. |
| **Base64 Evasion (0.1)** | 🏗️ Partial | Recursive scanning is planned in the roadmap but core decoding logic exists. |
| **Structural Heuristics (3)** | 🏗️ Partial | Simple proximity checks implemented in `engine.go` sort/merge logic. |

---

## How it works (The `Scan` function)

1.  **Iterate Registry**: For each `EntityDef` in the registry.
2.  **Pattern Match**: Run regex patterns.
3.  **Validate**: If a validator is present (e.g., `ValLuhn`), run the checksum.
4.  **Normalize**: Strip spaces/dashes before validation.
5.  **Hash**: Generate a deterministic hash of the matched value.
6.  **Redact**: Replace the match with a vault token.

```go
// From internal/pii/engine.go
if entity.Validator != ValNone {
    valFunc := GetValidator(entity.Validator)
    if valFunc != nil {
        valid = valFunc(valueForValidation)
    }
}
```

---

## Extending the Refinery

To add a new PII type:
1.  Add a new `EntityDef` to `internal/pii/registry.go`.
2.  If it needs a checksum, add a function to `internal/pii/validators.go`.
3.  Add a test case in `internal/pii/engine_test.go`.
