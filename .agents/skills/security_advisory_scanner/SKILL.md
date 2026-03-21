---
name: security-advisory-scanner
description: Supply-chain auditor for vulnerable dependencies (CVEs).
---

# Security Advisory Scanner (v1.1)

## Purpose

The SAS prevents "Dependency Decay". It ensures that no Ocultar component is released with critical vulnerabilities by cross-referencing the project's Manifest (SBOM) against global advisory databases (OSV, GitHub).

## Inputs / Outputs

### Inputs
- `sbom_path`: Path to `dist.manifest.yaml` or generated SBOM.
- `enforcement` (Enum): `ADVISORY` | `STRICT_HALT`.
- `min_severity`: `HIGH` | `CRITICAL`.

### Outputs
- `scan_verdict`: `CLEAN` | `VULNERABLE`.
- `vulnerability_list` (JSON): CVE IDs and affected components.

## Preconditions
- Network access for OSV/Snyk API (or local cached DB).

---

## Instructions

### 1. Manifest Parity Check
- Verify the `sbom_path` matches the current `go.mod` and `package.json` hashes.
- **Constraint**: If hashes differ, return `MANIFEST_DRIFT` error.

### 2. Vulnerability Lookup
- Query vulnerable components by version string.
- Identify "Reachability": Is the vulnerable function actually called in Ocultar's source?

### 3. Enforcement Gate
- If `enforcement` == `STRICT_HALT`:
    - **Logic**: Any `match.severity` >= `min_severity` results in `BLOCK_PIPELINE`.
- **Mitigation**: Suggest the specific `npm install` or `go get` command to patch.

## Failure Handling
- **`DB_UNREACHABLE`**: In `STRICT_HALT` mode, fail the build (Fail-Closed). In `ADVISORY` mode, log a warning.
- **`FALSE_POSITIVE`**: Allow whitelist only via signed `security_override.json`.

## Postconditions
- Results MUST be logged in the `continuous-ai-orchestrator` execution log.
