---
name: distribution-smoke-tester
description: A dedicated skill to spin up a clean Docker container, unpack the client-package, and verify the setup-community.sh runs successfully.
---

# Ocultar | Distribution Smoke Tester

## Purpose
This skill provides the final validation before a release. It ensures that the distributed package is self-contained and functional by simulating a "First Install" in a clean environment.

## When To Use This Skill
**MANDATORY** before any release is finalized.
- Must run **AFTER** `release-artifact-builder` has created the archive.

## Preconditions
- `docker` and `docker-compose` are available in the testing environment.
- The `dist/ocultar-community.zip` package has been built.

## Instructions

### 1. Clean Room Setup
1. Create a temporary directory `temp_smoke_test/`.
2. Unpack the distribution archive into this directory.
3. Ensure no local `.env` or persistent volumes are inherited.

### 2. Execution Flow
1. Run the setup script: `cd temp_smoke_test && bash scripts/setup-community.sh`.
2. Verify that `docker-compose.yml` is correctly initialized.
3. Start the services: `docker-compose up -d`.

### 3. Health & Redaction Check
1. Wait for services to be healthy (using `curl http://localhost:8081/healthz`).
2. Run a basic redaction test using the existing `smoke_test.sh` logic.
3. Confirm that the dashboard (`http://localhost:3000`) is reachable.

### 4. Cleanup
1. Shut down Docker services: `docker-compose down -v`.
2. Delete the temporary directory.

## Outputs
- **Smoke Test Report**: Success or failure log with system diagnostics.
- **Traceability**: Link the test run to the specific git commit hash.

## Failure Handling
- **Missing Dependency in Archive**: If `setup-community.sh` fails due to a missing file, YOU MUST return to the `client-package-updater` skill and fix the manifest.
- **Startup Timeout**: If the proxy fails to start within 60 seconds, check the container logs for configuration errors or missing environment defaults.
