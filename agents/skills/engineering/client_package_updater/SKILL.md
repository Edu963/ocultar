---
name: client-package-updater
description: Expert skill for synchronizing the Ocultar codebase with client-facing distribution artifacts. Ensures that any changes to binaries, configurations, or scripts are reflected in the build manifests and packaging logic.
---

# Ocultar | Client Package Updater

## Purpose
This skill ensures that the **Ocultar Distribution Manifest** (currently implemented in `tools/scripts/scripts/build_release.sh`) remains synchronized with the evolving codebase. It prevents "Dark Features" (implemented but not shipped) and "Broken Shipments" (shipped but missing dependencies).

## When To Use This Skill
**MANDATORY** whenever a change affects:
- **Core Binaries**: New `main.go` entry points or renamed service commands.
- **Enterprise Extensions**: Any addition to the `enterprise/` directory that must be bundled separately.
- **Static Assets**: Updates to the `apps/dashboard` UI or new public icons/locales.
- **AI Models**: New GGUF weights or configuration in the `models/` directory.
- **Configuration Templates**: New keys in `configs/config.yaml` or `.env.example`.
- **Deployment Scripts**: Modifications to `docker-compose.proxy.yml` or `setup-community.sh`.
- **Documentation**: New `/docs/*.md` guides that must be included in the `.zip` or `.tar.gz`.

## Preconditions
- The agent has completed a code or configuration change.
- The `tools/scripts/scripts/build_release.sh` script is available and readable.

## Instructions

### 1. Change Impact Mapping
Identify the specific category of the change:
- **Binary Change**: Does `go build` output a new filename?
- **Logic Change**: Does a new service require a new port mapping in `docker-compose.yml`?
- **Asset Change**: Has the `apps/dashboard/dist` structure changed?

### 2. Manifest Synchronization
The **Ocultar Distribution Manifest** (`dist.manifest.yaml`) is the single source of truth for all releases.
1. **Audit `dist.manifest.yaml`**: Ensure all new files (e.g., extensions, config keys) are explicitly listed in the appropriate `distributions` section.
2. **Component Mapping**: Update `src` and `dest` paths if directory structures change.
3. **Exclusion Logic**: Verify that `security` flags (`sanitization`, `egress_audit`) are enabled for new components.
4. **Integrity Manifests (CRITICAL)**: Whenever shipping an Enterprise binary, ensure the `security/dashboard_integrity.json` and other related manifests are explicitly bundled in the distribution.
5. **Archive Structure**: Verify that the distribution format (zip/tar.gz) includes a single parent folder named after the package to ensure clean extraction.
6. **Script Sync**: Ensure `tools/scripts/scripts/build_release.sh` reflects the logic defined in the manifest.

### 3. Security & Compliance (MANDATORY)
1. **Secret Scanning**: Invoke the `secret-scanner` skill on the `dist/` staging directory to prevent credential leaks.
2. **Content Sanitization**: Use the `content-redactor` to remove internal build paths and developer metadata.
3. **Policy Integrity**: Ensure the `regulatory_policy.json` is signed via the `compliance-certificate-signer` before final compression.
4. **SBOM Generation**: Invoke the `sbom-generator` skill to include dependency manifests in the bundle.

### 4. Verification Flow
1. **Dry Run**: Execute `tools/scripts/scripts/build_release.sh` locally.
2. **Parity Check**: Run `bash tools/scripts/scripts/check_parity.sh` to ensure Community/Enterprise isolation.
3. **Smoke Test**: Invoke the `distribution-integrity-validator` skill to verify the package in a clean environment.

## Inputs
- **Changeset**: Description of the changes made to the repository.
- **Package Target**: (Optional) Specific target to update (Community/Enterprise).

## Outputs
- **Updated Manifest**: Modified `tools/scripts/scripts/build_release.sh`.
- **Validation Report**: Confirmation of correct file placement and security sanitization.

## Failure Handling
- **Build Failure**: If `build_release.sh` fails after your change, analyze the diff of the script and revert if logic is circular or paths are broken.
- **Leaked Secret**: If `security-sanitizer` detects a secret in `dist/`, YOU MUST locate the source in the repo and fix it before re-running this skill.

## Examples

### Example 1: New Enterprise Connector
**Input**: Added `enterprise/connectors/sharepoint.go`.
**Action**: Update `build_release.sh` to ensure `enterprise/connectors` is built and the .so/.go binary is included in `ocultar-enterprise.tar.gz`.

### Example 2: Dashboard Redesign
**Input**: Moved frontend build artifacts to `apps/dashboard/dist`.
**Action**: Update the `npm run build` step and the `cp -r` path in `build_release.sh` to match the new `dist` directory.

### Example 3: Deployment Script Update
**Input**: Added `scripts/harden-linux.sh`.
**Action**: Add `cp "$SCRIPT_DIR/harden-linux.sh" "$DIST_DIR/community/scripts/"` to the community preparation section of `build_release.sh`.
