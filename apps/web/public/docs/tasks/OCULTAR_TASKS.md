# OCULTAR — Task List
> Revised 2026-04-26. Replaces ocultar_nous_antigravity_v2.md.
> Structured for Vikunja import (Project → Task → Description hierarchy).
> Labels: `ag-task` = delegate to AI agent | `manual` = requires human | `done` = completed | `blocked` = has dependency
> Priority: P0 = this week | P1 = next sprint | P2 = backlog

---

## PROJECT: OCULTAR — Core & Infrastructure

---

### EPIC 0 — OpenAI Privacy Filter Integration

#### ✅ DONE — TASK 0.A — Model-agnostic Tier 2 interface
**Status:** Done  
**Label:** done  
`Tier2Engine` interface exists in `apps/slm-engine/pkg/inference/`. `PrivacyFilterEngine` (Python sidecar) and `LlamaCppEngine` both implemented. `TIER2_ENGINE` env var switches between them. `PRIVACY_FILTER_MODEL_PATH` supports custom/fine-tuned model paths.

---

#### TASK 0.1 — Benchmark Privacy Filter on French finance entities
**Priority:** P0  
**Label:** ag-task  
**Description:**  
Evaluate `openai/privacy-filter` (HuggingFace, Apache 2.0) against the current Tier 2 engine on finance-specific text. Use these test strings:
- `Transfer €84,293 from IBAN FR76 3000 6000 0112 3456 7890 189`
- `Vendor payment to Acme Corp, account 4532015112830366, approved by John Smith`
- `Cost center 4420-EMEA-CORP, GL account 6100, controller Sarah Chen`
- `Invoice INV-2026-00847 pour Société Générale, SWIFT SOGEFRPP`
- `CPAM DE L ISERE 253350004466` (contextual health reference — known gap)
- `john.doe@company.fr called +33 6 12 34 56 78 re: Q1 close`

Record: entity type, value, confidence, latency. Note misses.  
Output: `docs/benchmarks/privacy_filter_eval.md` as markdown table.

---

#### TASK 0.2 — Fine-tune Privacy Filter on French finance corpus
**Priority:** P1  
**Label:** ag-task  
**Blocked by:** TASK 0.1 (need gap list first)  
**Description:**  
Create synthetic French finance NER training data (200 train, 50 eval examples) for gap entities: FR IBAN, SIRET/SIREN, French phone (+33), SARL/SAS company names, CPAM/health reference numbers.  
Script: `scripts/fine_tune_privacy_filter.py` (HuggingFace Trainer API).  
Target: F1 > 0.92 on eval set.  
Output model: `models/privacy-filter-fr-finance/` (excluded from git via .gitignore).

---

#### TASK 0.3 — Doppler secrets management setup
**Priority:** P0  
**Label:** ag-task  
**Description:**  
Set up Doppler for `OCU_MASTER_KEY`, `OCU_SALT`, `OCU_LICENSE_KEY`. Create `doppler.yaml` in project root. Update `docker-compose.yml` to remove hardcoded env vars and use `doppler run --`. Write `docs/SECRETS.md` covering key rotation and loss consequences.  
Do NOT generate or hardcode key values.

---

#### TASK 0.4 — Blog post: "OpenAI shipped a model. We built the system."
**Priority:** P0  
**Label:** ag-task  
**Description:**  
Technical post, 800–1000 words, no company names. Three-event narrative:
1. March 2026: AI startup breach via open-source pipeline tool, 40K+ PII records
2. April 15 2026: French government agency breach, 19M citizen records
3. April 22 2026: OpenAI Privacy Filter (Apache 2.0) validates local-first detection

Argument: events 1–2 show what happens without sovereign PII infrastructure; event 3 gives a model but not a system; OCULTAR is the system.  
Output: `docs/blog/zero-egress-supply-chain.md`. Include HN submission title (280 chars).

---

#### TASK 0.5 — IP boundary documentation
**Priority:** P0  
**Label:** manual  
**Description:**  
Write two paragraphs defining the OCULTAR/Nous integration boundary and add them to each README. OCULTAR's responsibility ends at `POST /refine`. Nous must fail loudly if OCULTAR is unavailable — no graceful degradation.

---

#### TASK 0.6 — Fail-closed integration tests
**Priority:** P0  
**Label:** ag-task  
**Description:**  
Write integration tests in `services/refinery/pkg/refinery/failclosed_test.go` for:
1. SLM sidecar unreachable → HTTP 500, no PII in response body
2. Vault write timeout (mock 15s delay) → HTTP 500
3. Vault storage full (mock `ErrFull`) → HTTP 429
4. `protected_entities.json` empty → boot failure (not runtime panic)

Each test: assert correct status, assert no PII leakage, assert `X-Ocultar-Redacted` header absent.  
Add positive test: normal flow sets `X-Ocultar-Redacted: true`.

---

#### TASK 0.7 — SSRF protection + tests
**Priority:** P1  
**Label:** ag-task  
**Description:**  
Block RFC 1918 ranges, 127.0.0.1/localhost, 169.254.169.254 (AWS metadata), 0.0.0.0, [::1] in proxy.  
Test file: `apps/proxy/ssrf_test.go`. Assert 403 for each blocked range, no outbound connection attempted.  
Document: `docs/security/SSRF.md`.

---

### EPIC 1 — Documentation

#### TASK 1.1 — Main README rewrite
**Priority:** P1  
**Label:** ag-task  
**Description:**  
Audience: senior security engineer, 10 minutes to evaluate.  
Required: one-line hook, problem (3 bullets referencing AI supply chain attacks), ASCII architecture diagram, tier pipeline table, working curl example, Ed25519 audit log sample, Docker Compose quickstart, security model (6 bullets), pricing table linking to ocultar.dev.

---

#### TASK 1.2 — Security model documentation
**Priority:** P1  
**Label:** ag-task  
**Description:**  
Create `docs/security/SECURITY_MODEL.md` covering: zero-egress guarantee, fail-closed behavior table, AES-256-GCM + HKDF-SHA256, Ed25519 audit trail, SSRF protection, key management, compliance mappings (GDPR Art. 25, EU AI Act, HIPAA if applicable, SOC 2 readiness).  
Audience: CISOs and compliance teams.

---

#### TASK 1.3 — Nous integration guide
**Priority:** P1  
**Label:** ag-task  
**Description:**  
Create `docs/integrations/NOUS_INTEGRATION.md` with Docker Compose sidecar config, Go `PIIRedactor` implementation stub (fail-loudly pattern), re-hydration function, 5-step pipeline diagram with ownership labeled, failure handling table.  
Target reader: a developer who needs to integrate Ocultar in under 2 hours.

---

#### TASK 1.4 — GDPR Article 25 compliance pack
**Priority:** P1  
**Label:** ag-task  
**Description:**  
Create `docs/compliance/GDPR_PRIVACY_BY_DESIGN.md`. Map each OCULTAR design choice (zero-egress, fail-closed, AES-256-GCM, deterministic tokens) to GDPR Art. 25(1) and 25(2) requirements. Include short DPA template. Add EU AI Act alignment note for Nous integration context.

---

### EPIC 2 — Distribution

#### TASK 2.1 — Goose MCP extension
**Priority:** P1  
**Label:** ag-task  
**Description:**  
Python, `extensions/goose/ocultar_mcp.py`. Single tool: `refine_text(text, tier?)`. MCP server on stdio transport (preserves zero-egress). Config via `OCULTAR_URL` and `OCULTAR_API_KEY` env vars. Fail with MCP error (no passthrough) if OCULTAR unreachable.  
Also create `extensions/goose/README.md` and `requirements.txt`.

---

#### TASK 2.2 — Anthropic Claude MCP connector
**Priority:** P1  
**Label:** ag-task  
**Description:**  
Python, `extensions/claude/ocultar_claude_mcp.py`. Same as Goose but add `check_health` tool returning version/vault status/tier config. Include `INSTALL.md` with Claude Desktop config snippet, prerequisites, verification step.

---

#### TASK 2.3 — AWS Marketplace listing documentation
**Priority:** P2  
**Label:** ag-task  
**Description:**  
Create `docs/distribution/AWS_MARKETPLACE.md`. Contents: 500-char listing description, AWS Secrets Manager CloudFormation snippet for `OCU_MASTER_KEY`, ECS task definition with IAM least-privilege policy, health check config, customer FAQ (3 Qs), pricing tiers to configure.  
See separate section below for pricing guidance.

---

#### TASK 2.4 — ocultar.dev landing page update
**Priority:** P0  
**Label:** ag-task  
**Description:**  
Update `apps/web` landing page to reflect current state:
1. Replace "Local SLM inference" with "OpenAI Privacy Filter — 1.5B param, 97% F1, runs locally"
2. Add "Why OCULTAR vs Privacy Filter alone" comparison (vault, proxy, rehydration, audit trail, compliance layer, zero-egress guarantee)
3. Update tier descriptions to reflect Community / Enterprise model (not the old 4-tier pricing)
4. Verify all demo/CTA links are live

---

### EPIC 3 — Pricing and Revenue

#### TASK 3.1 — Paddle billing integration
**Priority:** P1  
**Label:** manual + ag-task (split)  
**Description:**  
Manual: Create Paddle account. Create two products: Enterprise Annual (€24,900), with possible Professional Monthly (€490) if added later. Enable EU VAT.  
Then ag-task: Implement Paddle webhook handler in `apps/proxy/handlers/paddle_webhook.go` — maps `price_id` to tier, updates `api_keys` table, handles cancellation (downgrade to Community, preserve vault data). Include signature verification.

---

#### TASK 3.2 — Prometheus metrics endpoint
**Priority:** P1  
**Label:** ag-task  
**Description:**  
`GET /metrics` (Enterprise API key required). Expose: tier hit counts by tier name, latency p50/p95/p99, vault token count and size, monthly API calls by key hash, queue depth.  
Also create: `monitoring/grafana-dashboard.json` with a stacked bar for tier hit rate (most important business metric: shows SLM% vs cheaper tiers).

---

#### TASK 3.3 — Weekly metrics tracker
**Priority:** P2  
**Label:** ag-task  
**Description:**  
`docs/metrics/WEEKLY_TRACKER.md` template + `scripts/metrics_snapshot.sh` that queries DuckDB and outputs a markdown table: API keys by tier, calls this month, tier hit rates, Privacy Filter vs llama.cpp split, vault token count, estimated AI cost.

---

### EPIC 4 — Compliance and Funding

#### TASK 4.1 — French SAS incorporation
**Priority:** P0  
**Label:** manual  
**Description:**  
Via Legalstart.fr (~€800) or accountant (~€2,000). Required for Bpifrance, EU VAT, contracts, future fundraise. As an American in Grenoble, consult an accountant on fiscal residency implications.

---

#### TASK 4.2 — Bpifrance innovation grant dossier
**Priority:** P1  
**Label:** ag-task  
**Blocked by:** TASK 4.1 (need legal entity)  
**Description:**  
Create `docs/funding/BPIFRANCE_DOSSIER.md`. Sections: project summary (FR + EN), innovation claim (5 bullets), market opportunity (€2.5B→€68B, 39% CAGR, BFSI 47%), TRL level (current 5–6, target 8–9), team profile, funding ask (€100K Prêt d'Innovation, 6-month runway + first hire). Leave [BRACKETS] for manual fill.

---

#### TASK 4.3 — Investor one-pager
**Priority:** P1  
**Label:** ag-task  
**Description:**  
`docs/fundraising/ONE_PAGER.md`. One printed page. Sections: problem (2 sentences referencing three 2026 events), why now (3 bullets), what we build (OCULTAR + Nous), technical edge (4 bullets), traction [PLACEHOLDERS], team (2 sentences), ask (€500–750K pre-seed, 18 months + first hire). Leave target investors and contact blank.

---

#### TASK 4.4 — Financial model
**Priority:** P2  
**Label:** ag-task  
**Description:**  
`docs/fundraising/financial_model.md` in markdown table format. 3-year projection: Y1 (3 Enterprise design partners at 50% off = €37K ARR), Y2 (5 Enterprise full price + pre-seed raise = €125K ARR + runway), Y3 (15 Enterprise + Series A target). Show ARR, gross margin, burn, runway. Flag assumptions clearly.

---

## PROJECT: NOUS

---

### EPIC 5 — Core Architecture

#### TASK 5.1 — Judgment log schema
**Priority:** P1  
**Label:** ag-task  
**Description:**  
`internal/judgmentlog/schema.go`: `JudgmentEntry` struct with `WorkitemID` (cross-references OCULTAR audit log), `ActorID`, `RoleCardID`, `AIRecommendation` (tokenized), `HumanDecision`, `CleanContext`, `TokenMap` (JSONB), `Ed25519Sig`.  
Also: migration `001_judgment_log.sql`, repository interface with CRUD, index on `WorkitemID` and `ActorID`.

---

#### TASK 5.2 — PIIRedactor wrapper
**Priority:** P1  
**Label:** ag-task  
**Description:**  
`internal/privacy/redactor.go`. Thin HTTP wrapper calling `POST {OCULTAR_URL}/refine`. Zero business logic. Fail loudly on 500/429/403/timeout — NEVER pass raw text through. Config: `OCULTAR_URL` (default `http://localhost:8080`), `OCULTAR_API_KEY`, `OCULTAR_TIMEOUT` (default 10s). Tests: happy path, OCULTAR down, rate limited, timeout.

---

### EPIC 6 — GTM

#### TASK 6.1 — Design partner outreach templates
**Priority:** P0  
**Label:** ag-task  
**Description:**  
Three templates: (A) cold LinkedIn message 300 chars, (B) warm intro email with 3 subject line options, (C) post-OCULTAR-integration upsell ("you're protecting what goes IN — Nous governs what comes OUT"). Each: max 150 words, one CTA, no deck attachment, no "Hope this finds you well."

---

#### TASK 6.2 — LinkedIn positioning post
**Priority:** P0  
**Label:** manual (post from your account)  
**Description:**  
Draft (ag-task), then post yourself. Topic: "What finance teams get wrong when deploying AI agents." 250–300 words. Three specific failure modes from AP/AR, FP&A, or month-end close. End with question. Max 3 hashtags. Practitioner voice, not thought leader.

---

## PROJECT: PERSONAL / OPERATIONAL

---

#### TASK P1 — Consulting rate card
**Priority:** P2  
**Label:** ag-task  
**Description:**  
`docs/consulting/RATE_CARD.md`. Services: Fractional GTM Lead (€2,500–3,500/month), AI Readiness Workshop (€1,500 flat), AI Implementation Advisory (€2,500/month, 3-month minimum), Speaking (€500/half-day). Each: who it's for, what they get (3 bullets), price, how to engage. Also write a 3-sentence LinkedIn bio that positions these without revealing competing products.

---

## EXECUTION ORDER — Next 2 Weeks

```
WEEK 1 — Launch-critical items:

Day 1 (today):
  [manual] French SAS: contact Legalstart.fr for quote      # unblocks Bpifrance
  [ag-task] TASK 0.4: three-event blog post                 # news cycle open now
  [ag-task] TASK 0.3: Doppler secrets setup                 # before any public exposure
  [ag-task] TASK 2.4: landing page update                   # reflect current product state

Day 2:
  [ag-task] TASK 0.1: Privacy Filter benchmark              # tells you if fine-tuning needed
  [ag-task] TASK 0.6: fail-closed integration tests         # close the security gap
  [manual]  Post blog post to HN (9am EST)                  # Tuesday is optimal

Day 3:
  [ag-task] TASK 1.1: Main README rewrite                   # before distribution push
  [ag-task] TASK 2.1: Goose MCP extension                   # distribution
  [manual]  Submit to Goose directory

Day 4:
  [ag-task] TASK 2.2: Anthropic Claude MCP connector        # distribution
  [ag-task] TASK 6.1: Outreach templates                    # start design partner pipeline
  [manual]  Send 5 outreach messages

Day 5:
  Review all ag-task output, fix, commit, push
  [manual]  TASK 6.2: Post LinkedIn

WEEK 2 — Enterprise readiness:

  TASK 1.2: Security model doc                              # sales asset
  TASK 1.3: Nous integration guide                          # unblocks Nous integration
  TASK 1.4: GDPR Article 25 compliance pack                 # enterprise requirement
  TASK 0.7: SSRF protection                                  # security hardening
  TASK 3.1: Paddle billing (manual part first)              # revenue infrastructure
  TASK 4.2: Bpifrance dossier (after SAS confirmation)      # funding

Priority if time is short:
  1. Blog post (news cycle is closing)
  2. Landing page update (first impression for inbound)
  3. Doppler (security hygiene before HN)
  4. Outreach templates (pipeline is the business)
  5. Everything else
```
