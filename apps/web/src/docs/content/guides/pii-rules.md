# PII Detection Rules

OCULTAR uses a combination of Regex patterns and Dictionary matching to identify PII. This guide explains how to configure and extend these rules.

## Regex Rules
Regex rules are best for structured data like SSNs, Credit Cards, and Email addresses.

### Configuration
In `config.yaml`, add rules under `settings.regexes`.

```yaml
settings:
  regexes:
    - type: "GDPR_ID"
      pattern: "[A-Z]{2}[0-9]{8}"
    - type: "INTERNAL_SECRET"
      pattern: "OC-[a-f0-9]{32}"
```

### Validator Functions
Some patterns (like Credit Cards) require checksum validation (e.g., Luhn algorithm). OCULTAR has built-in validators that you can enable by setting the `type` to one of the following:

- `CREDIT_CARD`: Automatic Luhn-10 verification.
- `SSN`: Standard US Social Security Number format check.
- `IBAN`: International Bank Account Number validation.

---

## Dictionary Rules
Dictionary rules are for specific sensitive terms, names, or project titles.

### Global vs Local
- **Global Dictionary**: Defined in `config.yaml` or via the `/api/refine` payload.
- **Local Dictionary**: Loaded from `data/protected_entities.json` (Enterprise only).

### Example
```json
[
  "Héctor Eduardo",
  "Ouroboros"
]
```

## Rule Priority
OCULTAR processes rules in the following order:
1. **Tier 0 (Dictionary)**: Fastest, deterministic.
2. **Tier 1 (Regex)**: Heuristic pattern matching.
3. **Tier 2 (AI)**: Contextual scan (if enabled).

---

## Best Practices
- **Be Specific**: Avoid overly broad regex patterns (like `.*`) which can lead to false positives.
- **Anchor Your Regex**: Use `\b` (word boundaries) to ensure you aren't matching substrings within larger safe words.
- **Testing**: Use the `tools/diagnostic_suite` to verify rules against your test datasets before deploying to production.
