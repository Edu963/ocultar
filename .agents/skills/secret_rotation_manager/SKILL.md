---
name: secret-rotation-manager
description: Remediation expert for Ocultar infrastructure secrets. Automatically rotates cryptographically secure keys and salts to eliminate security drift.
---

# Secret Rotation Manager (v1.0)

## Purpose

This skill is the primary remediator for "Secret Drift". It replaces insecure defaults (`OCU_MASTER_KEY`, `OCU_SALT`) with production-grade, cryptographically secure values. It ensures data sovereignty by rotating the root of trust whenever a compromise is suspected or a policy threshold (90 days) is reached.

## Preconditions

- **Tools**: `openssl`, `sed`, `docker-compose`.
- **Access**: Write access to `docker-compose.proxy.yml` and root `.env` files.
- **Service Impact**: Aware that rotating the `OCU_MASTER_KEY` may render previously encrypted/vaulted data unreadable unless migration occurs.

## Inputs / Outputs

### Inputs
- `target_files` (List): Paths to YAML or ENV files containing secrets (default: `["docker-compose.proxy.yml", ".env"]`).
- `rotation_reason` (String): Reason for rotation (e.g., `DRIFT_DETECTED`, `POLICY_ROTATION`).
- `force` (Boolean): If true, skip backup steps (NOT RECOMMENDED).

### Outputs
- `rotation_log` (Artifact): Audit trail of changed files and new key checksums (SHA-256).
- `restart_required` (Boolean): Set to `true` if services need a restart to pick up changes.

---

## Instructions

### Step 1 – Entropy Generation
1.  Generate a new `OCU_MASTER_KEY`:
    - Requirement: 32 bytes, hex-encoded.
    - Command: `openssl rand -hex 32`.
2.  Generate a new `OCU_SALT`:
    - Requirement: 16 bytes, fixed-length salt.
    - Command: `openssl rand -hex 8` (16 chars).

### Step 2 – Backup & Staging
1.  Before any write, create a timestamped backup of all `target_files`.
    - Format: `filename.bak.YYYYMMDD_HHMMSS`.

### Step 3 – Atomic Update
1.  Use `sed` to replace the insecure defaults with the new entropy.
    - Match `OCU_MASTER_KEY=${OCU_MASTER_KEY:-.*}` or literal hex values.
    - Match `OCU_SALT=${OCU_SALT:-.*}` or literal hex values.
2.  **Safety Gate**: Verify the new file is still valid YAML/ENV before overwriting.

### Step 4 – Service Notification
1.  If the change was in a docker-compose file:
    - Instruct the orchestrator to run `docker compose up -d`.
2.  Log the SHA-256 checksum of the new keys (NEVER log the keys themselves).

---

## Failure Handling

- **Migration Conflict**: If the `VAULT` contains data, warn that rotation will break access. Request a `Data Migration` plan before proceeding.
- **Permission Denied**: If files are read-only, log failure and notify the `CISO & Compliance Officer`.

---

## Ecosystem Role
- **Category**: Remediator / Tasker.
- **Dependencies**: `compliance-integrity-suite` (Trigger), `artifact-signer` (Verification).
- **Triggers**: Infrastructure update.
