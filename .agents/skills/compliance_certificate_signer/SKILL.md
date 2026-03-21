---
name: compliance-certificate-signer
description: Expert skill for signing regulatory policies using Ed25519. Prevents policy-tampering by ensuring only authorized Ocultar policies are loaded by the Sombra Gateway.
---

# Compliance Certificate Signer (v1.0)

## Purpose

The Compliance Certificate Signer ensures the "Immutable Chain of Custody" for regulatory policies. It signs the final `regulatory_policy.json` with an Ocultar Ed25519 private key, appending a `.sig` or embedding a `signature` field to prevent unauthorized modifications to the policy.

## Preconditions

- **Input**: A `regulatory_policy.json` (JSON).
- **Security**: Access to the `OCU_POLICY_PRIVATE_KEY` (Safe environment variable or vault).
- **Tools**: `openssl`, `sha256sum`, or Ocultar's internal `artifact-signer`.

## Inputs / Outputs

### Inputs
- `policy_file` (Path): Path to the `regulatory_policy.json` (default: `security/regulatory_policy.json`).
- `private_key` (Secret): Ed25519 private key for signing.

### Outputs
- `signed_policy` (File): The policy with an embedded `signature` field.
- `certificate_manifest` (Artifact): A structured summary of the signature, signer ID, and timestamp.

---

## Instructions

### Step 1 – Policy Canonicalization
1.  Read the `policy_file`.
2.  Canonicalize the JSON (remove extra whitespace, sort keys) using `jq -S '.'` to ensure a deterministic hash.
3.  Generate a SHA-256 hash of the canonical JSON.

### Step 2 – Cryptographic Signing
1.  Sign the SHA-256 hash using the Ed25519 `private_key`.
2.  Encode the resulting signature in Base64 or Hex.

### Step 3 – Policy Enrichment
1.  Append the signature to the root level of the JSON:
    ```json
    {
      "version": "X.Y",
      "mappings": { ... },
      "signature": "BASE64_SIGNATURE_HERE",
      "signer": "OCULTAR_CISO_CA"
    }
    ```
2.  Write the enriched JSON back to `policy_file`.

### Step 4 – Verification
1.  Immediately verify the signature using the corresponding public key (`OCU_POLICY_PUBLIC_KEY`).
2.  If verification fails, **HALT** and alert the `Security Advisory Scanner`.

---

## Failure Handling

- **`PRIVATE_KEY_MISSING`**: If the signing key is not found in the environment.
- **`SIGNATURE_MISMATCH`**: If the verification step fails post-signing.

---

## Ecosystem Role
- **Role**: Security / Integrator.
- **Consumer**: Sombra Gateway (Boot sequence).
- **Trigger**: `continuous-ai-orchestrator` (Final release step).
