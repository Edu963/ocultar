# 📖 Operational Runbooks

These runbooks provide step-by-step procedures for critical operational tasks. Follow them carefully to ensure the integrity and security of the OCULTAR system.

---

## 🛡️ Security Runbooks

### 1. [Key Rotation](./key-rotation.md)
*Frequency: Every 90 days or upon suspected compromise.*
How to update the `OCU_MASTER_KEY` and re-encrypt the vault.

### 2. [Vault Backup & Recovery](./vault-recovery.md)
*Frequency: Daily backups.*
How to securely back up the DuckDB/PostgreSQL vault and restore it after a failure.

---

## ⚙️ Maintenance Runbooks

### 3. [Scaling OCULTAR](./scaling.md)
*Trigger: High latency or CPU saturation.*
How to horizontal scale Sombra and SLM Engine instances.

### 4. [Updating AI Models](./updating-models.md)
*Trigger: New model release.*
How to swap the GGUF model in the `slm-engine` without downtime.

---

## 🚀 Emergency Runbooks

### 5. [Incident Response: Potential Data Leak](../playbooks/incident-response.md)
What to do if you suspect PII has bypassed the refinery.
