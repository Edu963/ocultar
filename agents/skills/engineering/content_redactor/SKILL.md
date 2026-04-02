---
name: content-redactor
description: Sanitizes distribution artifacts and client-facing packages. Removes environment files, local configurations, and internal-only data from the final deliverable.
---

# Content Redactor (Post-build / Egress)

## Purpose
The Content Redactor ensures that the final "deliverable" (the package the client actually receives) is clean, safe, and free of any Ocultar-internal metadata or temporary development data.

## When To Use This Skill
- During the `Phase C` of the `continuous-ai-orchestrator`.
- Immediately before generating a `.tar`, `.zip`, or container image for distribution.
- When preparing documentation for external publication.

## Instructions

### 1. Enforce Distribution Policies
- Verify that every file in the distribution list complies with the `dist.manifest.yaml`.
- **Blacklist Enforcement**: Ensure `.env`, `.git`, `license.key`, `.DS_Store`, and `node_modules` (if not bundled) are purged.

### 2. Dynamic Redaction
- Scan the compiled/bundled content for any lingering safe-placeholders (e.g., `YOUR_API_KEY_HERE`) to ensure they haven't been replaced by actual keys during a botched build.

### 3. Internal Path Scrubbing
- Recursively search for and remove any references to internal build paths or developer-specific file URIs in logs or documentation.

### 4. Integrity Check
- Generate a checksum for the sanitized bundle.
- **Gate**: If a blacklisted file is found in the bundle, **HALT** the release process.

## Examples

### Scenario: Preparing a Community Build
- **Input**: A `dist/` folder containing `ocultar-refinery` and a lingering `.env.local`.
- **Action**: Delete `.env.local` and verify that the `config.yaml` uses default placeholders.
