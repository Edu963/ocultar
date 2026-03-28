#!/bin/bash

# OCULTAR Master Verification Script
# Senior Verification Engineer & Red Team Specialist

set -u

# Configuration
MOCK_LLM_PORT=9999
PROXY_PORT=8080
VAULT_PATH="security/test_vault.db"
MOCK_LOG="/tmp/mock_received.log"
RES_DIR="reports/verification"
mkdir -p "$RES_DIR"

# Clean up any lingering processes
cleanup() {
    echo "[INFO] Cleaning up..."
    pkill -9 -f "mock_llm.py" || true
    pkill -9 -f "ocultar-proxy" || true
    rm -f "$MOCK_LOG"
    rm -f "security/test_vault.db" # Remove vault to avoid locks and ensure fresh start
    sleep 1
}
trap cleanup EXIT

# 1. Start Mock LLM
python3 security/mock_llm.py $MOCK_LLM_PORT > /dev/null 2>&1 &
echo "[INFO] Mock LLM started on $MOCK_LLM_PORT"
sleep 1

# 2. Function to start Proxy
start_proxy() {
    local dev_flag=$1
    export OCU_MASTER_KEY="1111111111111111111111111111111111111111111111111111111111111111"
    export OCU_SALT="0000000000000000"
    export OCU_LICENSE_KEY="HSKTMfEnIYdTRmyOohu8DcQh8PRGUIhvTlaOFV+dc+LLm04kAok5RmgpzGldoqwTuiwdy3wi4JlftcOLn+4ICQ==.eyJDdXN0b21lck5hbWUiOiJEZXZlbG9wbWVudCIsIlRpZXIiOiJlbnRlcnByaXNlIiwiRXhwaXJ5RGF0ZSI6MTgwNTY2NzkyNSwiQ2FwYWJpbGl0aWVzIjozfQ=="
    export OCU_PROXY_TARGET="http://127.0.0.1:$MOCK_LLM_PORT"
    export OCU_PROXY_PORT=$PROXY_PORT
    export OCU_VAULT_PATH=$VAULT_PATH
    
    ./bin/ocultar-proxy $dev_flag > "$RES_DIR/proxy.log" 2>&1 &
    # Wait for proxy to be ready
    for i in {1..10}; do
        if curl -s http://localhost:$PROXY_PORT/healthz > /dev/null; then
            echo "[INFO] Proxy ready"
            return 0
        fi
        sleep 0.5
    done
    echo "[ERROR] Proxy failed to start"
    exit 1
}

run_test() {
    local name=$1
    local input_file=$2
    local extra_args=$3
    echo "[TEST] Running: $name"
    
    local res
    res=$(curl -s -X POST -H "Content-Type: application/json" $extra_args \
        -d @"$input_file" http://localhost:$PROXY_PORT/ 2>&1)
    
    echo "$res" > "$RES_DIR/$name.json"
    
    # Check mock log for leaks (specifically email and address which Tier 1 should catch)
    if grep -iE "john\.doe@example\.com|123 Maple St" "$MOCK_LOG" > /dev/null; then
        echo "[FAIL] $name: PII LEAKED to Upstream!"
        return 1
    else
        echo "[PASS] $name: Zero-Egress verified for Tier 1 PII"
    fi
    
    # Check response for tokens
    if grep -q "\[[A-Z_]*_[0-9a-f]\{8\}\]" "$RES_DIR/$name.json"; then
        echo "[INFO] $name: Tokens found in response (expected if re-hydration is off or simulated)"
    fi
}

# --- STAGE 1: Functional Testing ---
echo "--- STAGE 1: Functional Testing ---"
start_proxy "--dev"
run_test "Functional_Standard" "security/test_data/functional_pii.json" ""
sleep 1

# --- STAGE 2: Adversarial Testing ---
echo "--- STAGE 2: Adversarial Testing ---"
run_test "Adversarial_Obfuscation" "security/test_data/adversarial_pii.json" ""
sleep 1
cleanup
python3 security/mock_llm.py $MOCK_LLM_PORT > /dev/null 2>&1 & # Restart mock

# --- STAGE 3: SSRF Protection ---
echo "--- STAGE 3: SSRF Protection ---"
start_proxy "--dev"
echo "[TEST] SSRF Protection"
# Attempt to reach localhost:22 via Ocultar-Target
# Use -w to check for 403 status code
ssrf_res=$(curl -s -o "$RES_DIR/ssrf_local.json" -w "%{http_code}" -X POST -H "Content-Type: application/json" \
     -H "Ocultar-Target: http://127.0.0.1:22" \
     -d '{"msg":"hello"}' http://localhost:$PROXY_PORT/)

if [ "$ssrf_res" == "403" ] || grep -q "SSRF" "$RES_DIR/proxy.log"; then
    echo "[PASS] SSRF Protection: Blocked local access (Status: $ssrf_res)"
else
    echo "[FAIL] SSRF Protection: Did not explicitly block local access (Status: $ssrf_res)"
fi

# --- STAGE 4: Load Testing ---
echo "--- STAGE 4: Load Testing ---"
echo "[INFO] Flooding with safe requests (Concurrency 120)..."
ab -n 1000 -c 120 -p security/test_data/functional_pii.json -T application/json http://localhost:$PROXY_PORT/ > "$RES_DIR/load_test.txt" 2>&1
# Check if proxy log contains 429
if grep -q "429" "$RES_DIR/proxy.log" || grep -q "Too Many Requests" "$RES_DIR/load_test.txt"; then
    echo "[PASS] Load Safety: 429 detected under pressure"
else
    echo "[FAIL] Load Safety: No 429 detected (Queue size might be too large or requests too fast)"
fi

# --- STAGE 5: Failure Mode (Fail-Closed) ---
echo "--- STAGE 5: Failure Mode ---"
cleanup
echo "[TEST] Fail-Closed: Missing Master Key"
unset OCU_MASTER_KEY
./bin/ocultar-proxy > "$RES_DIR/fail_closed_key.log" 2>&1 &
sleep 2
if ! curl -s http://localhost:$PROXY_PORT/healthz > /dev/null; then
    echo "[PASS] Fail-Closed: Proxy refused to start without key"
else
    echo "[FAIL] Fail-Closed: Proxy started without master key!"
fi

echo "[INFO] All tests completed. Review $RES_DIR for details."
