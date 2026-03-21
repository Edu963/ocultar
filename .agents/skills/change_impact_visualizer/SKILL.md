---
name: change-impact-visualizer
description: Deterministic analysis of code changes to map architectural, security, and regulatory impacts. Step 12 of the Ocultar 16-Step Protocol.
---

# Change Impact Visualizer (Step 12)

## Purpose
This skill provides a high-fidelity "Impact Narrative" for every modification. It ensures that engineers and auditors understand the ripple effects of a change across the Ocultar ecosystem by cross-referencing `ARCHITECTURE.md` and `PII_DETECTION.md`.

## Inputs
- **Primary**: `git diff` (staged or unstaged changes).
- **Secondary**: `docs/ARCHITECTURE.md` (Package Dependency Graph, C4 Container Diagram).
- **Tertiary**: `docs/PII_DETECTION.md` (Token Type Compliance Mappings).

## Outputs
- **Impact Matrix**: A structured Markdown table summarizing impacts.
- **Risk Assessment**: A list of regulatory and doc-drift risks.
- **Verification Note**: Confirmation that all changed files are analyzed.

## Preconditions
1. The technical task is marked as "Completed" or "In Review".
2. `docs/ARCHITECTURE.md` and `docs/PII_DETECTION.md` are present and readable.
3. Access to `git` tools is available.

## Instructions

### 1. Architectural Component Mapping
- Identify all modified files from the `git diff`.
- Use the **Container Diagram** in `docs/ARCHITECTURE.md` to map each file to an Ocultar component:
    - `pkg/engine` → **Core Engine** (High Impact)
    - `pkg/proxy` → **Sombra Gateway** (High Impact)
    - `pkg/vault` → **Identity Vault** (Critical Impact)
    - `docs/` → **Documentation Cluster**
    - `dist/` → **Distribution Artifacts**
    - `apps/web` → **Enterprise Dashboard**

### 2. Regulatory & PII Traceability
- If changes occur in `pkg/engine` (Redaction tiers) or `pkg/vault` (Storage):
    - Locate any new or modified regex/logic patterns.
    - Reference `docs/PII_DETECTION.md` (Section 2) to determine the **Compliance Requirement** (e.g., GDPR Art 4, HIPAA, PCI-DSS).
    - Flag if a change to a detection Tier (0-3) alters the "Deterministic Pseudonymization" guarantee.

### 3. Doc Drift & Stale Logic Detection
- Compare logic changes in `pkg/` against the **Refinery Pipeline** sequence in `docs/ARCHITECTURE.md`.
- **Condition**: If the data flow (e.g., adding a new Evasion Shield tier) has changed, flag `docs/ARCHITECTURE.md` as "Required Update."
- Check if `PII_DETECTION.md` requires updates due to new `[TOKEN_TYPE]` additions.

### 4. Distribution Impact Assessment
- Determine if the change alters the contents of the final release (refer to `docs/RELEASE_GUIDE.md` if available).
- **Rules**:
    - Binary code changes → Impact **Community/Enterprise Builds**.
    - Config schema changes → Impact **Onboarding Guides**.

### 5. Final Output Generation
- Create an **Impact Matrix** following this format:

| File(s) | Component | Primary Impact | Regulatory/Risk |
|---|---|---|---|
| `engine.go` | **Core Engine** | Hardened AES-GCM rotation | GDPR (Data at Rest) |
| `PII.md` | **Docs** | Updated PII Mapping | Auditor Transparency |

## Failure Handling
- **Missing Architecture**: If `ARCHITECTURE.md` is missing, use absolute file paths and flag as a "Critical Doc Failure."
- **Ambiguous Change**: If the `git diff` contains >50 modified files, categorize by directory and focus on "High Impact" components first.

## Validation Step
- **Integrity Check**: Ensure every file listed in `git diff --name-only` is represented in the final Impact Matrix. If missing, append an "Uncategorized Impact" section.
