---
name: sbom-generator
description: Expert skill for generating a Software Bill of Materials (SBOM) for Ocultar release artifacts. Ensures compliance with enterprise and government security standards (Cyber Trust Mark, EO 14028).
---

# Ocultar | SBOM Generator

## Purpose
This skill generates a comprehensive list of all software components, dependencies, and licenses included in an Ocultar distribution. It provides transparency and allows clients to perform vulnerability scanning on the delivered package.

## When To Use This Skill
**MANDATORY** during the release process, specifically:
- After `client-package-updater` has finalized the manifest.
- Before `artifact-signer` signs the final package.
- Whenever a new third-party dependency (Go, **CGO/llama.cpp**, or NPM) is added.

## Instructions

### 1. Automation via Manifest
The SBOM generation is now **automated** within the `tools/scripts/scripts/manifest_executor.py`. 
When the `security.sanitization` flag is set to `true` in `dist.manifest.yaml`, the executor will automatically:
- **Go Dependencies**: Run `go list -json -m all > dist/[dist]/sbom_go.json`.
- **NPM Dependencies**: Run `npm list --json > dist/[dist]/sbom_npm.json`.

### 2. Manual Attribution (Failsafe)
If the automated process fails or excludes specific proprietary modules:
- Capture the state manually using the commands above.
- Ensure the JSON files are placed in the `dist/` staging directory before packaging.

### 3. SBOM Formatting
Merge the data into a compliant format (JSON/CycloneDX). 
> [!NOTE]
> If `syft` or `cyclonedx-cli` are available in the container, use them to generate a standard `.spdx` or `.json` file. Otherwise, maintain the raw `sbom_go.json` and `sbom_npm.json` as the source of truth.

### 4. Inclusion in Distribution
Ensure the generated SBOM files are moved to the staging directory before the final archive is created:
- Community: `cp dist/sbom_*.json dist/community/`
- Enterprise: `cp dist/sbom_*.json dist/enterprise/`

## Outputs
- `sbom_go.json`: Detailed map of Go modules and versions.
- `sbom_npm.json`: Detailed map of NPM packages and versions.
- `LICENSE_SUMMARY.md`: A human-readable summary of all third-party licenses found.

## Failure Handling
- **Missing Dependencies**: If `go list` fails, ensure `go work sync` has been run.
- **Vulnerability Detected**: If an SBOM scan (external) reveals a high-risk CVE, YOU MUST report this to the `CISO & Compliance Officer` skill before proceeding.
