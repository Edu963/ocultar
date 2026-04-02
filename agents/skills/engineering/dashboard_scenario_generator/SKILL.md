---
name: dashboard-scenario-generator
description: Demo Orchestrator for the Ocultar Enterprise Dashboard. Provisions the Mock-API with synthetic "Attack & Defense" scenarios.
---

# Dashboard Scenario Generator

## Purpose
To demonstrate the value of OCULTAR without needing live sensitive data. This skill allows Sales and Refineryering teams to "inject" specific security scenarios into the dashboard to visualize how the system reacts to threats.

## Role
- **Category**: Orchestrator / Simulation
- **Ecosystem Positioning**: Sales & Demos

## Instructions

### 1. Define Scenario
- **Finance Breach**: Simulate 10,000 IBAN leaks blocked by the Refinery.
- **Healthcare Leak**: Simulate patient records being intercepted at the Sombra Gateway.
- **M&A Leak**: Simulate internal project codenames (e.g., "Project Nightshade") being filtered from a chatbot.
- **SLM Timeout**: Simulate an AI inference delay (>5s) to trigger fail-closed redaction.
- **Queue Saturation**: Simulate 500 concurrent requests to verify throttle behavior.

### 2. Provision Mock-API
- Call the `Mock-API` endpoint: `POST /api/scenario`.
- Payload: `{"scenario": "finance", "scale": "enterprise", "blocked_pii": 15000}`.
- Goal: Update the dashboard's "ROI" and "Risk Matrix" instantly.

### 3. Synthetic Log Generation
- (Optional) Append synthetic entries to a temporary `demo_audit.log` to allow for deep-dive investigations during the demo.

## Preconditions
- `services/mock-api` must be running.
- Dashboard must be in `DEMO` mode.
