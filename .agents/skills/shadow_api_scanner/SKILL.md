---
name: shadow-api-scanner
description: Proactively detects endpoints exposed at the network level that are NOT registered in Sombra's policy. Prevents "Shadow AI" adoption and unmanaged data leaks.
---

# Shadow API Scanner (v1.0)

## Purpose

The Shadow API Scanner ensures that every service listening for AI-related requests is under the control of the Sombra Gateway. It identifies "rogue" models or endpoints that may have been deployed by developers or compromised systems without PII protection or audit logging.

## Preconditions

- **Tools**: `netstat` (or `ss`), `lsof`, `curl` (active spidering), `yq`.
- **Access**: Root/Sudo privileges (recommended) to see all listening ports. Read access to `apps/sombra/configs/sombra.yaml`.

## Inputs / Outputs

### Inputs
- `expected_ports` (List): Ports the Sombra Gateway expects to listen on (default: `["8081", "8082"]`).
- `scan_range` (String): IP range or interface to scan (default: `localhost`).
- `config_source` (Path): Path to the active `sombra.yaml`.

### Outputs
- `clandestine_endpoints` (List): Ports/Services found that are NOT in the `expected_ports` or `sombra.yaml`.
- `alert_level` (Enum): `NONE` | `CRITICAL_SHADOW_DETECTED`.

---

## Instructions

### Step 1 – Listener Discovery
1.  Run a network-level scan to find all listening TCP ports.
    - Command: `netstat -tunlp | grep LISTEN`.
2.  Filter results for services that look like AI APIs (e.g., listening on common ports like `8000`, `5000`, `11434` for Ollama, etc.).

### Step 2 – Configuration Parity Check
1.  Parse `sombra.yaml` to identify all "Authorized Endpoints" (Connectors and Models).
2.  Extract the `listen_port` for the Gateway and the `endpoint` for each Model.
3.  **Shadow Check**:
    - For every active port found in Step 1, verify if it is an "Authorized Endpoint".
    - **Drift**: If a port is active but NOT authorized, flag as **CRITICAL_SHADOW_DETECTED**.

### Step 3 – Active Probing (Spider)
1.  Attempt a lightweight `GET /health` or `GET /v1/models` on any detected clandestine endpoints.
2.  Capture the `Server` header or response body to identify the model type (e.g., llama.cpp, vLLM, OpenAI-compatible).

### Step 4 – Reporting
1.  Consolidate all unauthorized services into the `clandestine_endpoints` list.
2.  **Action Required**: If `alert_level == CRITICAL`, notify the `CISO & Compliance Officer` immediately to shut down the unmanaged service.

---

## Failure Handling

- **Tools Missing**: If `netstat` is not available, try `ss -tlpn`.
- **Permission Denied**: If ports cannot be attributed to a process, mark as a "Ghost Listener" and request elevated privileges.

---

## Ecosystem Role
- **Category**: Discovery / Auditor.
- **Dependencies**: `drift-detector`, `repository-knowledge-map`.
- **Triggers**: Weekly security sweep or manual audit.
