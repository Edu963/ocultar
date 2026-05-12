## Secret Management
Ocultar uses Doppler for secret management.

### Local development
Prerequisites: doppler CLI installed, authenticated, project set to ocultar/dev
Start: ./scripts/start.sh

### Secrets reference
| Secret | Description | Rotation |
|--------|-------------|---------- |
| OCU_MASTER_KEY | 32-byte AES key for vault encryption | Quarterly |
| OCU_SALT | 16-byte HKDF salt | With master key |
| OCU_JWT_SECRET | HS256 secret for Sombra Bearer token validation (enterprise) | Quarterly |
| OCU_LICENSE_KEY | Enterprise license | On expiry |
| OPENAI_API_KEY | OpenAI API access | Monthly recommended |
| GEMINI_API_KEY | Gemini API access | Monthly recommended |

### Key rotation procedure
1. Generate new key: openssl rand -hex 32
2. Update in Doppler: doppler secrets set OCU_MASTER_KEY=<new_value>
3. WARNING: rotating OCU_MASTER_KEY invalidates all existing vault tokens.
   All vaulted PII becomes unrecoverable. Do NOT rotate in production
   without a full vault migration plan.

### What happens if OCU_MASTER_KEY is lost
All encrypted vault data is permanently unrecoverable.
There is no backdoor. Treat this key like a database encryption key.
