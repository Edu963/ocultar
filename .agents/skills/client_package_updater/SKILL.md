---
name: client-package-updater
description: Use this skill whenever changes to the codebase may affect files distributed to clients in packaged formats such as .tar or .zip. The skill ensures that client deliverables remain complete, functional, and aligned with the current version of the software.
---

# Client Package Updater

## Purpose

This skill ensures that all artifacts delivered to clients (such as .tar or .zip archives) remain consistent with the current codebase. Whenever relevant code, dependencies, configuration files, or scripts change, the packaging process and contents must be reviewed and updated if necessary.

## When To Use This Skill

Use this skill whenever:

- Files included in client distributions change
- Build scripts or packaging scripts are modified
- New dependencies are added
- Configuration files required by clients change
- Installation scripts change
- Directory structure changes
- Runtime assets are added or removed

Do NOT use this skill for internal-only code changes that do not affect distributed artifacts.

## Instructions

1. Detect the change made to the repository.
2. Determine whether the change impacts the client distribution package.
3. Identify the packaging mechanism used by the project. This may include:
   - build scripts
   - packaging scripts
   - Docker exports
   - release workflows
4. Identify the files included in the client package (.tar, .zip, or similar).
5. Verify whether the new code changes require:
   - adding files
   - removing files
   - updating configuration
   - updating install scripts
   - ensuring `README.md`, `roadmap.md`, and relevant `/documentation/*.md` guides are bundled
6. Update the packaging configuration if necessary.
7. Ensure the package remains minimal, functional, and reproducible.
8. If applicable, regenerate or validate the archive structure.

## Examples

Example 1

Input:
A new runtime configuration file is required for the application to start.

Action:
Ensure the configuration file is included in the client package.

Example 2

Input:
A dependency script is moved from `tools/scripts/scripts/` to `tools/`.

Action:
Update the packaging process so the script is included in the distribution.

Example 3

Input:
A new installation script is added.

Action:
Include the script in the package and update any installation documentation if necessary.
