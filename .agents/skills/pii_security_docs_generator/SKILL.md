---
name: pii-security-docs-generator
description: Creates documentation about new PII rules, Dictionary Shield updates, and audit compliance processes.
---

# PII & Security Docs Generator

## Purpose

This skill ensures that every security enhancement is matched with high-quality, auditor-ready documentation. It maintains the technical and regulatory narrative of the Ocultar Refinery.

## When To Use This Skill

Use this skill when:
- The `refinery-rule-generator` creates a new detection rule.
- `Dictionary Shield` keywords are updated.
- Regulatory requirements (GDPR, HIPAA) for certain PII types change.
- Preparing for a security audit or providing a transparency report to a client.

## Instructions

1.  **Analyze Rule Definition**: Review the regex, dictionary, or SLM logic for the new rule.
2.  **Define Compliance Mapping**: Explicitly state which regulatory article (e.g., GDPR Art. 28) this rule supports.
3.  **Update PII_DETECTION.md**: Add the new rule description, example matches, and redaction behavior to the master PII documentation.
4.  **Generate Audit Logs Summary**: (Internal) Update documentation on how this specific rule is reported in the Enterprise SIEM logs.
5.  **Draft User Transparency Notes**: Create simple explanations for end-users on how their data is being "refined" by this new rule.
6.  **Verify Cross-References**: Ensure setup guides and developer docs are updated if the new rule requires specific configuration tweaks.

## Examples

### Documenting German IBAN Rule
**Action**: Add German IBAN to `PII_DETECTION.md`, attach it to the `REDU_REG_FINANCIAL` category, and provide examples of redacted vs. original strings.
