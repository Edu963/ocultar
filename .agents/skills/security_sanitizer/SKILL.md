---
name: security-sanitizer
description: Use this skill before building client packages or release artifacts. The skill scans the repository and distribution contents to detect and remove sensitive information such as API keys, secrets, private credentials, internal paths, or confidential test data.
---

# Security Sanitizer

## Purpose

This skill protects the project from accidentally exposing sensitive information when sharing code or distributing software to clients.

It scans files and packaging artifacts to detect secrets, credentials, or internal data that should never be included in distributed packages.

## When To Use This Skill

Use this skill when:

- Preparing a client distribution package (.tar or .zip)
- Creating a release archive
- Exporting code for external sharing
- Updating packaging rules
- Preparing a repository snapshot

Always run this skill before generating distributable artifacts.

## Instructions

1. Scan the repository and packaging contents for sensitive information.

2. Detect potential secrets, including:

- API keys
- private tokens
- passwords
- secret environment variables
- database credentials
- SSH keys
- private certificates

3. Identify files that should never be distributed, such as:

- `.env`
- `.env.local`
- `.env.production`
- local configuration overrides
- credential files
- development logs
- internal datasets
- temporary files
- local caches

4. Replace sensitive values with safe placeholders when needed.

Example:

API_KEY=YOUR_API_KEY_HERE

5. Ensure that secrets are loaded from environment variables rather than hardcoded values.

6. Verify that sensitive files are excluded from packaging rules.

7. Confirm that the final distribution package contains no credentials or internal data.
8. Validate enterprise artifact security by checking checksums and signature files if applicable.
9. Ensure that no internal-only documentation or development/debug scripts are included in the final deliverable.

## Examples

Example 1

Input:
A `.env` file contains database credentials.

Action:
Exclude the file from the distribution package and ensure environment variables are used instead.

Example 2

Input:
A configuration file contains a hardcoded API key.

Action:
Replace the key with a placeholder and document that the value must be provided via environment variables.

Example 3

Input:
A client release archive is about to be generated.

Action:
Scan the files included in the archive and remove any sensitive data before packaging.
