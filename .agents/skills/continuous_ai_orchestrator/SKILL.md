---
name: continuous-ai-orchestrator
description: Metadata-driven master orchestrator for the Ocultar AI ecosystem. Coordinates specialized skills and protocols through impact-aware execution profiles and deterministic validation gates.
---

# Continuous AI Orchestrator (v2.0)

## Purpose

The Continuous AI Orchestrator is the "Ecosystem Brain" responsible for autonomous lifecycle management. Instead of running a static sequence, it analyzes the **Impact Matrix** of a change and dynamically dispatches specialized skills and sub-protocols. It ensures that every modification is validated for security, documentation parity, and architectural integrity.

## Inputs / Outputs

### Inputs
- `change_set` (Git Diff): The raw source of change.
- `trigger_event` (Enum): `COMMIT`, `PR_OPEN`, `MANUAL_RELEASE`, `PILOT_UPDATE`.
- `impact_matrix` (From `change-impact-visualizer`): Component-level impact mapping.

### Outputs
- `orchestration_summary` (Artifact): A report of triggered skills, their outcomes, and validation status.
- `gate_status` (Boolean): `PASS` or `FAIL` (Blocks downstream release if FAIL).

---

## Instructions

### 1. Change Sensitivity Analysis
**Dependency**: `change-impact-visualizer`
- Immediately execute the `change-impact-visualizer` on the `change_set`.
- Parse the resulting **Impact Matrix** to identify affected functional domains.

### 2. Profile Selection (Trigger Logic)
Map the `impact_matrix` to one or more **Execution Profiles**:

| Impact Domain | Execution Profile | Required Skills / Protocols |
| :--- | :--- | :--- |
| **Core Engine / Sombra** | `CORE_SECURITY` | `ai-development-protocol`, `secret-scanner`, `security-advisory-scanner`, `sombra-gateway-policy-enforcer`, `red-team-evasion-scanner` |
| **PII / Policy Lifecycle** | `PRIVACY_COMPLIANCE` | `regulation-digest-ingestor`, `regulatory-intent-decoder`, `policy-schema-generator`, `policy-impact-simulator`, `compliance-certificate-signer`, `compliance-integrity-suite`, `pii-regression-suite-runner` |
| **Documentation Only** | `DOCS_SYNC` | `compliance-docs-orchestrator` |
| **Enterprise Dashboard**| `UI_INTEGRITY` | `enterprise-dashboard-integrity-checker`, `secret-scanner` |
| **Pilot / Onboarding** | `PILOT_OPS` | `Ocultar | Pilot Operations Manager`, `industry-snapshot-generator` |

### 3. Orchestrated Execution Flow
Follow the DAG (Directed Acyclic Graph) for the active profiles:

#### Phase A: Initialization (Sequential)
1. **Ecosystem State Tracker**: Check for redundant tasks and initialize the execution audit trail.
2. **Secret Scanner**: Scan for hardcoded keys and internal paths.

#### Phase B: Parallel Processing
Execute based on selected profiles:
- **Security**: Run `red-team-evasion-scanner`, `sombra-gateway-policy-enforcer`, and `security-advisory-scanner`.
- **Compliance (The Policy DAG)**:
    1. `regulation-digest-ingestor` -> `regulatory-intent-decoder`.
    2. `policy-schema-generator` (v2.1) consumes the manifest.
    3. `policy-impact-simulator` validates against historical logs.
    4. `compliance-certificate-signer` signs the final JSON.
    5. `compliance-integrity-suite` performs the final audit.
- **Documentation**: Run `compliance-docs-orchestrator` to sync all technical and regulatory docs.

#### Phase C: Validation & Content Redaction (Sequential)
1. **Quality Gates**: Verify all parallel tasks returned `SUCCESS`.
2. **Content Redactor**: Sanitize the distribution area for internal metadata and temporary files.
3. **Release Builder**: If `trigger_event` == `MANUAL_RELEASE` or `PR_MERGE` to `main`:
   - Run `release-artifact-builder`.
   - Run `artifact-signer`.
   - Run `sbom-generator`.

### 4. Post-Flight Integrity Check
- Update the `ecosystem-state-tracker` with final hashes and outcomes.
- Ensure `orchestration_summary` is generated and signed.

---

## Failure Handling

- **Critical Failure**: Any "Security" (Scanner, Evasion) or "Core" skill failure triggers an immediate **LOCKDOWN** state.
- **Compliance Gap**: If `pii-regression-suite-runner` fails, the ruleset must be rolled back.

## Examples

### Example: Core Engine Optimization
- **Trigger**: Change in `pkg/engine/refinery.go`.
- **Action**:
  1. `change-impact-visualizer` identifies `CORE_SECURITY`.
  2. `ecosystem-state-tracker` initializes.
  3. `secret-scanner` clears the source.
  4. `ai-development-protocol` & `security-advisory-scanner` run in parallel.
  5. `content-redactor` scrubs internal build paths.
  6. `orchestration_summary` generated.
