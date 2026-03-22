# OCULTAR Enterprise Proxy — 1-Click Deployment Guide

Deploying the OCULTAR Enterprise Proxy is designed to be frictionless, taking less than 10 minutes to secure your AI pipeline.

## Prerequisites

- **OS**: Linux, macOS, or Windows (with WSL 2)
- **Docker**: Docker Refinery + Compose plugin (or Docker Desktop on Windows/macOS)
- No host-level Go, Python, or ML libraries required — everything runs in containers

---

## Phase 1: Environment Setup (2 Minutes)

**1. Unpack the archive**

If you received a distribution tarball:
```bash
tar -xzf ocultar-enterprise.tar.gz
cd ocultar-enterprise
```

If deploying from source:
```bash
git clone https://github.com/Edu963/ocultar.git
cd ocultar
```

**2. Create your environment file**

```bash
cp .env.example .env
```

**3. Configure your secrets**

Open `.env` and set the following values:

```bash
# REQUIRED — AES-256 master encryption key (min 32 chars).
# ⚠️  Never change this after first run: it invalidates all vault entries.
OCU_MASTER_KEY=$(openssl rand -hex 32)

# REQUIRED — Per-deployment HKDF salt.
# ⚠️  Also invalidates vault if changed after first run.
OCU_SALT=$(openssl rand -hex 16)

# REQUIRED — Your upstream LLM API URL (the service you want to protect).
# Examples: https://api.openai.com  |  http://your-internal-llm:11434
OCU_PROXY_TARGET=https://api.openai.com

# OPTIONAL — Proxy listener port (default: 8081).
OCU_PROXY_PORT=8081

# OPTIONAL — Activate Enterprise features (Tier 2 AI scan, PostgreSQL vault,
# SIEM audit log). Leave empty for Community feature set.
OCU_LICENSE_KEY=
```

> A one-liner that fills in the keys automatically:
> ```bash
> sed -i "s/replace-with-a-secure-32-byte-key/$(openssl rand -hex 32)/" .env
> sed -i "s/ocultar-kdf-salt-placeholder/$(openssl rand -hex 16)/" .env
> ```
> (On macOS replace `-i` with `-i ''`)

---

## Phase 2: Start the Cluster (4–5 Minutes)

```bash
docker compose -f docker-compose.proxy.yml up -d
```

On first run the cluster will:
1. Pull the **Qwen 1.5B Q4_K_M GGUF** model (~1.2 GB) from HuggingFace into a local volume.
2. Boot the `llama.cpp` inference refinery (`slm-ner` container) and wait for it to pass its health check.
3. Start the `ocultar-proxy` container with pre-flight validation of your master key and vault path.

Watch startup progress:
```bash
docker compose -f docker-compose.proxy.yml logs -f
```

Expected final output:
```
ocultar-proxy  | [+] All pre-flight checks passed! Starting Proxy.
ocultar-proxy  | [INFO] OCULTAR proxy listening on :8080
```

> **Subsequent starts are instant** — the model is cached in the `slm_data` Docker volume.

---

## Phase 3: Validation (1 Minute)

```bash
bash scripts/smoke_test.sh
```

Expected output:
```
[+] Proxy is healthy!
[*] Running smoke test with leaky payload...
[+] SUCCESS: PII successfully intercepted and redacted!
```

**Your applications should now point to `http://localhost:${OCU_PROXY_PORT:-8081}`** instead of the upstream API directly. All JSON payloads will be automatically scrubbed before forwarding.

---

## What Each Port Does

| Port | Service | Purpose |
|---|---|---|
| `${OCU_PROXY_PORT:-8081}` (host) | `ocultar-proxy` | Transparent PII proxy — point your app here |
| `8080` (internal only) | `slm-ner` | Local AI inference — not exposed to host |

> The Enterprise Dashboard (`/index_v2.html`) is part of the **standalone binary** deployment, not the proxy stack. See [`ENTERPRISE_SETUP_GUIDE.md`](../documentation/ENTERPRISE_SETUP_GUIDE.md) for the full feature walkthrough.

---

## Shut Down

```bash
docker compose -f docker-compose.proxy.yml down
```

Vault data is persisted in the `proxy_vault` Docker volume and survives restarts.

To wipe everything (including the vault):
```bash
docker compose -f docker-compose.proxy.yml down -v
```
---

## Maintenance & Integrity

### Key Rotation Strategy
The `OCU_MASTER_KEY` and `OCU_SALT` are the root of your Sovereign Vault's security.
* **⚠️ WARNING**: Changing either of these values after data has been vaulted will invalidate all existing tokens.
* **Fallback Behavior**: In the event of a key rotation, OCULTAR will not crash but will return **un-hydrated tokens** (e.g., `[PERSON_a1b2c3d4]`) instead of original PII. This prevents data leaks while alerting you to the key mismatch.
* **Best Practice**: Backup your `.env` file securely. If you must rotate keys, you should clear the vault (`rm vault.db`) to ensure fresh deterministic mappings.
### Encryption & Security Protocol
All Sovereign Data Objects (SDOs) are protected using industry-standard authenticated encryption:
1. **Key Derivation**: On boot, the `OCU_MASTER_KEY` and `OCU_SALT` are passed through **HKDF-SHA256** to derive the 32-byte AES refinery key. 
2. **Deterministic Hashing**: PII is hashed using **SHA-256** to create a unique vault index.
3. **Authenticated Encryption**: SDOs are encrypted using **AES-256-GCM**. Each record includes a unique cryptographically secure random nonce, ensuring that the same PII encrypted twice produces different ciphertexts.
4. **Data-at-Rest**: Only the hashed index and the GCM-protected ciphertext are stored in the vault. **Raw PII never touches the disk.**
