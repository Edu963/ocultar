# 🚀 Deployment Guide

OCULTAR is designed to run anywhere you have Docker. This guide covers the two most common deployment scenarios.

## 1. Local Development / Quick Start
*Target: Individual developers and trial evaluations.*

### Prerequisites
- Docker Desktop installed and running.
- Port 3030 (Dashboard) and 8086 (Sombra) available.

### Steps
1.  **Clone the repo**: `git clone https://github.com/Edu963/ocultar.git && cd ocultar`
2.  **Run the setup script**:
    - **Windows**: `scripts/setup-community.ps1`
    - **Mac/Linux**: `scripts/setup-community.sh`
3.  **Wait for the model**: The first run will pull a ~1.2GB AI model.
4.  **Access the Dashboard**: Go to `http://localhost:3030`.

---

## 2. Production Deployment
*Target: Production environments and private clouds.*

### Prerequisites
- **Doppler account** (recommended) or a secure `.env` file for secret injection.
- **Hardware**: Minimum 4 vCPUs and 8GB RAM (for local SLM inference).

### Configuration (The `.env` file)
```bash
OCU_MASTER_KEY="<output of: openssl rand -hex 32>"
OCU_SALT="<output of: openssl rand -hex 16>"
OCU_VAULT_PATH="/var/lib/ocultar/vault.db"
SLM_SIDECAR_URL="http://slm-engine:8086"
```

### Steps
1.  **Setup Secrets**: Two options:
    - **Doppler** (recommended for production): `doppler run -- docker compose up -d`
    - **`.env` file** (air-gapped): `cp .env.example .env`, fill in real values, then `docker compose up -d`
2.  **Verify Status**: Check the health endpoint:
    ```bash
    curl http://localhost:8081/healthz
    ```

---

## 🏗️ Deployment Topology

### Sidecar Pattern (Recommended)
Deploy the `slm-engine` as a sidecar container to the `sombra` gateway. This ensures low latency for Tier 2 AI NER scans.

### High Availability
For HA deployments, run multiple instances of Sombra behind a load balancer. Ensure they all point to a shared PostgreSQL vault (see the Advanced Setup Guide for configuration).

> [!IMPORTANT]
> **Data Residency**: Because OCULTAR is a Zero-Egress solution, all data stays within the network segment where you deploy it. Ensure your firewall rules allow Sombra to reach your chosen upstream AI providers (OpenAI, Anthropic, etc.).
