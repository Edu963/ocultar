---
name: change-impact-visualizer
description: Automatically maps which docs, packages, or components are impacted by a code change. Useful for compliance audits and enterprise reporting.
---

# Change Impact Visualizer

## Purpose

This skill provides a high-level "Impact Narrative" for every modification. It ensures that engineers and auditors understand the ripple effects of a change across the Ocultar ecosystem.

## When To Use This Skill

Use this skill:
- After completing a task involving code, configuration, or documentation.
- When generating a pull request or release summary.
- During a compliance review to verify that all impacted areas were updated.

## Instructions

1.  **Trace Code Changes**: Identify all modified files and their functional roles (Engine, Proxy, Dashboard, Config).
2.  **Map to Regulatory Articles**: Determine if the change touches PII handling (GDPR), Health data (HIPAA), or Proprietary terms (Business Secrets).
3.  **Identify Document Drift**: Check if technical docs (`ARCHITECTURE.md`) or pilot guides (`onboarding.md`) need updates based on the change.
4.  **Assess Distribution Effect**: Determine if the change affects the contents of the client-facing `.tar` or `.zip` packages.
5.  **Generate Impact Summary**: Create a concise table or bulleted list summarizing the "Primary Impact" and "Secondary Effects."

## Examples

### Changing Encryption Logic
**Change**: Updated `AES-GCM` key rotation in `engine.go`.
**Impact Summary**:
- **Security**: Hardens data at rest (GDPR compliance).
- **Documentation**: Requires update to `SECURITY.md` (internal) and `SETUP_GUIDE.md` (master key handling).
- **Distribution**: No new files added, but binary builds are impacted.
- **Audit**: Log events will now include new rotation metadata.
