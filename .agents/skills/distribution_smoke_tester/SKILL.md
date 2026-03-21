---
name: distribution-integrity-validator
description: Production-grade validator for Ocultar distribution artifacts. Executes cryptographic signature verification, clean-room environment provisioning, and automated health/redaction audits.
---

# Ocultar | Distribution Integrity Validator

## Purpose
This skill serves as the final, immutable gate in the Ocultar distribution pipeline. It ensures that every release artifact is cryptographically authentic, structurally complete, and functionally secure in a zero-knowledge, clean-room environment.

## When To Use This Skill
**MANDATORY GATING** for all release candidates (`rc-*`) and production releases.
- Must execute after `artifact-signer` has generated `.sig` files.
- Must execute prior to any public distribution or deployment.

## Preconditions
- `docker` (v20.10+) and `docker-compose` (v2.0+) available on the runner.
- The **Ocultar Release Public Key** is accessible for signature verification.
- Artifacts (`.zip`, `.tar.gz`) and their corresponding `.sig` files exist in the `dist/` directory.

## Inputs / Outputs

### Inputs:
- `artifact_path` (Path): Absolute path to the distribution package.
- `distribution_type` (Enum): `community` | `enterprise`.
- `verify_signature` (Boolean): Default `true`.

### Outputs:
- `validation_report` (Artifact): Detailed log of signature, setup, and health checks.
- `integrity_status` (Boolean): `true` if all checks pass.
- `diagnostics_bundle` (Path): (On Failure) Archive of container logs and environment state.

## Instructions

### 1. Cryptographic Verification
1. Locate the signature file: `{artifact_path}.sig`.
2. Verify the artifact integrity using the Ocultar Public Key.
3. **FAIL-CLOSED**: If signature verification fails or the `.sig` file is missing, abort immediately. DO NOT execute scripts from an unverified archive.

### 2. Isolated Environment Provisioning
1. Create a randomized, isolated workspace: `/tmp/ocultar_val_<uuid>/`.
2. Unpack the artifact into the workspace.
3. **Network Isolation**: Ensure the test environment uses a dedicated Docker network.
4. **Dynamic Port Mapping**: Do not use hardcoded host ports. Use `0` (auto-assign) or a verified available range to avoid host collisions.

### 3. Execution & Health Audit
1. Run the targeted setup script (e.g., `bash scripts/setup-community.sh`).
2. Verify the generation of `docker-compose.yml` and `.env` (ensure defaults are production-safe).
3. Lifecycle Start: `docker-compose up -d --wait`.
4. **Health Gate**: Query the internal health endpoint (e.g., `/healthz`) until status is `200 OK`. Timeout after 120s.

### 4. Security & Redaction Smoke Test
1. Execute the internal `smoke_test.sh` logic (or equivalent test runner).
2. **Mandatory Check**: Send a PII-heavy payload and verify that the core engine intercepts and redacts/pseudonymizes as per the global policy.
3. **Web Dashboard**: Verify the Enterprise Dashboard is reachable and serving valid localized assets.

### 5. Teardown & Reporting
1. Extract all container logs: `docker-compose logs > validation.log`.
2. Collect system stats: `docker stats --no-stream`.
3. Perform cleanup: `docker-compose down -v` and wipe the temporary workspace.

## Failure Handling
- **Signature Mismatch**: Exit with `SECURITY_VIOLATION`. Report to CTO/Security Lead.
- **Dependency Missing**: Exit with `MANIFEST_MISMATCH`. Return to `client-package-updater`.
- **Startup Timeout**: Inspect `validation.log` for service crashes or database connectivity issues.

## Quality Standards
- **Determinism**: The test MUST produce the same result for the same artifact regardless of host state.
- **Idempotency**: Running the validator multiple times must not leave residue on the host.
