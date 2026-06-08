# 🚀 Deployment Guide

OCULTAR is designed to run anywhere you have Docker. This guide covers the two most common deployment scenarios.

## 1. Local Development / Pilot (Community)
*Target: Individual developers and trial evaluations.*

### Prerequisites
- Docker Desktop installed and running.
- Port 3030 (Dashboard) and 8086 (Sombra) available.

### Steps
1.  **Extract the package**: Unzip `ocultar-community.zip`.
2.  **Run the script**:
    - **Windows**: `scripts/setup-community.ps1`
    - **Mac/Linux**: `scripts/setup-community.sh`
3.  **Wait for the model**: The first run will pull a ~1.2GB AI model.
4.  **Access the Dashboard**: Go to `http://localhost:3030`.

---

## 2. Enterprise Deployment
*Target: Production environments and private clouds.*

### Prerequisites
- **Doppler account**: For managing production secrets.
- **Enterprise License Key**: Provided by the OCULTAR team.
- **Hardware**: Minimum 4 vCPUs and 8GB RAM (for local SLM inference).

### Configuration (The `.env` file)
The enterprise binary expects the following variables:
```bash
OCU_LICENSE_KEY="Your.Enterprise.Key"
OCU_MASTER_KEY="A-Long-Secure-Random-String"
OCU_VAULT_PATH="/var/lib/ocultar/vault.db"
SLM_SIDECAR_URL="http://slm-engine:8086"
```

### Steps
1.  **Setup Secrets**: Two options:
    - **Doppler** (recommended for production): `doppler run -- docker compose -f dist/ocultar-enterprise/docker-compose.yml up -d`
    - **`.env` file** (pilots / air-gapped): `cp .env.example .env`, fill in real values, then `docker compose -f dist/ocultar-enterprise/docker-compose.yml up -d`
2.  **Verify Status**: Check the health endpoint:
    ```bash
    curl http://localhost:8081/healthz
    ```

---

## 🏗️ Deployment Topology

### Sidecar Pattern (Recommended)
Deploy the `slm-engine` as a sidecar container to the `sombra` gateway. This ensures low latency for Tier 2 AI NER scans.

### High Availability
For HA deployments, run multiple instances of Sombra behind a load balancer. Ensure they all point to a shared, high-availability Vault (e.g., PostgreSQL for Enterprise).

> [!IMPORTANT]
> **Data Residency**: Because OCULTAR is a Zero-Egress solution, all data stays within the network segment where you deploy it. Ensure your firewall rules allow Sombra to reach your chosen upstream AI providers (OpenAI, Anthropic, etc.).
