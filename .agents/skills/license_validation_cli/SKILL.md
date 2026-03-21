---
name: license_validation_cli
description: Deterministic validator for OCULTAR license tokens. Verifies signature, expiry, and capabilities against the active engine public key.
---

# License Validation CLI (v1.1)

## Purpose

The LVC provides a "Dry-Run" channel to verify license integrity without affecting production runtime. It ensures that tokens are cryptographically aligned with the engine's hardcoded public key.

## Inputs / Outputs

### Inputs
- `license_token`: The full `Signature.Payload` string.
- `public_key_override` (Optional): Force validation against a specific key.

### Outputs
- `is_valid` (Boolean).
- `payload` (JSON): Decoded Customer, Tier, and Expiry metadata.
- `verification_log` (Artifact): Step-by-step verification trace.

## Preconditions
- Access to `services/engine/pkg/license/license.go` to extract the public key.

---

## Instructions

### 1. Extraction & Decoding
- Split `license_token` at the `.` delimiter.
- **Payload**: Base64-decode the second part to extract JSON metadata.
- **Signature**: Base64-decode the first part for crypto verification.

### 3. Cryptographic Verification
- Extract `PUB_KEY` from `license.go`.
- Use the standard `ocultar-license-tool check --token="{{license_token}}" --pubkey="{{PUB_KEY}}"` if available.
- **Fallback**: Use the `crypto/ed25519` verification logic.
- **Integrity**: Return `FALSE` if the signature does not match the payload + public key.

### 4. Expiry Check
- Parse `ExpiryDate` from the payload.
- Compare against current system time.
- **Verdict**: If `now > ExpiryDate`, return `is_valid = FALSE` with reason `EXPIRED`.

## Failure Handling
- **`MALFORMED_TOKEN`**: If the token contains fewer than two parts, return `INVALID_FORMAT`.
- **`BASE64_ERROR`**: If decoding fails, flag as `TAMPER_SUSPECTED`.

## Postconditions
- Verification MUST be completed before any automated rotation or deployment task.
