# 🔐 Key Rotation Runbook

Key rotation is a critical security procedure. In OCULTAR, the `OCU_MASTER_KEY` is used to derive the AES-256-GCM key that encrypts all PII stored in the vault.

> [!CAUTION]
> **Destructive Process**: The current version of OCULTAR does not support automatic in-place re-encryption. Changing the `OCU_MASTER_KEY` will make all existing vault entries unreadable.

---

## Scenario A: Regular Rotation (Planned)
If you are rotating keys as part of a regular schedule and can afford to reset the vault.

1.  **Notify Stakeholders**: Inform data teams that vault tokens will be reset.
2.  **Backup Old Vault**: Save a copy of the current `vault.db`.
3.  **Generate New Key**: Create a new secure random string (at least 32 characters).
4.  **Update Configuration**:
    - Update the `OCU_MASTER_KEY` in Doppler or your `.env` file.
5.  **Wipe the Vault**: Delete or move the `vault.db` file.
6.  **Restart Sombra**:
    ```bash
    docker compose restart sombra
    ```
7.  **Verify**: Send a test query to ensure new tokens are being generated and stored.

---

## Scenario B: Suspected Compromise (Emergency)
1.  **Immediate Stop**: Shut down all OCULTAR services.
    ```bash
    docker compose down
    ```
2.  **Revoke Old Key**: Mark the old key as compromised in your secret management system.
3.  **Deploy New Infrastructure**:
    - New `OCU_MASTER_KEY`.
    - New `OCU_SALT` (to change the HKDF derivation).
    - Fresh `vault.db`.
4.  **Restart**: Follow the standard deployment guide.

---

## 🏗️ Future: Seamless Re-encryption
We are working on a `vault-rotate` utility that will:
1.  Read the vault with the `OLD_KEY`.
2.  Decrypt each entry.
3.  Re-encrypt with the `NEW_KEY`.
4.  Write back to the vault.

Until then, key rotation must be treated as a "reset" event for the vault tokens.
