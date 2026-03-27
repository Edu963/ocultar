#!/bin/bash
# OCULTAR Architectural Linter (Functional Stub)
# Verifies package structure and forbidden imports.

echo "[*] Verifying architectural integrity..."

# Rule: No direct imports of 'internal/pii' from 'pkg/gateway'
if grep -r "github.com/Edu963/ocultar/internal/pii" pkg/gateway; then
    echo "[!] VIOLATION: Gateway must not depend on internal PII logic directly. Use the Refinery engine."
    exit 1
fi

echo "[+] Architectural linting passed."
exit 0
