---
name: sombra-gateway-policy-enforcer
description: Automatically injects policy hooks when new endpoints or features are added. Ensures all AI requests go through the Fail-Closed logic.
---

# Sombra Gateway Policy Enforcer (v1.1)

## Purpose

The SGPE acts as the "Architectural Linter" for the Gateway layer. It ensures that every code path that handles AI payloads (Prompt or Response) is explicitly bound to a Refinery hook and the **Hardened Concurrency Queue**, preventing accidental unmasked egress or service instability.

## Inputs / Outputs

### Inputs
- `change_set` (Git Diff): The code modifications to evaluate.
- `registry_source`: Path to `sombra.yaml` or route registry.

### Outputs
- `policy_verdict` (Enum): `ENFORCED` | `BYPASS_DETECTED`.
- `critical_gaps` (List): Description of paths missing security hooks.

## Preconditions
- Access to `apps/sombra/pkg/handler/` source code.

---

## Instructions

### 1. Data Path Identification
- Audit the `change_set` for new HTTP routes or model connector implementations.
- **Goal**: Identify any function that handles `[]byte` or `string` payloads destined for an LLM.

### 2. Hook Validation
- For every path identified:
    - **Requirement**: Must call `Refinery.Refine()` and be wrapped in the `ConcurrencySemaphore` or `HardenedQueue`.
    - **Routing**: Ensure **Multi-model router** selections are logged in the audit trail.
    - **Policy**: If the hook is missing or commented out, flag as `BYPASS_DETECTED`.

### 3. Fail-Closed Verification
- Verify that the error return from the Refinery is NOT ignored.
- **Logic**: Any `if err != nil` following a refine call MUST result in a terminal request block (e.g., `return nil, err`).

## Failure Handling
- **`COMPLEX_BYPASS`**: If custom encryption or encoding is used before refinery, flag for `red-team-evasion-scanner` audit.

## Postconditions
- `BYPASS_DETECTED` MUST result in a mandatory human review and block the CI pipeline.
