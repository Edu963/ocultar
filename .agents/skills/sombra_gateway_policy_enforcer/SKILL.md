---
name: sombra-gateway-policy-enforcer
description: Automatically injects policy hooks when new endpoints or features are added. Ensures all AI requests go through the Fail-Closed logic.
---

# Sombra Gateway Policy Enforcer

## Purpose

This skill acts as an architectural validator for the Sombra Gateway. It ensures that no data-handling path bypasses the Ocultar Refinery or violates the Fail-Closed principle.

## When To Use This Skill

Use this skill when:
- Adding new HTTP routes to Sombra (`/query`, `/stream`, etc.).
- Implementing new Model Connectors or Provider wrappers.
- Modifying or adding modular OCULTAR Pro Connectors (Slack, SharePoint, etc.).
- Modifying the primary gateway handler in `apps/sombra/pkg/handler/`.
- Changing the policy enforcement middleware.

## Instructions

1.  **Identify Data Entry Points**: Locate all methods where external AI requests enter the system.
2.  **Verify Refinery Hook**: Ensure that `eng.ProcessInterface` or `eng.RefineString` is called before any data is sent to an upstream provider.
3.  **Validate Fail-Closed Logic**: Confirm that if the Refinery returns an error, the request is blocked and an appropriate security event is logged.
4.  **Vault Alignment**: Verify that Sombra shares the same `vault.db` as the core engine (at `services/engine/`) to enable re-hydration.
5.  **Audit Configuration**: Verify that `sombra.yaml` correctly mapping models to their required security policies.
6.  **Enforce Multi-Model Consistency**: Ensure that new model providers follow the same redaction-hydration lifecycle as existing ones (OpenAI, Gemini).
7.  **Validate Connector Boundaries**: Confirm that Pro Connectors (SharePoint, Teams) do not leak ingestion logic into the core engine and strictly implement the `Connector` interface.

## Examples

### Adding a new `/stream` endpoint
**Action**: Verify that the streaming chunk handler invokes the refinery and that the stream is severed immediately if a security violation is detected.
