---
name: continuous-ai-orchestrator
description: Automatically orchestrates all Ocultar AI skills whenever meaningful changes occur in the codebase, documentation, or configuration. Ensures all actions respect product context, repository knowledge, and enterprise security/privacy principles.
---

# Continuous AI Orchestrator

## Purpose

This skill monitors changes in the Ocultar repository and automatically triggers the appropriate sequence of skills, ensuring:

- Documentation is updated
- Client packages are accurate
- Secrets are never leaked
- Release artifacts are correctly built
- All changes respect Product Context and Repository Knowledge Map

It effectively **creates a self-managing AI workflow**.

## When To Use This Skill

Use this skill whenever:

- Code, configuration, or architecture changes occur  
- New features or bug fixes are implemented  
- Pilot program workflows are updated  
- Release candidates are prepared  
- Documentation is modified  

Do NOT run for trivial formatting or comment changes.

## Instructions

1. **Detect Change Scope**  
   - Identify which files and components were modified  
   - Classify the type of change: code logic, API, architecture, configuration, documentation, pilot program, or packaging

2. **Consult Product Context**  
   - Determine which product principles, components, or regulatory rules are affected  

3. **Consult Repository Knowledge Map**  
   - Locate the correct directories and documentation to modify  
   - Identify impacted components (Ocultar engine, Sombra Gateway, Dashboard, Identity Vault, etc.)

4. **Run Skills in Correct Sequence**  

    a. **Documentation Updater** – Update affected documentation  
    b. **PII Security Docs Generator** – Generate auditor reports for new rules  
    c. **Client Package Updater** – Ensure packaging rules are correct for client deliverables  
    d. **Security Sanitizer** – Remove secrets and validate general package safety  
    e. **PII Detection Updater** – Update refinery rules and regression tests  
    f. **Refinery Rule Generator** – Automatically generate regex/configs for new rules  
    g. **Dictionary Shield Manager** – Update keyword protection dictionaries  
    h. **Audit Log Validator** – Verify regulatory mapping and log integrity  
    i. **Sombra Gateway Policy Enforcer** – Validate architectural compliance in Sombra  
    j. **Enterprise Dashboard Integrity Checker** – Prevent UI regressions and license mismatches  
    k. **Release Artifact Builder** – Build clean, versioned, Zero-Egress artifacts  
    l. **Change Impact Visualizer** – Generate final impact summary for auditors  
    m. **Red-Team Evasion Scanner** – Proactively stress-test the refinery for bypasses  
    n. **Industry-Specific Snapshot Generator** – Pre-configure vertical-specific demo/onboarding environments  
    o. **ROI & Cost-Efficiency Accountant** – Quantify the financial impact and security cost-avoidance  
    p. **Sombra Performance Benchmarker** – Monitor latency and optimize the refinery pipeline  

5. **Validate Outputs**  
   - Ensure documentation matches the latest code  
   - Confirm packaged artifacts are complete and secure  
   - Check that all changes adhere to Product Context principles (Zero-Egress, Fail-Closed, Privacy-by-Design, Regulatory Alignment).
- **Fail-Closed Security**: If the engine or `protected_entities.json` (located in `services/engine/configs/`) is missing or unreachable, the workflow must terminate immediately.

6. **Optional Logging**  
   - Record a summary of actions taken for audit and compliance purposes

## Examples

Example 1

Input:
A new API endpoint is added to Sombra and changes are made to PII detection logic.

Action:
- Update API_REFERENCE.md via Documentation Updater  
- Adjust packaging rules if endpoint scripts are included via Client Package Updater  
- Run PII Detection Updater to verify new refinement rules  
- Run Audit Log Validator to ensure regulatory mapping is correct  
- Run Security Sanitizer to verify final package security  
- Prepare a versioned release artifact via Release Artifact Builder  
- Validate that all steps comply with Product Context principles

Example 2

Input:
Pilot program onboarding workflow is updated.

Action:
- Update docs/pilot/onboarding.md and playbook.md  
- Verify pilot deliverables via Client Package Updater  
- Ensure no secrets or sensitive data leak via Security Sanitizer  
- Generate release artifact if it’s a deliverable milestone

Example 3

Input:
Enterprise dashboard sidebar is modified.

Action:
- Update relevant UI documentation  
- Run Enterprise Dashboard Integrity Checker to prevent regressions  
- Adjust client package only if necessary  
- Run Security Sanitizer  
- Build release artifact if part of a stable milestone
