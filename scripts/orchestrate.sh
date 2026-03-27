#!/bin/bash
# OCULTAR Orchestration Script (Skeleton v2.1.0)
# This script enforces the 16-Step Ocultar Protocol via automated security gates.

set -e

echo "[INFO] Starting OCULTAR Orchestration Pipeline..."

# 1. Secret Scanner
echo "[STEP 1] Running Secret Scanner..."
./tools/scripts/run_secret_scanner.sh

# 2. Architectural Linter
echo "[STEP 2] Running Architectural Linter..."
./tools/scripts/run_arch_linter.sh

# 3. Zero-Egress Validator
echo "[STEP 3] Running Zero-Egress Validator..."
./tools/scripts/run_zero_egress_validator.sh

# 4. Tests
echo "[STEP 4] Running Build and Unit Tests..."
# Using the root Makefile which now handles the clean/build/test cycle correctly.
make build
make test

echo "[SUCCESS] Orchestration complete. System is compliant."
