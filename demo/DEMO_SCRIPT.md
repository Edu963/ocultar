# OCULTAR Enterprise — Client Demo Script

**Duration**: 12–15 minutes  
**Command**: `./demo/run_demo.sh --record`  
**Audience**: CISO, Head of Engineering, Compliance Officer  
**Goal**: Make them feel the risk first, then show the solution works live.

---

## Part 0 — Before the Meeting  *(discovery, 24–48 hours before)*

Send a short email or ask on the intro call. Four questions:

1. **"Which AI providers are you currently using or evaluating?"**  
   If they say OpenAI → emphasize zero-egress from US clouds.  
   If they say Azure OpenAI → they think they're safe; show them they aren't (prompts still leave the org).  
   If they say nothing yet → you're shaping the architecture from day one.

2. **"Are you under GDPR, HIPAA, PCI DSS, or a sector-specific framework?"**  
   This tells you which scenario C audit entry to emphasise and which objection you'll face.

3. **"Who else will be in the room?"**  
   CISO only → keep it risk and compliance, skip the Go build step.  
   Engineering lead → they'll want to see the code; offer to open `services/refinery/pkg/refinery/refinery.go`.  
   Legal/DPO → skip latency numbers, lean hard on the audit log and GDPR Art. 25 pack.

4. **"Are you already using a DLP or data-masking solution?"**  
   If yes → your angle is rehydration (DLP blocks, OCULTAR restores).  
   If no → your angle is time-to-compliance (one command, done).

---

## Part 0.5 — Setup Checklist  *(15 minutes before)*

- [ ] `cd ~/ocultar && git pull` — run on latest
- [ ] `./demo/run_demo.sh` once without `--record` to pre-warm the build cache
- [ ] Terminal font size ≥ 16pt, dark background, full screen
- [ ] Close Slack, email, notifications
- [ ] If using Gemini: `export GEMINI_API_KEY=...` so the demo shows a real AI call
- [ ] Have `docs/compliance/GDPR_PRIVACY_BY_DESIGN.md` open in a second tab — you may want to show it during the close

---

## Before you start the demo

Open a terminal, navigate to the repo root, and have the command ready to paste.
Do **not** run it yet. The cold open is verbal.

---

## 1. The Cold Open  *(~90 seconds, no screen)*

> "I want to ask you something before I show you anything.
>
> Every time one of your engineers pastes a support ticket into ChatGPT,
> every time someone drops a spreadsheet into Copilot,
> every time your app calls OpenAI with user data in the prompt —
> **where does that data go?**
>
> OpenAI's terms say they may use API inputs for safety monitoring.
> Gemini's terms say Google may review conversations.
> Your DPA with the client says their data stays in the EU.
>
> Those two things cannot both be true at the same time.
>
> OCULTAR solves that. Not with a policy. With a technical guarantee.
> Let me show you exactly how it works."

---

## 2. The Architecture in One Sentence  *(~30 seconds)*

Point to the terminal.

> "What you're about to see is a live system running on this machine.
> No cloud. No sandbox. Your data never leaves the box.
>
> OCULTAR sits between your application and the AI provider.
> It strips out every piece of sensitive data before the request goes out,
> replaces it with encrypted tokens, and restores it in the response
> before your user sees it.
>
> The AI never sees real PII. Ever."

Now run the script:

```
./demo/run_demo.sh --record
```

---

## 3. Pre-flight & Build  *(let it run, narrate lightly)*

While it builds (~20 seconds):

> "We're compiling and booting four services right now:
> the Enterprise Refinery, the Sombra Gateway, and two lightweight stubs
> standing in for the AI backend and the NLP sidecar.
>
> In production you'd point this at your actual OpenAI or Gemini endpoint.
> We'll use mock services today so nothing leaves this room."

When you see all four `✓` checks:

> "Four services up, healthy, zero configuration beyond a license key.
> Docker Compose in production — single command."

---

## 4. Scenario A — Enterprise Refinery  *(direct PII detection)*

When the raw input appears:

> "This is a realistic enterprise prompt. A user is asking an AI assistant
> to summarise a performance review. It contains a full name, a work email,
> a French social security number, an IBAN, and a mobile number.
>
> Five categories of PII in one sentence. Completely normal in finance or HR."

When the refined output appears — **pause here, let them read it**:

> "This is what the AI provider receives.
>
> Jean-Pierre Dupont became a token. His email, his SSN, his IBAN, his phone —
> all replaced with short opaque identifiers.
>
> These tokens are deterministic: the same input always produces the same token.
> That means if the same person appears in two different documents,
> the AI sees the same token both times — consistent, but anonymous.
>
> The real data is stored in an AES-256-GCM encrypted vault on your hardware.
> The AI provider has zero access to it."

When the count appears:

> "Four entities, detected and vaulted in under 50 milliseconds.
> No GPU. No cloud call. Pure local pattern matching."

---

## 5. Scenario B — Sombra Gateway end-to-end  *(the 'aha' moment)*

When the user request appears:

> "Now Sombra. This is the Enterprise gateway layer.
>
> Your application sends a perfectly normal OpenAI SDK request —
> same format, same headers, same model name.
> The only change is the base URL points to Sombra instead of OpenAI directly.
> **Zero code changes in your application.**"

While the AI call runs:

> "Right now, Sombra is doing three things in sequence:
> stripping the PII and vaulting it,
> forwarding the clean request to the AI,
> and waiting to rehydrate the response."

When the mock AI section appears (showing tokens):

> "This is what the AI saw. Tokens, not people.
> If this were a real Gemini or GPT call, Anthropic's servers,
> Google's servers, OpenAI's servers — none of them would have
> seen a single real name, number, or identifier.
>
> Zero egress. Not 'we try our best'. Zero. Technical guarantee."

When the rehydrated response appears — **pause again**:

> "And the caller gets back the original PII, fully restored.
> From the user's perspective, nothing changed.
> From the compliance perspective, everything changed."

---

## 6. Scenario C — Immutable Audit Trail  *(the compliance closer)*

When the audit log entry appears:

> "Every single request produces a signed log entry.
>
> Ed25519 signature. SHA-256 hash chain — each entry references
> the hash of the previous one. You cannot alter, delete, or reorder
> entries without breaking the chain.
>
> This is what GDPR Article 5(2) means by accountability.
> This is what your SOC 2 auditor is going to ask for.
> This is what your ISO 27001 assessor checks.
>
> Most companies point their auditor at CloudTrail logs or Splunk dashboards.
> You can point them at a mathematically tamper-proof record."

---

## 7. The Summary Screen  *(read it slowly)*

> "Four PII types detected. Zero data reached the AI provider.
> The caller received everything they expected.
> Every event signed and chained.
>
> That is the Enterprise tier in a twelve-minute demo."

---

## 8. The Enterprise vs Community Contrast  *(if asked, or volunteer it)*

> "We have a free Community version. It does the proxy and the regex detection —
> structured PII, emails, IBANs, phone numbers. Real and useful.
>
> Enterprise adds three things that matter at your scale:
>
> **First**, Tier 2 AI detection. Regex catches structured PII.
> But most leaks in LLM workflows aren't structured.
> They're 'send this to our CFO Pierre Martin' — a name in a sentence.
> The local SLM catches that. The regex doesn't.
>
> **Second**, the audit log you just saw. Community has no audit trail.
> You cannot pass a GDPR audit without one.
>
> **Third**, Sombra. The community proxy is one-to-one: one upstream, one path.
> Sombra routes to Gemini, GPT, Claude, or your own local model —
> dynamically, with the same API. You swap providers in config, not code."

---

## 9. Handling Objections

**"We already have a DLP solution."**
> "DLP watches data at rest and in transit. It doesn't rehydrate.
> When your AI returns a response, DLP has already let the PII through or blocked the whole request.
> We redact before the call and restore after. The AI never sees it."

**"Can't we just instruct the AI not to store our data?"**
> "You can ask. You cannot verify. We give you a cryptographic proof."

**"What if the SLM misses something?"**
> "The tiers run in sequence. Tier 2 is additive — it catches what regex misses.
> If it misses something, Tiers 0–1.5 already ran. And every miss is logged,
> so you can tune the pattern library. It improves over time."

**"We're on Azure / AWS / GCP, not bare metal."**
> "Runs on any compute: VM, Kubernetes, ECS, bare metal.
> The vault is DuckDB by default — one file, no infrastructure.
> For HA you swap to PostgreSQL in one config line."

**"What about response latency?"**
> "Tiers 0–1 add under 5ms. Tier 2 SLM adds 20–80ms depending on hardware.
> For chat interfaces that's imperceptible. For batch pipelines we have
> async modes. We can show benchmarks for your specific hardware."

---

## 10. The Close

> "Two things I'd suggest as next steps.
>
> One: we set up a proof-of-concept on your infrastructure —
> point it at one internal tool, one AI endpoint, run it for two weeks.
> You'll see exactly which PII is flowing where today.
>
> Two: we send you the GDPR Article 25 compliance pack.
> It maps every technical control you just saw to the specific
> regulatory requirement it satisfies. Your DPO will want to see it."

---

## 11. Audience Variants  *(adjust emphasis, not content)*

**CISO in the room**  
Lead with regulatory risk, not features. Use the word "liability" early. Scenario C (audit log) is your closer, not Scenario B. End with: *"If your DPO had to produce evidence of AI data handling tomorrow, what would they point to?"*

**Head of Engineering in the room**  
They'll ask about latency, deployment model, and whether it breaks their SDK. Jump to the health check URL after boot (`/healthz`), offer to show the proxy config (two env vars). Say: *"You literally change one line — OPENAI_BASE_URL — and you're protected."*

**Compliance Officer / DPO in the room**  
Skip the build narrative. Jump straight to Scenario C and spend twice as long on it. Show `docs/compliance/GDPR_PRIVACY_BY_DESIGN.md`. The phrase that lands: *"This is a technical implementation of Privacy by Design under Article 25 — not a policy, a control."*

**Skeptical buyer who's seen too many demos**  
Don't narrate the output. Let them read it in silence. When they look up, say: *"Nothing I just showed you is mocked except the AI endpoint. Want me to point it at your actual OpenAI key right now?"*

---

## 12. After the Meeting  *(same day)*

**Send within two hours:**

> Subject: OCULTAR — what you saw today + next step
>
> [Name],
>
> Thanks for the time today. Three things as promised:
>
> 1. **Recording** — attached. You can replay it with `asciinema play ocultar_demo_YYYYMMDD.cast`
>    (or I can send an MP4 if easier).
>
> 2. **GDPR Article 25 pack** — attached. Maps every control you saw to the specific
>    regulatory clause. Designed to go straight to your DPO.
>
> 3. **PoC proposal** — 2-week proof of concept on your infrastructure,
>    one internal tool, one AI endpoint. We configure it, you validate it.
>    No commitment beyond the two weeks.
>
> What's the best way to set up a 30-minute technical call with your engineering lead
> to scope the PoC?
>
> — [Your name]

**Attach:**
- `demo/ocultar_demo_YYYYMMDD.cast` (or an MP4 export)
- `docs/compliance/GDPR_PRIVACY_BY_DESIGN.md`

---

## 13. PoC Scope Template  *(leave-behind for engineering)*

If they agree to a PoC, send this one-pager:

**What we need from you:**
- One internal application that currently calls an AI provider
- The AI provider endpoint (OpenAI, Gemini, Azure OpenAI, etc.)
- A VM or container slot with 2 vCPU / 4 GB RAM

**What we deliver in week 1:**
- OCULTAR running in front of that endpoint
- A dashboard showing what PII was flowing before vs. after
- A sample audit log for your DPO

**What you validate in week 2:**
- Application behaviour is unchanged
- AI responses are correct
- Latency impact is acceptable (target: < 80ms p99 added)

**Success criteria (you define these):**
- [ ] Zero raw PII in outbound AI requests
- [ ] Audit log captures all events
- [ ] No application code changes required

---

## Timing Reference

| Section | Duration |
|---|---|
| Cold open | 90 sec |
| Architecture sentence | 30 sec |
| Build + boot | 20 sec |
| Scenario A | 2 min |
| Scenario B | 3 min |
| Scenario C | 90 sec |
| Summary | 30 sec |
| Enterprise vs Community | 2 min (if needed) |
| Objections | as needed |
| Close | 60 sec |
| **Total (no objections)** | **~12 min** |
