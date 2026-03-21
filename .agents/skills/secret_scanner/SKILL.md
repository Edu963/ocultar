---
name: secret-scanner
description: Proactively scans the repository and staging area for sensitive information such as API keys, secrets, private credentials, and internal paths. Intended for use during development and pre-commit checks.
---

# Secret Scanner (Pre-commit)

## Purpose
The Secret Scanner is the first line of defense in the Ocultar Zero-Trust model. Its primary goal is to prevent sensitive information from ever being committed to the version control system.

## When To Use This Skill
- During active development when new code is written.
- As a mandatory pre-commit hook or PR check.
- When cleaning up a legacy codebase.

## Instructions

### 1. Identify Sensitive Patterns
- Scan for high-entropy strings, known secret headers (e.g., `-----BEGIN RSA PRIVATE KEY-----`), and common API key formats (AWS, OpenAI, Ocultar).
- Detect hardcoded credentials in config files, scripts, and source code.

### 2. Flag Internal Paths
- Identify hardcoded absolute paths that reveal internal server structures or developer-specific directories (e.g., `/home/edu/...`).

### 3. Verification Gate
- If a secret is detected:
  - **HALT**: Do not proceed with the commit or pull request.
  - **REMEDY**: Instruct the developer to use environment variables or the Ocultar Identity Vault.

### 4. False Positive Management
- Maintain a local `.secretignore` for known-safe test keys or public headers.

## Examples

### Scenario: Staging an API key
- **Input**: `apiKey := "sk-1234567890abcdef"` in `main.go`.
- **Action**: Flag the line and block the check-in.
- **Suggestion**: "Move this key to an environment variable or `services/engine/configs/secrets.json`."
