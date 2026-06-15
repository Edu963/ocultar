# 🚀 Deployment Guide

OCULTAR is designed to run anywhere you have Docker. This guide covers the two most common deployment scenarios.

## 1. Local Development / Quick Start
*Target: Individual developers and trial evaluations.*

### Prerequisites
- Docker or Go 1.24+ with CGO enabled (GCC required).
- Port 4141 available.

### Steps
1.  **Clone the repo**: `git clone https://github.com/ocultar-dev/ocultar.git && cd ocultar`
2.  **Start the server**:
    ```bash
    export OCU_MASTER_KEY=$(openssl rand -hex 32)
    export OCU_SALT=$(openssl rand -hex 16)
    export OCU_AUDITOR_TOKEN=dev
    docker run -e OCU_MASTER_KEY -e OCU_SALT -e OCU_AUDITOR_TOKEN \
      -p 4141:4141 ghcr.io/ocultar-dev/ocultar:latest -serve 4141
    ```
3.  **Verify**: `curl http://localhost:4141/api/health`

---

## 2. Production Deployment
*Target: Production environments and private clouds.*

### Prerequisites
- A secure `.env` file or your preferred secret manager for secret injection.
- **Hardware**: Minimum 4 vCPUs and 8GB RAM (for local SLM inference).

### Configuration (The `.env` file)
```bash
OCU_MASTER_KEY="<output of: openssl rand -hex 32>"
OCU_SALT="<output of: openssl rand -hex 16>"
OCU_VAULT_PATH="/var/lib/ocultar/vault.db"
SLM_SIDECAR_URL="http://slm-engine:8085"
```

### Steps
1.  **Setup Secrets**: `cp .env.example .env`, fill in real values, then `docker compose up -d`
2.  **Verify Status**: Check the health endpoint:
    ```bash
    curl http://localhost:4141/api/health
    ```

---

## 🏗️ Deployment Topology

### Sidecar Pattern (Recommended)
Deploy the `slm-engine` as a sidecar container to the `sombra` gateway. This ensures low latency for Tier 2 AI NER scans.

### High Availability
For HA deployments, run multiple instances of Sombra behind a load balancer. Ensure they all point to a shared PostgreSQL vault (see the Advanced Setup Guide for configuration).

> [!IMPORTANT]
> **Data Residency**: Because OCULTAR is a Zero-Egress solution, all data stays within the network segment where you deploy it. Ensure your firewall rules allow Sombra to reach your chosen upstream AI providers (OpenAI, Anthropic, etc.).
