---
name: repository-knowledge-map
description: Programmable discovery skill for understanding the Ocultar repository structure, documentation clusters, and conceptual component boundaries.
---

# Repository Knowledge Map (v2.1)

## Purpose

This skill provides AI agents with a deterministic, programmatic mapping of the repository. It moves beyond static lists to provide structured discovery of technical, operational, and compliance artifacts.

## Inputs / Outputs

### Inputs
- `target_domain` (Enum): `CORE`, `DOCS`, `PILOT`, `REPORTS`, `CONNECTORS`.
- `recursive` (Boolean): Whether to walk the full subtree.

### Outputs
- `map_metadata` (JSON): A dictionary of path mappings, purposes, and distribution status (`INTERNAL` | `CLIENT`).
- `structure_verdict` (Boolean): Confirms if the current repo layout matches the `architectural-linter` rules.

## Preconditions
- Filesystem MUST be accessible with READ permissions.
- `architectural-linter` MUST be available for layout validation.

---

## Instructions

### 1. Domain Discovery
Identify the subtrees related to the `target_domain`:

| Domain | Root Path | Primary Purpose |
| :--- | :--- | :--- |
| **SOMBRA** | `apps/sombra/` | Agentic data gateway and PII redaction orchestrator. |
| **SLM_ENGINE**| `apps/slm-engine/` | Local AI inference engine for PII detection. |
| **DASHBOARD**| `apps/dashboard/` | Operational control plane and telemetry UI. |
| **PROXY** | `apps/proxy/` | Secure transparent proxy for LLM traffic. |
| **BRIDGE** | `apps/automation_bridge/` | CLI orchestration and documentation server. |
| **WEB** | `apps/web/` | Public-facing portal and marketing interface. |
| **CORE** | `docs//` | Technical specifications (TECH_DOCS.md) and setup guides. |
| **PILOT** | `docs/pilot/` | Onboarding materials and pilot playbooks. |
| **CONNECTORS**| `services/refinery/pkg/connector/` | Third-party ingestion logic. |

### 2. Metadata Enrichment
For each file in the domain, extract:
- `is_internal`: TRUE if path contains `internal`.
- `last_sync`: Timestamp of the last documentation update.
- `related_component`: Derived from the directory name.

### 3. Consistency Validation
- Run a scan to ensure that every `CORE` document has a corresponding entry in `TECH_DOCS.md`.
- Flag any undocumented directories in the project root.

## Failure Handling
- **Path Violation**: If a file is in the wrong directory (e.g., a report in `CORE`), return an `ARCHITECTURAL_DRIFT` error and block the caller.

## Examples

### Example: API Documentation Check
- **Input**: `target_domain` = `CORE`.
- **Action**: Map all files in `docs/`. Match `TECH_DOCS.md` against the refinery's current routes.

### Example: Connector Discovery
- **Input**: `target_domain` = `CONNECTORS`.
- **Action**: List all sub-packages and their README status.
