#!/bin/bash

# Verification Script for Ecosystem State Tracker (EST)
# This script simulates the new partitioned, dependency-aware state store logic.

STATE_ROOT=".agents/state"
SKILL_ID="test-skill"
TEST_FILE="test_dependency.txt"

# 1. Setup
mkdir -p "$STATE_ROOT/$SKILL_ID"
echo "Initial Content" > "$TEST_FILE"

# Function to compute state hash
compute_hash() {
    local input_params="$1"
    local dependency_file="$2"
    local input_hash=$(echo -n "$input_params" | sha256sum | cut -d' ' -f1)
    local dep_hash=$(sha256sum "$dependency_file" | cut -d' ' -f1)
    echo -n "$input_hash$dep_hash" | sha256sum | cut -d' ' -f1
}

# 2. Test Case: New Execution
echo "--- Test Case 1: New Execution ---"
INPUT_PARAMS='{"mode":"full"}'
HASH=$(compute_hash "$INPUT_PARAMS" "$TEST_FILE")
STATE_FILE="$STATE_ROOT/$SKILL_ID/$HASH.json"

if [ ! -f "$STATE_FILE" ]; then
    echo "Creating state for HASH: $HASH"
    cat <<EOF > "$STATE_FILE.tmp"
{
    "status": "SUCCESS",
    "state_hash": "$HASH",
    "expires_at": $(($(date +%s) + 3600)),
    "artifacts": ["/abs/path/to/result.txt"]
}
EOF
    mv "$STATE_FILE.tmp" "$STATE_FILE"
    echo "State registered successfully."
fi

# 3. Test Case: Redundancy Hit
echo "--- Test Case 2: Redundancy Hit ---"
NEW_HASH=$(compute_hash "$INPUT_PARAMS" "$TEST_FILE")
if [ -f "$STATE_ROOT/$SKILL_ID/$NEW_HASH.json" ]; then
    echo "Redundancy HIT for HASH: $NEW_HASH"
else
    echo "Redundancy MISS (Error: expected hit)"
fi

# 4. Test Case: Redundancy Miss (Dependency Change)
echo "--- Test Case 3: Dependency Change ---"
echo "Updated Content" >> "$TEST_FILE"
MIS_HASH=$(compute_hash "$INPUT_PARAMS" "$TEST_FILE")
if [ ! -f "$STATE_ROOT/$SKILL_ID/$MIS_HASH.json" ]; then
    echo "Redundancy MISS (Success: dependency change detected)"
else
    echo "Redundancy HIT (Error: expected miss)"
fi

# Cleanup
rm "$TEST_FILE"
rm -rf "$STATE_ROOT/$SKILL_ID"
echo "Verification complete."
