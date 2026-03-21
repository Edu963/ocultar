---
name: manage_ocultar_license
description: >
  Professional-grade orchestrator for the OCULTAR Ed25519 offline license ecosystem.
  Manages keypair generation, license issuance, and validation integrity.
---

# Skill: Manage OCULTAR License

## Context

OCULTAR implements a **Zero-Egress** offline license verification model using Ed25519 signatures. No external license server is required.

### Core Architecture
1. **Verification Logic**: [license.go](file:///home/edu/ocultar/services/engine/pkg/license/license.go) contains the hardcoded `licensePubKeyBase64`.
2. **Environment Key**: The engine expects the signed token in the `OCU_LICENSE_KEY` environment variable.
3. **Tooling**: [keygen.go](file:///home/edu/ocultar/tools/scripts/scripts/keygen.go) handles keypair generation and token signing.

---

## Inputs / Outputs

### Inputs
| Variable | Description | Default |
|---|---|---|
| `CUSTOMER_NAME` | Legal name of the licensee | "Acme Corp" |
| `TIER` | Support tier (`community` \| `enterprise`) | "enterprise" |
| `EXPIRY_YEARS` | Validity period in years | 1 |
| `EXISTING_PRIV_KEY` | (Optional) Private key for signing without rotation | N/A |

### Outputs
| Variable | Description | Destination |
|---|---|---|
| `OCU_LICENSE_KEY` | Signed license token | `.env` |
| `PUBLIC_KEY` | Ed25519 Public Key | `license.go` |
| `PRIVATE_KEY` | Ed25519 Private Key | Password Manager |

---

## Task 1 — Issue New License (Existing Keypair)

> Use this for standard renewals or adding new clients to an existing deployment.

### 1. Retrieve Private Key
Securely retrieve the `PRIVATE_KEY` from the enterprise password manager matching the current public key in `license.go`.

### 2. Verify Key Matching
```bash
# Check current public key in engine
grep "licensePubKeyBase64 =" services/engine/pkg/license/license.go
```

### 3. Sign License
Because currently `keygen.go` forces rotation, you must use the deterministic `sign_license` utility (or proposed `keygen.go` upgrade).
**CAUTION**: Do not leave private keys in shell history or temporary files.

```bash
# [Proposed Upgraded Command]
# go run tools/scripts/scripts/keygen.go --customer="Target Client" --private-key="<secret>"
```

---

## Task 2 — Bootstrap / Key Rotation

> Use this for initial deployment or after a suspected keypair compromise.

### 1. Generate Keypair & Token
```bash
cd /home/edu/ocultar
go run tools/scripts/scripts/keygen.go --customer="Internal Dev" --tier="enterprise" --expiry=2
```

### 2. Update Engine Verification
Update the hardcoded public key in `services/engine/pkg/license/license.go`:

```go
// services/engine/pkg/license/license.go:14
var licensePubKeyBase64 = "NEW_PUBLIC_KEY_OUTPUT_FROM_STEP_1"
```

### 3. Deploy Token to Environment
Update all relevant infrastructure `.env` files:

```bash
# Update .env
sed -i 's/OCU_LICENSE_KEY=.*/OCU_LICENSE_KEY=SIGNATURE.PAYLOAD/' .env
```

---

## Task 3 — Validation & Troubleshooting

### 1. Verify Token Payload
```bash
TOKEN="<OCU_LICENSE_KEY_VALUE>"
echo "$TOKEN" | cut -d. -f2 | base64 -d | jq .
```

### 2. Verify Expiry Date
```bash
EXPIRY=$(echo "$TOKEN" | cut -d. -f2 | base64 -d | jq .ExpiryDate)
date -d @$EXPIRY
```

## Failure Handling

| Issue | Remediation |
|---|---|
| `Invalid or expired license` | Check if `licensePubKeyBase64` matches the signing private key. |
| `Marshalling Error` | Ensure `CUSTOMER_NAME` contains no illegal characters for JSON. |
| `Signature Mismatch` | Verify the token was not truncated during copy-paste. |

## Postconditions
- **Clean Environment**: No `/tmp` scripts containing private keys remain.
- **Verification**: `go build ./services/engine/...` completes without error.
- **Audit Log**: Record the license issuance in the internal compliance ledger.
