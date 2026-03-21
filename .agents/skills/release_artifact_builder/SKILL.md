---
name: release-artifact-builder
description: Use this skill when preparing a new version of the software for distribution. The skill generates clean, versioned release artifacts such as .tar or .zip archives that contain all required files for deployment or delivery to clients.
---

# Release Artifact Builder

## Purpose

This skill creates reproducible release artifacts for the project. It packages the necessary files into a versioned archive suitable for distribution to clients or deployment environments.

The skill ensures that releases are clean, consistent, and contain only the required components.

## When To Use This Skill

Use this skill when:

- A feature or milestone is completed
- A stable version is ready for distribution
- A client delivery package must be created
- A release candidate is prepared
- Versioned artifacts are needed for testing or deployment

Do NOT use this skill for minor internal commits or experimental changes.

## Instructions

1. Determine the current version of the project.
   - Use version files, tags, or project metadata if available.
2. Identify the files required for a distributable release.
   - Source code required to run the application
   - Configuration templates
   - Runtime assets
   - Installation scripts
3. Exclude unnecessary files such as:
   - development artifacts
   - temporary files
   - local caches
   - test data not required for production
4. Validate that required dependencies and runtime files are included.
5. Perform Zero-Egress Check: Ensure no hardcoded external URLs or unauthorized phone-home scripts are present in the final binaries or scripts.
6. PII Leakage Scan: Run the `security-sanitizer` specifically on the final staging directory to ensure no local PII (e.g., from developer testing) is captured in the archive.
7. Create a versioned archive using a consistent naming format.

Example naming formats:

project-name-v1.2.0.tar  
project-name-v1.2.0.zip

6. Ensure the archive structure is clean and reproducible.
7. If necessary, verify the package by inspecting its contents.

## Examples

Example 1

Input:
Version 1.3.0 is ready for client distribution.

Action:
Generate the archive:

project-name-v1.3.0.tar

containing all necessary runtime files.

Example 2

Input:
A new feature release is completed.

Action:
Build a new versioned package and prepare it for client delivery.

Example 3

Input:
A release candidate needs to be shared for testing.

Action:
Generate a versioned archive suitable for external distribution.
