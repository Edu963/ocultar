---
name: secret-scanner
description: Proactively scans the repository and staging area for sensitive information.
---

# Secret Scanner (Pre-commit) (v1.1)

## Purpose

The SS is the primary gateway for preventing "Secret Sprawl". It ensures that credentials never touch the code or config files unless they are in an Ocultar-sanctioned Identity Vault. This skill is implemented by the `./tools/scripts/run_secret_scanner.sh` functional gate.

## Inputs / Outputs

### Inputs
- `scan_scope` (Git Diff | Path).
- `approved_storage`: List of paths allowed to contain secrets. (Default: `["services/refinery/configs/secrets.json"]`).

### Outputs
- `scanner_verdict` (Enum): `SAFE` | `BLOCK_COMMIT`.
- `findings`: Artifact mapping files to detected entropy/keys.

## Preconditions
- Access to high-entropy detection patterns (regex).

---

## Instructions

### 1. Entropy & Pattern Matching
- Scan all text files in `scan_scope` for:
    - RSA/ECDSA Private Keys.
    - Cloud API Keys (Ocular, AWS, OpenAI).
    - Database URIs with embedded passwords.

### 2. Path Isolation Enforcement
- If a match is found:
    - **Check**: Is the file path in `approved_storage`?
    - **Action**: If YES, ignore. If NO, flag as `BLOCK_COMMIT`.

### 3. Verification Gate (Pre-commit)
- If `scanner_verdict` == `BLOCK_COMMIT`:
    - **Instruction**: "Secret detected in `{{file}}`. Secrets MUST be moved to `services/refinery/configs/secrets.json` or the Ocultar Identity Vault."

## Failure Handling
- **`AMBIGUOUS_STRING`**: If a string looks like a secret but is a test asset, instruct the developer to add it to `.secretignore`.

## Postconditions
- `scanner_verdict` MUST be checked by the `ai-development-protocol`.
