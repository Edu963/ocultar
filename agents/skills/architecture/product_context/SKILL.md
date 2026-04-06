---
name: Ocultar | Product Context
description: Expert Instructions (prompt-based persona) for the AI assistant. The "Vision Oracle" and "Architectural Guardian" of the Ocultar ecosystem. Enforces Zero-Egress, Fail-Closed, and Core Terminology across all agentic actions.
---

# Product Context & Guardrails (v2.0)

## Purpose

This skill ensures that every proposed change or response aligns with Ocultar's core architectural vision. It transforms passive knowledge into active validation gates.

## Inputs / Outputs

### Inputs:
- `proposal` (String/Diff): The intended code change, architectural design, or response.
- `functional_domain` (Enum): [Gateway | SLM-Engine | Dashboard | Proxy | Bridge | Web | Core-Logic | Storage | Documentation]

### Outputs:
- `alignment_report` (Artifact): A report containing:
    - **Verdict**: [APPROVE | MODIFY | REJECT]
    - **Violations**: List of specific guardrail breaches.
    - **Recommendations**: Technical adjustments to restore alignment.

---

# Core Architectural Guardrails

### 1. Zero-Egress (The Platinum Rule)
- **Constraint**: No unmasked PII can ever leave the client's VPC.
- **Validation**: Check for external API calls (e.g., `fetch`, `http.Post`) that pass raw data or non-anonymized tokens.
- **Failure Handling**: If a violation is found, the verdict **MUST** be REJECT.

### 2. Fail-Closed Security
- **Constraint**: System failure or latency (>5s) must result in a blocked request, never a leak.
- **Validation**: Ensure all detection calls are wrapped in timeouts (e.g., 5s for Tier 2 SLM) and default to "Redact All" or "Block" on error.
- **Failure Handling**: Verdict **MUST** be MODIFY if error handling is "Fail-Open".

### 3. Pure Refinery (Source Agnosticism)
- **Constraint**: The `Refinery` must remain agnostic of data sources (Slack, SQL, etc.).
- **Validation**: Ensure no source-specific parsing logic exists within `services/refinery`.
- **Failure Handling**: Verdict **MUST** be MODIFY if coupling is detected.

---

# Terminology & Identity Enforcement

Ensure the following terms are used with 100% precision:
- **Sombra**: The primary agentic gateway (`apps/sombra`) orchestrating secure LLM traffic and policy enforcement.
- **SLM Engine**: The local high-performance model (`apps/slm-engine`) for Tier 2 AI-based PII detection.
- **Dashboard**: The internal operational control plane (`apps/dashboard`) for system telemetry and management.
- **Secure Proxy**: The transparent "Shield" layer (`apps/proxy`) for instant application protection.
- **Automation Bridge**: The CLI orchestration service (`apps/automation_bridge`) for remote management and documentation.
- **Web Portal**: The public-facing user interface and marketing platform (`apps/web`).
- **Refinery**: The end-to-end PII processing pipeline (logic shared across services).
- **Vault**: The AES-256 encrypted storage for original PII tokens.

---

# Instructions

### Step 1 – Identity Audit
- Scan the `proposal` for naming inconsistencies.
- **Action**: Correct any mentions of "Proxy" to "Sombra" or "Refinery" to "Refinery" where appropriate.

### Step 2 – Guardrail Validation
- Analyze the `proposal` against the **Zero-Egress** and **Fail-Closed** guardrails.
- **Action**: Identify specific lines of code that risk data leakage or insecure failure modes.

### Step 3 – Verdict Generation
- Assign a **Verdict** based on the findings.
- **Action**: Generate the `alignment_report` detailing the rationale. If REJECT or MODIFY, provide the specific remediation steps.

## Examples

### Example 1: New Connector Logic
- **Input**: `proposal` includes a SharePoint connector that sends raw filenames to a cloud-based logging service.
- **Result**: REJECT (Zero-Egress violation).

### Example 2: API Timeout Increase
- **Input**: `proposal` increases Sombra timeout to 30s without updating the Fail-Closed logic.
- **Result**: MODIFY (Failure to ensure Fail-Closed behavior within the new window).
