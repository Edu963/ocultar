---
name: license_validation_cli
description: >
  Deterministic validator for OCULTAR license tokens.
  Verifies signature, expiry, and capabilities against the active engine public key.
---

# Skill: License Validation CLI

## Context
Use this skill to verify if a given `OCU_LICENSE_KEY` is valid for the current engine build without needing to deploy or restart the engine.

## Inputs / Outputs

### Inputs
| Variable | Description |
|---|---|
| `OCU_LICENSE_KEY` | The full license token (Signature.Payload) |

### Outputs
| Variable | Description |
|---|---|
| `IS_VALID` | Boolean indicating if the signature is authentic |
| `DECODED_PAYLOAD` | JSON object containing Customer, Tier, and Expiry |

## Steps

### 1. Extract and Decode Payload
```bash
TOKEN="<OCU_LICENSE_KEY>"
PAYLOAD_B64=$(echo "$TOKEN" | cut -d. -f2)
echo "$PAYLOAD_B64" | base64 -d | jq .
```

### 2. Check Expiry
```bash
EXPIRY=$(echo "$TOKEN" | cut -d. -f2 | base64 -d | jq .ExpiryDate)
date -d @$EXPIRY
```

### 3. Verify Signature (Self-Contained Go Script)
Run the following check to verify the signature against the current hardcoded key in `license.go`:

```bash
# This requires querying the public key from the engine first
PUB_KEY=$(grep "licensePubKeyBase64 =" services/engine/pkg/license/license.go | cut -d'"' -f2)

# Create a temporary validator
cat <<EOF > /tmp/val.go
package main
import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strings"
)
func main() {
	pub := "$PUB_KEY"
	token := "$TOKEN"
	parts := strings.Split(token, ".")
	sig, _ := base64.StdEncoding.DecodeString(parts[0])
	val, _ := base64.StdEncoding.DecodeString(parts[1])
	puk, _ := base64.StdEncoding.DecodeString(pub)
	fmt.Printf("Valid: %v\n", ed25519.Verify(puk, val, sig))
}
EOF

go run /tmp/val.go
rm /tmp/val.go
```

## Validation
- If `Valid: true`, the token is cryptographically sound for this engine version.
- If `Valid: false`, the token was signed with a different private key or has been tampered with.
