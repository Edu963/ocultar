# OpenAI shipped a model. We built the system.

> **Audience:** Senior security engineers and DevOps leads evaluating PII infrastructure for AI-connected systems.

Three events in the past eight weeks drew a line around a specific class of infrastructure risk. Taken together, they validate a thesis we have been operating on since day one: PII that leaves your trust boundary is a liability, and any system that lets it do so — intentionally or not — is a breach waiting to happen.

---

## Event 1: The supply chain vector (March 2026)

A prominent AI startup suffered a breach through an open-source pipeline tool positioned between their internal data stores and an external AI provider. The attacker did not compromise the AI provider, the data store, or the application layer. They compromised the connector.

40,000+ PII records exfiltrated. The entry point was a third-party tool doing exactly what it was designed to do: passing data along.

This is the attack surface specific to AI adoption — the middleware. Every tool that touches plaintext PII in transit between your systems and an AI provider is a node in a trust chain you did not design and cannot fully audit. The breach was not a failure of the upstream. It was a failure of the layer in between.

## Event 2: Regulated infrastructure is not immune (April 15, 2026)

A national government agency responsible for managing citizen identity documents confirmed a breach: 19 million records. Names, dates of birth, addresses, phone numbers. A GDPR-regulated entity with mature security controls, full compliance posture, audited annually.

The scale matters less than what it implies. The cost function of an identity breach is not linear with record count. At this scale, you are not failing four audits — you are failing a generation of trust in the underlying digital infrastructure.

The shared failure mode between events 1 and 2: plaintext PII was accessible at rest or in transit in a recoverable form. The breach surface was the infrastructure layer, not the application.

## Event 3: Market thesis, validated (April 22, 2026)

A major AI provider released a 1.5B-parameter PII detection model under Apache 2.0. Local inference. No API call. No data leaves the host.

When an organization of that scale ships a local-first PII detection model under a permissive open-source license, they are making a statement about where the industry is heading. Local detection is not a niche compliance requirement. It is the correct architecture.

But a model is not a system.

---

## An engine without a car

The released model identifies PII. It does not:

- Intercept requests before they reach an upstream API
- Tokenize PII deterministically so the same entity produces the same token across all documents and sessions
- Vault the ciphertext under AES-256-GCM with HKDF-derived keys that never leave process memory
- Emit a structured, tamper-evident audit trail of every vault event
- Enforce RBAC-gated re-hydration of tokens back to plaintext on the response path
- Detect obfuscation: Base64-encoded PII, URL-encoded payloads, PII embedded in email greetings and signatures

A detection model answers: *is there PII here?* A system answers: *what happens to it, and can we prove it?*

---

## Zero-egress eliminates the supply chain attack surface

The architecture that prevents the March breach is not better authentication on the middleware. It is removing plaintext PII from the middleware entirely. If the connector never holds plaintext PII, there is nothing for the attacker to exfiltrate.

```yaml
services:
  ocultar-proxy:          # intercepts all traffic to the upstream LLM API
    environment:
      - OCU_PROXY_TARGET=https://api.your-llm-provider.com
      - SLM_SIDECAR_URL=http://slm-engine:8085

  slm-engine:             # Tier 2 AI NER — local, no network egress
    environment:
      - SLM_ENGINE=privacy-filter
      - PRIVACY_FILTER_URL=http://privacy-filter-svc:8086

  privacy-filter-svc:     # openai/privacy-filter, Apache 2.0, 1.5B params
    build: ./apps/slm-engine/privacy_filter_server
```

Every request passes through the proxy. The refinery tokenizes all PII before the payload reaches the upstream. The upstream API sees tokens, not names or account numbers.

```bash
curl -s -X POST http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{
      "role": "user",
      "content": "Process the application for Jean-Pierre Dumont,
                  DOB 1978-04-12, tel +33 6 12 34 56 78,
                  IBAN FR76 3000 6000 0112 3456 7890 189"
    }]
  }'
```

Payload forwarded to the upstream:

```json
{
  "messages": [{
    "role": "user",
    "content": "Process the application for [PERSON_a3f1c2d4],
                DOB [DATE_9b2e7f1a], tel [PHONE_4d8c3b2e],
                IBAN [IBAN_c7f2a109]"
  }]
}
```

The upstream never sees the name, date, phone number, or bank account. There is nothing on the wire to intercept.

---

## The detection pipeline

Tokenization is not a single step. It is a nine-tier defense-in-depth pipeline that runs before every upstream call:

| Tier | Shield | What it catches |
|------|--------|-----------------|
| 0.1 | Base64 Evasion | PII hidden inside encoded blobs — decoded, scanned, re-encoded |
| 0 | Dictionary | Named entities: VIPs, internal projects, sensitive org names |
| 0.5 | Pattern + Entropy | High-entropy strings; Shannon scoring for keys and tokens |
| 1 | Rule Engine | EMAIL, SSN, IBAN (MOD97), CC (Luhn mod-10), 30+ national ID types |
| 1.1 | Phone Shield | libphonenumber validation — reduces false positives on digit sequences |
| 1.2 | Address Shield | Heuristic street address parser across EN/FR/ES/DE |
| 1.5 | Greeting/Signature | "Regards, Jean-Pierre" and "My name is…" detection |
| **2** | **AI NER** | **openai/privacy-filter — local token classifier, Apache 2.0** |
| 3 | Structural Heuristics | Proximity expansion: `[TOKEN] ET Dupont` → re-tokenized as single entity |

Tiers 0–1.5 are deterministic and require no model. Tier 2 now runs `openai/privacy-filter` — the best available open-weight PII detection model — locally, with no network call. Tier 3 catches what structural context reveals after the AI pass.

---

## The audit trail

Every vault event is written as a structured JSON line:

```json
{"timestamp":"2026-04-22T09:14:03Z","actor":"10.0.1.44","action":"vaulted","token":"[PERSON_a3f1c2d4]","regulation":"GDPR_Art4"}
{"timestamp":"2026-04-22T09:14:03Z","actor":"10.0.1.44","action":"vaulted","token":"[IBAN_c7f2a109]","regulation":"PCI_DSS"}
{"timestamp":"2026-04-22T09:14:07Z","actor":"10.0.1.44","action":"matched","token":"[PERSON_a3f1c2d4]","regulation":"GDPR_Art4"}
```

The token is what is logged. Never the original PII. The audit trail satisfies GDPR Article 32(1)(d) without creating a secondary exposure surface. It is signed with Ed25519 and exportable to any SIEM via Filebeat or Fluent Bit.

---

## What the three events mean, together

Event 1 showed that the middleware is the attack surface. Event 2 showed that compliance posture does not substitute for architectural controls. Event 3 showed that the industry has concluded local-first detection is correct — but shipped a model, not a deployable system.

The gap is between a detection capability and an enterprise-deployable, auditable, zero-egress data pipeline. The proxy intercepts. The refinery detects across nine tiers, with `openai/privacy-filter` now as the AI backbone. The vault encrypts under keys that never leave memory. The audit log proves every action. The upstream sees nothing it should not see.

That is the system.

---

*Zero-egress architecture. Self-hosted. Docker Compose in under five minutes.*
*Community tier: Apache 2.0. Enterprise tier: PostgreSQL HA vault, SIEM export, Ed25519 audit signing.*

---

## HN Submission

**Title / comment (280 characters):**

> OpenAI's privacy-filter model validates local-first PII detection — but a model isn't a system. No vault, no proxy, no audit trail. Two recent breaches show why the middleware layer is the attack surface. Here's the full zero-egress architecture. (self-hosted, Docker Compose)
