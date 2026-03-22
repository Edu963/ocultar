---
name: pii_registry_manager
description: "Expert AI orchestrator for maintaining the Ocultar PII Central Registry (`internal/pii/registry.go`). Validates new regex additions, ensures deterministic compliance, and enforces presence of checksum validators."
---

# Ocultar PII Registry Manager

## 🎯 Purpose
The `pii_registry_manager` is the **Architectural Guardian of the PII Detection Engine**. It enforces strict quality control over all Tier 0 (deterministic) entities added to the Ocultar pipeline.

As the central source of truth moved to `internal/pii/registry.go` in the EU Sovereign Detection Pack (v1), this skill guarantees that:
1. No regular expressions are defined outside of the registry.
2. All Tier 0 detections possess strict boundary enforcement (`\b`).
3. High-fidelity financial and health identifiers have active validation endpoints (e.g., Mod-97, Luhn).

## 🛠️ Triggers & Preconditions
Use this skill when:
- A new PII entity type is requested by a user or regulatory update.
- The `internal/pii/registry.go` file is modified.
- A false-positive/evasion gap is reported, requiring regex adjustments.

## 📋 Execution Protocol

### Step 1: Pre-flight Audit
1. Map the requested entity to its regulatory requirement (e.g., GDPR Art 4, HIPAA).
2. Query `internal/pii/registry.go` to ensure the entity does not already exist.

### Step 2: Quality & Determinism Engineering
When generating or modifying a regex pattern, you MUST enforce the following constraints:
- **Boundary Caps**: Start and end with `\\b` (or context-aware anchors) to prevent substring false-positives.
- **Capture Groups**: Define precisely what group holds the value using `CaptureGroup` if wrapping text is required.
- **Complexity Limits**: Ensure no exponential backtracking exists (validate using linear scaling regex principles).
- **Validation**: If the identifier contains a standard checksum digit (e.g., National IDs, Credit Cards), you MUST build a corresponding validation function in `internal/pii/validators.go` and link it to the registry.

### Step 3: Registration
Register the entity strictly using the `EntityDef` struct:
```go
{
    Type:          "ENTITY_NAME",
    Pattern:       regexp.MustCompile(`...`), // Raw string literals heavily preferred
    Validator:     ValFuncIdentifier, // or ValNone
    MinLength:     10,
    Normalization: true,
    CaptureGroup:  0,
}
```

### Step 4: Verification
1. Scaffold an evasion resistance test case in `internal/pii/engine_test.go` combining:
   - Valid formats (spaced, dashed).
   - Invalid checksum formats.
   - Non-boundary overlaps.
2. Execute `go test ./internal/pii -v` to prove determinism before committing.

## 🚨 Fail-Closed Constraints
- **Zero-Egress**: Under no circumstances should PII entity detection definitions log the sensitive content.
- If a validation function cannot be mathematically proven (no public checksum available), log an architectural risk notice and default the entity to Context-Dependent (Tier 1) rather than Deterministic (Tier 0).
