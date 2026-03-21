---
name: ai-development-protocol
description: Use this skill whenever meaningful changes are made to the codebase. The skill coordinates documentation updates, packaging verification, security sanitation, and release preparation to maintain a stable and distributable project.
---

# AI Development Protocol

## Purpose

This skill defines the standard workflow the AI agent must follow whenever significant changes are made to the project. It ensures that the repository remains consistent, secure, and ready for distribution.

The protocol coordinates the following project skills:

- Documentation Updater
- Client Package Updater
- Security Sanitizer
- Release Artifact Builder

By following this protocol, the AI agent maintains project quality, documentation accuracy, safe packaging, and reliable release artifacts.

## When To Use This Skill

Use this skill whenever:

- New features are added
- Code logic changes
- APIs change
- Project structure changes
- Configuration files change
- Dependencies change
- Packaging or distribution files change

Do NOT run the full protocol for trivial formatting or comment changes.

## Instructions

Follow the steps below in order.

### Step 1 – Analyze the Change

Determine the scope and impact of the code change.

Identify whether it affects:

- documentation
- packaging
- configuration
- dependencies
- client deliverables
- release artifacts

### Step 2 – Update Documentation

If the change affects system behavior, architecture, APIs, configuration, or workflows:

Run the Documentation Updater skill to ensure that project documentation remains accurate.

### Step 3 – Verify Client Packaging

If the change affects files distributed to clients:

Run the Client Package Updater skill to ensure packaging rules and included files remain correct.

### Step 4 – Sanitize Sensitive Data

Before generating distributable artifacts:

Run the Security Sanitizer skill to ensure no secrets, credentials, or internal data are exposed.

### Step 5 – Build Release Artifact (If Needed)

If the change represents a stable milestone, feature completion, or deliverable version:

Run the Release Artifact Builder skill to generate a clean versioned archive.

### Step 6 – Validate Project Integrity

Ensure that:

- documentation matches the codebase
- packaging rules remain correct
- no secrets are exposed
- release artifacts are clean and reproducible

## Examples

Example 1

Input:
A new API feature is implemented.

Action:
Update documentation, verify packaging rules, run security sanitization, and prepare a new release artifact if the feature is part of a deliverable version.

Example 2

Input:
The project structure changes and several configuration files are updated.

Action:
Update documentation, verify packaging rules, and sanitize configuration files before distribution.

Example 3

Input:
A major feature milestone is completed.

Action:
Run the full protocol including release artifact generation.
