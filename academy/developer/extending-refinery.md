# 🧪 Extending the Refinery

There are two ways to extend OCULTAR's detection capabilities: adding simple rules (No-Code) or adding complex validation (Code).

## 1. Adding a Detection Rule (No-Code)
*Target: Fast iteration — no recompilation required.*

Add rules via `configs/config.yaml`. They take effect at the next startup.

```yaml
# Custom regex rule
regexes:
  - type: INTERNAL_ID
    pattern: '\bINT-[0-9]{4}-[A-Z]{2}\b'

# Custom dictionary rule
dictionaries:
  - type: SECRET_PROJECT
    terms:
      - "Project Nightshade"
      - "Operation Dusk"
```

---

## 2. Adding a Detection Rule (In Code)
*Target: Core contributors adding a permanent rule to the registry.*

To add a permanent rule to the OCULTAR registry:

### Step 1: Update the Registry
Open `internal/pii/registry.go` and add a new `EntityDef` to the `Registry` slice.

```go
{
    Type: "PASSPORT_NUMBER",
    Pattern: regexp.MustCompile(`\b[A-Z]{1,2}[0-9]{6,9}\b`),
    Validator: ValNone,
    MinLength: 7,
    Normalization: true,
}
```

### Step 2: Add a Checksum Validator (Optional)
If the PII type has a checksum (like Luhn for Credit Cards or Mod97 for IBAN), add a validator.

1.  Add a new constant to `ValidationMethod` in `registry.go`.
2.  Implement the validation logic in `internal/pii/validators.go`.
3.  Register it in the `GetValidator` function.

```go
// internal/pii/validators.go
func validateMyID(id string) bool {
    // your logic here
    return true
}
```

### Step 3: Test Your Rule
Add a test case to `internal/pii/engine_test.go` to verify matches and non-matches.

---

## 3. Adding a New Detection Tier
If you need to add a completely new *way* of detecting PII (e.g., a new AI engine or a complex parser):

1.  Create a parser function in `services/refinery/pkg/refinery/`.
2.  In `refinery.go`, update the `RefineString` method to include your new tier.
3.  Ensure you handle overlaps! The system uses a sorting and merging strategy to prevent tokenizing parts of other tokens.

---

## 🔐 The "Fail-Closed" Rule
When extending the refinery, remember: **If your code errors, the refinery must fail.**

```go
refined, err = e.myNewTier(refined)
if err != nil {
    return "", fmt.Errorf("my new tier failed: %w", err) // This will block the request
}
```
Never return a partially cleaned string on error.
