---
name: manage_ocultar_license
description: >
  Generate, embed, and deploy OCULTAR Ed25519 license tokens.
  Use when: creating a new client license, renewing an expiring one,
  rotating the keypair after a compromise, or adding SOV_LICENSE_KEY
  to a missing/broken .env file.
---

# Skill: Manage OCULTAR License

## Context

OCULTAR uses **offline Ed25519 signatures** — no license server. The flow is:

1. `tools/tools/scripts/scripts/tools/scripts/scripts/keygen.go` — vendor tool. Generates a keypair AND signs a dynamic payload via CLI flags in one run.
2. `services/engine/services/engine/pkg/license/license.go` — has `licensePubKeyBase64` hardcoded. Must match the private key used to sign tokens.
3. `SOV_LICENSE_KEY` in `.env` — the signed token the client puts in their environment.

> **Critical**: `keygen.go` generates a **new** keypair every run. Only run it when you intend to rotate. To issue a new license with the **existing** keypair, you must use a custom signing script (or edit keygen.go to accept an existing private key).

---

## Task 1 — Issue a license (existing keypair, no rotation)

> Use this for new clients or renewals when the keypair in `license.go` is already correct.

### Step 1 — Check current public key

```bash
grep licensePubKeyBase64 /home/edu/services/engine/services/engine/pkg/license/license.go
```

Note the current public key — this tells you which private key to use for signing (stored in password manager).

### Step 2 — Write a one-off signing script

Because `keygen.go` always generates a new keypair, write `/tmp/sign_license.go` with the existing private key:

```go
//go:build ignore

package main

import (
    "crypto/ed25519"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "time"
)

const existingPrivKeyB64 = "PASTE_PRIVATE_KEY_FROM_PASSWORD_MANAGER_HERE"

func main() {
    privKeyBytes, _ := base64.StdEncoding.DecodeString(existingPrivKeyB64)
    privKey := ed25519.PrivateKey(privKeyBytes)

    payload := struct {
        CustomerName string `json:"CustomerName"`
        Tier         string `json:"Tier"`
        ExpiryDate   int64  `json:"ExpiryDate"`
    }{
        CustomerName: "CLIENT_NAME_HERE",
        Tier:         "enterprise",
        ExpiryDate:   time.Now().AddDate(1, 0, 0).Unix(), // 1 year
    }

    payloadBytes, _ := json.Marshal(payload)
    sig := ed25519.Sign(privKey, payloadBytes)

    fmt.Printf("SOV_LICENSE_KEY=%s.%s\n",
        base64.StdEncoding.EncodeToString(sig),
        base64.StdEncoding.EncodeToString(payloadBytes))
}
```

```bash
go run /tmp/sign_license.go
```

Copy the `SOV_LICENSE_KEY=...` output and deliver to the client securely.

---

## Task 2 — Bootstrap from scratch (first run OR rotation)

> Use this when `license.go` still has the placeholder key, or after a keypair compromise.

### Step 1 — Generate a new keypair + test token

```bash
cd /home/edu/dev/ocultar
# Example: Generate for 'Acme Corp' expiring in 2 years
go run tools/tools/scripts/scripts/tools/scripts/scripts/keygen.go --customer="Acme Corp" --tier="enterprise" --expiry-years=2
```

Save the output. You need:
- **Public Key** → embed in `license.go`
- **Private Key** → save to password manager, then delete from terminal history
- **SOV_LICENSE_KEY** → the test/dev token

### Step 2 — Embed the public key

```bash
# Edit services/engine/services/engine/pkg/license/license.go line 14:
#   var licensePubKeyBase64 = "NEW_PUBLIC_KEY_HERE"
```

Use the `replace_file_content` tool to update the line.

### Step 3 — Update .env files

There are two `.env` files that need `SOV_LICENSE_KEY`:

| File | When |
|---|---|
| `/home/edu/.env` | Dev environment (in-repo, distrobox) |
| `/home/edu/Testing/client_box/enterprise/.env` | Client-box simulation |

Add or update:
```
SOV_LICENSE_KEY=<token from keygen output>
```

### Step 4 — Verify the build compiles

```bash
cd /home/edu/dev/ocultar && go build ./...
```

Zero output = success.

---

## Task 3 — Debug a failing license

If the engine logs `[WARN] Invalid or expired license key. Reverting to Community Mode.`:

### Check 1 — Expiry

```bash
TOKEN="<value of SOV_LICENSE_KEY>"
PAYLOAD_B64=$(echo "$TOKEN" | cut -d. -f2)
echo "$PAYLOAD_B64" | base64 -d
# → {"CustomerName":"...","Tier":"enterprise","ExpiryDate":XXXXXXXXXX}
date -d @XXXXXXXXXX  # check if in the past
```

### Check 2 — Public key mismatch

```bash
grep licensePubKeyBase64 /home/edu/services/engine/services/engine/pkg/license/license.go
# Must match the keypair used to sign the token
```

If mismatched: either reissue the token with the correct private key, or re-embed the correct public key and rebuild.

### Check 3 — Token format

```bash
echo "$TOKEN" | grep -o '\.' | wc -l
# Must be exactly 1 dot (two parts)
```

---

## Key Files Reference

| Path | Purpose |
|---|---|
| `services/engine/services/engine/pkg/license/license.go` | Public key + verification logic |
| `tools/tools/scripts/scripts/tools/scripts/scripts/keygen.go` | Keypair generation + test token signing |
| `documentation/internal/LICENSE_MANAGEMENT.md` | Full internal vendor guide |
| `documentation/ENTERPRISE_SETUP_GUIDE.md §13` | Client-facing activation instructions |
| `.env` (ocultar) | Dev license token |
| `.env` (sombra) | Sombra gateway keys (no license token needed here) |
