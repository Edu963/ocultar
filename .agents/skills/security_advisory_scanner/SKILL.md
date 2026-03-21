---
name: security-advisory-scanner
description: Integrate with OSV, Snyk, or GitHub Advisory Database to detect vulnerable dependencies before release. Ensures all Ocultar components (Go, JS, Python) are free from known CVEs.
---

# Security Advisory Scanner

## Purpose
This skill proactively scans the Software Bill of Materials (SBOM) and project manifest files (go.mod, package.json, requirements.txt) for known vulnerabilities. It ensures that no production build contains critical security flaws in its supply chain.

## Inputs
- `sbom_path`: Absolute path to the generated SBOM file.
- `severity_threshold`: (Enum) `LOW`, `MEDIUM`, `HIGH`, `CRITICAL`.
- `fail_on_vulnerability`: Boolean (Trigger HALT if a match is found).

## Outputs
- `vulnerability_report`: List of detected CVEs, affected versions, and fix version suggestions.
- `scanner_status`: `SAFE` or `VULNERABLE`.

---

## Instructions

### 1. Extract Dependencies
- Parse `go.mod` (Core Engine/Sombra), `package.json` (Dashboard), and any Python `requirements.txt`.
- Map all indirect and direct dependencies.

### 2. Query Advisory Databases
- Use `osv-scanner` or equivalent API tools to query the **Open Source Vulnerability (OSV)** database.
- Cross-reference with the **GitHub Advisory Database**.

### 3. Evaluate Impact
- Determine if the vulnerability is reachable in the Ocultar execution context (e.g., if a crypto bug affects our specific AES implementation).
- Categorize by `severity_threshold`.

### 4. Enforcement (Gate)
- If `fail_on_vulnerability` is TRUE and `CRITICAL` or `HIGH` vulnerabilities are found:
  - **HALT** the `continuous-ai-orchestrator`.
  - Report the specific CVE and the required update.

## Failure Handling
- **Network Error**: If advisory databases are unreachable, report a "Compliance Gap" but do not block staging builds (unless in STRICT mode).
- **Scanner Failure**: If the scanner tool crashes, assume `VULNERABLE` state for SAFETY.

## Examples

### Scenario: Vulnerable Go Module
- **Input**: `golang.org/x/crypto` version with a known padding oracle bug.
- **Process**: OSV lookup returns CVE-202X-XXXX.
- **Result**: Scanner flags the build and suggests `go get golang.org/x/crypto@latest`.
