---
name: shadow-api-scanner
description: Proactively detects endpoints exposed at the network level that are NOT registered in Sombra's policy.
---

# Shadow API Scanner (v1.1)

## Purpose

SAS prevents "Shadow AI" shadow deployments. It identifies listening ports that could be used to bypass the Sombra Gateway, ensuring every AI-related request is protected and audited.

## Inputs / Outputs

### Inputs
- `authorized_ports`: Port list from `sombra.yaml`.
- `scan_interface` (Enum): `LO` | `ETH0` | `ALL`.

### Outputs
- `clandestine_hits` (List): Process IDs and ports discovered.
- `verdict` (Enum): `CLEAN` | `SHADOW_DETECTED`.

## Preconditions
- Root/Sudo privileges for port-to-process attribution.

---

## Instructions

### 1. Listener Discovery
- Execute `ss -tunlp` (Preferred) or `netstat -tunlp`.
- **Filter**: Isolate listening TCP ports in common AI ranges (5000, 8000, 11434).

### 2. Config Reconciliation
- Cross-reference hits against `sombra.yaml`.
- **Verdict**: If a port is active but NOT in the authorized list, it is a `SHADOW_DETECTED` event.

### 3. Active Probing
- Perform a stealthy `GET /` on suspected ports.
- Identify "Mock OpenAI" or "Ollama" signatures.

## Failure Handling
- **`GHOST_LISTENER`**: If a port is open but no process is visible, flag as `TAMPER_SUSPECTED`.
- **`PROBE_BLOCKED`**: If the port is open but unreachable, assume a hardened shadow service exists.

## Postconditions
- Any `SHADOW_DETECTED` verdict MUST trigger an immediate `LOCKDOWN` in `continuous-ai-orchestrator`.
