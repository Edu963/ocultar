---
name: compliance-certificate-signer
description: Expert skill for signing regulatory policies using Ed25519.
---

# Compliance Certificate Signer (v1.1)

## Purpose

Ensures the "Immutable Chain of Custody" for regulatory policies. It prevents unauthorized modifications to the `regulatory_policy.json` by appending a cryptographic signature that the Sombra Gateway validates during boot.

## Inputs / Outputs

### Inputs
- `policy_file` (Path): Global policy path.
- `signing_key` (Secret): Ed25519 private key.

### Outputs
- `signed_policy` (File): Rich JSON containing the `signature` and `signer_id`.
- `receipt` (JSON): Verification metadata for the `ecosystem-state-tracker`.

## Preconditions
- `jq` MUST be available for canonicalization.
- `signing_key` MUST be loaded into the ephemeral session.

---

## Instructions

### 1. Canonicalization
- Read the policy and pipe through `jq -S '.'`.
- **Requirement**: ALL keys must be sorted alphabetically and all whitespace removed to ensure a stable hash.

### 2. Hashing & Signing
- Generate SHA-256 hash of the canonical string.
- Sign using Ed25519.
- Encode in Base64.

### 3. Payload Enrichment
Inject the signature into the JSON root:
```json
{
  "protocol": "OCU_SIG_V1",
  "signature": "BASE64",
  "mappings": { ... }
}
```

### 4. Verification Check
- Immediately verify the signed file using the corresponding Public Key.
- **Fail**: If verification fails, return `SIGNATURE_BOOT_FAILURE`.

## Failure Handling
- **`KEY_CORRUPTION`**: If the private key fails to load, halt the pipeline.
- **`TAMPER_ON_WRITE`**: If the file contents change between signing and verification, trigger an audit alert.

## Postconditions
- Signed policy MUST be registered in `ecosystem-state-tracker`.
