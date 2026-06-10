# 🔐 Security & Trust Model

OCULTAR is designed for high-security environments where data leakage to third-party AI providers is a critical risk. Our security model provides verifiable guarantees through multiple layers of protection.

---

## 1. Zero-Trust for Data
The fundamental principle of OCULTAR is that **untrusted providers should never see plaintext PII.**
- **Isolation**: The refinery and vault run entirely within your infrastructure (on-prem or private cloud).
- **No External Detection**: Unlike other solutions, OCULTAR does *not* call external APIs (like Google DLP or Amazon Macie) to detect PII. Detection is 100% local.

---

## 2. Secure Vaulting
When PII is detected, it is replaced with a token and stored in a secure vault.
- **Encryption**: Data is encrypted using **AES-256-GCM**.
- **Key Derivation**: The encryption key is never stored on disk. It is derived at runtime from the `OCU_MASTER_KEY` using **HKDF-SHA256** with a configurable salt.
- **Persistence**: The vault (DuckDB) stores the encrypted blobs. The master key remains only in memory.

---

## 3. SSRF Protection
Sombra acts as a hardened proxy to prevent Server-Side Request Forgery (SSRF) attacks.
- **Domain Whitelisting**: Sombra only allows requests to a predefined list of trusted AI provider domains:
  - `api.openai.com`
  - `api.anthropic.com`
  - `api.mistral.ai`
  - `generativelanguage.googleapis.com`
- **Network Validation**: Internal logic blocks RFC 1918 (private IPs) and cloud metadata services (`169.254.169.254`) to prevent lateral movement within your network.

---

## 4. Fail-Closed Mechanics
OCULTAR is designed to "fail loudly" rather than "degrade gracefully."
- **Total Block**: If the Refinery service is unavailable or the Vault fails to initialize, Sombra will return a `500 Internal Server Error` and block the outgoing request.
- **No Plaintext Leakage**: There is no "bypass" mode. If the system cannot guarantee the removal of PII, the data does not leave the boundary.

---

## 5. Tamper-Proof Audit Logs
Every security-relevant event is logged in a way that is difficult to forge.
- **Immutable Logger**: Uses hash-chained entries.
- **Signing**: Logs can be signed using **Ed25519** to ensure integrity.
- **Traceability**: Audit logs track:
  - Which rule matched.
  - When a vault entry was created.
  - Which upstream provider was called.

---

## 6. Deterministic Tokenization
Tokens are generated using `SHA-256(original_PII)`.
- **Consistency**: The same PII always produces the same token.
- **Safe Analytics**: This allows data teams to perform joins, counts, and frequency analysis on tokenized data without ever needing to de-tokenize it.
- **Example**: `John Doe` -> `token_abc123`. If `John Doe` appears in 10 different documents, all will have `token_abc123`.
