---
name: zero-egress-validator
description: Specialized security linter to detect unmasked PII leakage and unauthorized external network calls.
---

# Zero-Egress Validator (v1.1)

## Purpose

Enforces the "Platinum Rule": **No unmasked PII ever leaves the client's VPC.** This skill acts as a static gate to block code that attempts to bypass the Sombra logic. This skill is implemented by the `./tools/scripts/run_zero_egress_validator.sh` functional gate.

## Inputs / Outputs

### Inputs
- `change_set` (Git Diff): The code diff to evaluate.
- `strict_mode` (Boolean): Default `TRUE`. If false, only log warnings for suspicious patterns.

### Outputs
- `verdict` (Enum): `PASS` | `FAIL` | `MANDATORY_REDACTION`.
- `leak_report` (Artifact): File/Line mapping of egress risks.

## Preconditions
- Access to the list of "Approved Internal Domains" in `sombra.yaml`.

---

## Instructions

### 1. Heuristic Egress Search
Identify outbound communication that bypasses Ocultar hooks:
- **Heuristic A (JS/TS)**: `fetch(`, `axios.`, `navigator.sendBeacon`.
- **Heuristic B (Go)**: `http.Client`, `net.Dial`, `grpc.Dial`.
- **Heuristic C (Python)**: `requests.`, `urllib.`, `httpx.`.

### 2. Payload Sanitization Audit
For every hit in Step 1, verify variable names against the PII category list.
- **Violation**: If variables like `user_ssn`, `raw_prompt`, or `unmasked_payload` are passed to external IPs.
- **Logic**: If `uri` is NOT in the "Approved Internal Domains", the data MUST be piped through `Refinery.Refine()` first.

### 3. Verdict Generation
- `FAIL`: Unmasked PII sent to external domain.
- `MANDATORY_REDACTION`: Egress found, but data sensitivity is unknown (Assume PII).
- `PASS`: No egress or egress is piped through Ocultar.

## Failure Handling
- **`OBFUSCATION_DETECTED`**: If `eval()` or `base64.decode` is used near an egress point, trigger an automatic `FAIL` and escort to `red-team-evasion-scanner`.

## Postconditions
- `FAIL` verdict MUST block the current task in `continuous-ai-orchestrator`.
