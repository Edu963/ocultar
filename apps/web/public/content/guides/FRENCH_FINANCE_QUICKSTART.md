# OCULTAR | French Finance Sovereign AI — 5-Minute Quickstart

**Audience:** CTOs, lead engineers, and DevOps teams at French financial institutions.  
**Goal:** Deploy a zero-egress AI proxy with GDPR Art. 25 + CNIL + PCI-DSS policies active in under 5 minutes.

> For the compliance narrative (GDPR mapping, CNIL specifics, DPO sign-off checklist), see [`GDPR_FRENCH_FINANCE.md`](../compliance/GDPR_FRENCH_FINANCE.md).

---

## What You Get

| Capability | Detail |
|---|---|
| **Zero-egress PII proxy** | Intercepts every AI API call. PII is tokenized locally before leaving your perimeter. |
| **French finance entity coverage** | IBAN (MOD97), NIR (French SSN), SIRET/SIREN (Luhn), EU VAT, BIC/SWIFT, credit cards (Luhn) — all validated, not just pattern-matched. |
| **Policy-as-code governance** | Pre-configured rules block credentials and GDPR Art. 9 special-category data; redact financial identifiers. |
| **Audit trail** | Every tokenization and policy block is logged. The compliance evidence endpoint gives your DPO a machine-readable snapshot. |
| **Local vault** | AES-256-GCM encrypted. Tokens are deterministic: same IBAN → same token across requests, enabling relational analytics on pseudonymized data. |

---

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (or Docker Engine + Compose v2)
- No Go toolchain, no API keys, no internet access required after first pull.

---

## Step 1 — Clone

```bash
git clone https://github.com/Edu963/ocultar
cd ocultar
cp .env.example .env
```

The `.env.example` ships with safe demo keys. For production, set `OCU_MASTER_KEY` and `OCU_SALT` to high-entropy values before connecting real data.

---

## Step 2 — Start the French Finance Stack

```bash
docker compose -f docker-compose.yml -f docker-compose.french-finance.yml up --build
```

First build takes ~5 minutes (Go + DuckDB/CGO compilation). Every subsequent start is instant.

**What's running:**

| Service | Port | Role |
|---|---|---|
| `ocultar-proxy` | 8081 | OpenAI-compatible PII proxy (intercepts AI API calls) |
| `ocultar-refinery` | 8080 | Refinery API + compliance evidence endpoint |
| `echo-upstream` | 8082 | Mock AI API (reflects request back for demo purposes) |

---

## Step 3 — Run the Demo

```bash
bash scripts/demo_french_finance.sh
```

This runs four tests:

1. **Proxy redaction** — sends a French banking payload through the proxy and shows tokenized output.
2. **Refinery API** — calls `/api/refine` directly and shows the detection report (entity type, tier, confidence).
3. **Policy block** — sends a payload containing an API key. Expects `HTTP 403` — credentials are blocked before they can reach any model.
4. **Compliance evidence** — pulls `GET /api/compliance/evidence` and shows the DPO-ready snapshot.

### Sample Payload

```
Virement de 4 500 EUR depuis le compte de Jean Dupont
(NIR: 1 85 06 75 115 423 18)
vers IBAN FR76 3000 6000 0112 3456 7890 189.
Société émettrice: Dupont & Associés SARL
SIRET 552 100 554 03333, TVA FR83552100554
Contact: jean.dupont@bpce.fr, tél. +33 6 12 34 56 78
```

### Expected Redacted Output

```
Virement de 4 500 EUR depuis le compte de [PERSON_a1b2c3d4]
(NIR: [FR_NIR_9f8e7d6c])
vers [IBAN_12ab34cd].
Société émettrice: [PERSON_5e4f3a2b]
[FRANCE_SIRET_NUMBER_7c8d9e0f], TVA [EU_VAT_3a4b5c6d]
Contact: [EMAIL_2b3c4d5e], tél. [PHONE_6f7a8b9c]
```

No original values reach the echo upstream or any external system.

---

## Step 4 — Pull the Compliance Evidence (for your DPO)

```bash
curl -s http://localhost:8080/api/compliance/evidence | python3 -m json.tool
```

Returns a JSON snapshot with:
- Vault entry count (unique PII tokens stored)
- Active policy list (what governance rules are enforced)
- Tier coverage (which detection layers are active)
- Last 10 audit log entries

This response can be polled by SOC 2 / ISO 27001 tools (Vanta, Drata, Secureframe) for automated evidence collection.

---

## Active Governance Policies

The French finance config (`configs/config.french-finance.yaml`) ships with four pre-wired policies:

| Policy | Entities | Action | Regulatory basis |
|---|---|---|---|
| `block-credentials-always` | CREDENTIAL, SECRET, AWS_KEY, AWS_SECRET | **Block (403)** | OWASP, PCI-DSS |
| `block-special-category-data` | HEALTH_ENTITY, SENSITIVE_EVENT | **Block (403)** | GDPR Art. 9, CNIL |
| `redact-financial-identifiers` | IBAN, CREDIT_CARD, SIRET, SIREN, EU_VAT, BIC | Redact | PCI-DSS, GDPR Art. 4 |
| `redact-national-ids` | FR_NIR, SSN, UK_NINO, DE_STEUER_ID, ES_DNI | Redact | GDPR Art. 9, CNIL |

To add or modify policies, edit `configs/config.french-finance.yaml` and restart.

---

## Integrating with Your AI Stack

Change your application's LLM base URL:

```
Before: https://api.openai.com
After:  http://localhost:8081
```

No other code change required. OCULTAR preserves all headers, paths, and query parameters. It is fully compatible with the OpenAI SDK, LangChain, LlamaIndex, and any library that supports a base URL override.

---

## Shut Down

```bash
docker compose -f docker-compose.yml -f docker-compose.french-finance.yml down
```

The vault persists in the `vault-data` Docker volume across restarts. To wipe the vault entirely:

```bash
docker compose -f docker-compose.yml -f docker-compose.french-finance.yml down -v
```

---

## Next Steps

- Read the **[DPO Compliance Narrative](../compliance/GDPR_FRENCH_FINANCE.md)** for GDPR Art. 25/32 mapping, CNIL specifics, and the technical evidence your legal team needs.
- Configure a real upstream: set `OCU_PROXY_TARGET=https://api.openai.com` (or your Mistral/internal endpoint) in `.env`.
- Enable Tier 2 AI NER for contextual entity detection: `docker compose -f docker-compose.yml -f docker-compose.french-finance.yml --profile ai up --build`
