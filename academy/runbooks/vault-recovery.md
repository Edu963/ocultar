# 🔐 Vault Backup & Recovery

The OCULTAR Vault stores the mapping between tokens and original PII. Losing this vault means tokens can never be re-hydrated (de-tokenized).

---

## 💾 Backup Strategy

### 1. DuckDB (Default)
DuckDB is a file-based database. Backing it up is as simple as copying the file.

- **Source**: The file specified in `OCU_VAULT_PATH` (default: `sombra_vault.db`).
- **Procedure**:
  ```bash
  # Ensure the process is not actively writing (or use DuckDB's checkpoint)
  cp sombra_vault.db /backups/vault_$(date +%F).db
  ```
- **Frequency**: Daily (Automated via CRON).

### 2. PostgreSQL (HA Deployments)
For HA deployments, use standard PostgreSQL backup tools.

- **Procedure**:
  ```bash
  pg_dump -h localhost -U ocultar -d vault_db > vault_backup.sql
  ```

---

## 🆘 Recovery Procedure

If the vault file is corrupted or deleted:

1.  **Stop Sombra**: `docker compose stop sombra`
2.  **Restore File**:
    - Move your latest backup to the `OCU_VAULT_PATH` location.
    - Ensure permissions are correct: `chmod 600 sombra_vault.db`.
3.  **Check Key Consistency**: Ensure the `OCU_MASTER_KEY` currently in use is the **same** key that was used when the backup was created.
4.  **Restart Sombra**: `docker compose start sombra`
5.  **Validation**:
    - Run `duckdb vault.db "SELECT count(*) FROM vault;"` to verify records exist.
    - Test a re-hydration query using a known token from the audit logs.

---

## ⚠️ Important Security Note
**Never store backups and the `OCU_MASTER_KEY` in the same location.** If an attacker gains access to both, the vault is essentially plaintext.
- Store `vault.db` in your standard backup infrastructure.
- Store `OCU_MASTER_KEY` in a secure HSM or Secret Manager (Doppler, HashiCorp Vault).
