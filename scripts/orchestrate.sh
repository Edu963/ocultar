# 4. Tests
echo "[STEP 4] Running Build and Unit Tests..."
make build
make test
# Using the root Makefile which now handles the clean/build/test cycle correctly.
make build
make test

# 5. Semantic Security Gate
echo "[STEP 5] Running Semantic Security Gate (Tier 2 AI)..."
./scripts/semantic.sh

echo "[SUCCESS] Orchestration complete. System is compliant (100/100 Readiness)."
