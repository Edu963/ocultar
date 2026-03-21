---
name: manage_ocultar_license
description: Professional-grade orchestrator for the OCULTAR Ed25519 offline license ecosystem. Manages keypair generation, license issuance, and validation integrity.
---

# Manage OCULTAR License (v1.1)

## Purpose

The MLC manages the "Offline Root-of-Trust". It handles the generation, signing, and integration of Ed25519 license tokens for Enterprise deployments.

## Inputs / Outputs

### Inputs
- `customer_name`: Legal entity name.
- `tier` (Enum): `COMMUNITY` | `ENTERPRISE` | `PRO_PILOT`.
- `capabilities` (Bitmask): e.g., `1` (Slack), `2` (SharePoint).
- `rotation_required` (Boolean): Default `FALSE`.

### Outputs
- `license_token`: The full `Signature.Payload` string.
- `public_key_patch`: Go code snippet for `license.go`.
- `verdict`: `ISSUED` | `FAILED`.

## Preconditions
- Access to `tools/scripts/scripts/keygen.go`.
- No plaintext keys stored in environment history.

---

## Instructions

### 1. Keypair Selection
- If `rotation_required` == `TRUE`: Generate new pair via `keygen.go`.
- Else: Retrieve `OCU_MASTER_PRIVATE_KEY` from the Secure Vault.

### 2. Token Issuance
- Execute the signing command:
  `go run tools/scripts/scripts/keygen.go --customer="{{customer_name}}" --tier="{{tier}}" --caps={{capabilities}}`
- **Integrity Check**: Validate the generated token immediately via `license_validation_cli` before deployment.

### 3. Registry Update
- Update `services/engine/pkg/license/license.go` with the new public key if rotated.
- Deploy the token to the target `.env` file via `sed` or the configuration manager.

## Failure Handling
- **`KEY_MISMATCH`**: If the token fails validation against the public key, ABORT and check the Private Key source.
- **`UNAUTHORIZED_CAPS`**: If `capabilities` exceed the `tier` permissions, flag as a Business Logic violation.

## Postconditions
- Artifact: `license_audit_log.json` must be updated with the `rotation_id`.
- The `licensePubKeyBase64` in the source code MUST match the token.
