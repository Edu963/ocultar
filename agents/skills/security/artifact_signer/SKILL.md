---
name: artifact-signer
description: Expert skill for signing release artifacts using Ed25519. Prevents supply-chain attacks by ensuring only authorized Ocultar artifacts are distributed.
---

# Ocultar | Artifact Signer

## Purpose
This skill adds a cryptographic layer of trust to Ocultar distributions. It generates detached signatures for `.tar` and `.zip` archives, allow clients to verify the integrity and origin of the software.

## When To Use This Skill
**MANDATORY** for every release package before distribution.
- Must run **AFTER** `release-artifact-builder` has created the final archives.
- Must run **AFTER** `sbom-generator` has included the SBOM in the bundle.

## Preconditions
- Access to the **Ocultar Release Private Key** (stored securely, never in code).
- The `dist/*.zip` and `dist/*.tar.gz` files exist.

## Instructions

### 1. Identify Signing Target
Locate the final distribution archives:
- Community: `dist/ocultar-community.zip`
- Enterprise: `dist/ocultar-enterprise.tar.gz`

### 2. Implementation Logic
If a dedicated tool is not available, use the following Go implementation pattern:

```go
// Pattern for a one-off signing script
package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"os"
)

func main() {
	privKey, _ := base64.StdEncoding.DecodeString(os.Getenv("RELEASE_PRIVATE_KEY"))
	data, _ := os.ReadFile("dist/ocultar-community.zip")
	sig := ed25519.Sign(privKey, data)
	os.WriteFile("dist/ocultar-community.zip.sig", []byte(base64.StdEncoding.EncodeToString(sig)), 0644)
}
```

### 3. Generate Signatures
For each archive:
1. Load the archive into memory or stream it.
2. Sign with the **Ocultar Ed25519 Private Key**.
3. Save the detached signature as `FILE_NAME.sig`.

### 4. Provide Verification Instructions
Include the Public Key and instructions in the package or on the download page:
- `ocultar --verify dist/ocultar-community.zip`

## Outputs
- `ocultar-community.zip.sig`: Detached signature.
- `ocultar-enterprise.tar.gz.sig`: Detached signature.

## Failure Handling
- **Key Missing**: Abort if `RELEASE_PRIVATE_KEY` is not set.
- **Verification Failure**: If you cannot verify your own signature immediately after creation, the artifact is corrupt or the key is mismatched. REBUILD.
