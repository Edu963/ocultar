# OCULTAR | Comprehensive FAQ

This document provides a clear, structured, and concise overview of the Ocultar project for internal teams, sales, and clients.

---

## 1. General Overview

### What is Ocultar and its main purpose?
Ocultar is a **Zero-Egress Data Refinery** designed to protect sensitive data (PII, secrets, proprietary terms) before it leaves a trusted environment. Its main purpose is to convert regulatory liabilities into audit-safe assets by redacting sensitive information in real-time before it reaches external AI providers like OpenAI, Gemini, or Claude.

### What does "Zero-Egress" mean?
"Zero-Egress" means that your sensitive data never leaves your infrastructure. Redaction happens locally on your hardware, and only non-sensitive tokens are sent to external APIs. The original data is stored in a local, encrypted vault that remains entirely under your control.

### How does Ocultar help with regulatory compliance?
Ocultar aligns with major frameworks like **GDPR**, **HIPAA**, **SOC 2 Type II**, **PCI DSS v4.0**, **NIS2**, **EU AI Act**, **BSI C5**, and **ISO 27001**. By ensuring PII never reaches third-party clouds, it eliminates reportable liabilities and helps avoid significant fines (e.g., up to €20M or 4% of global turnover under GDPR).

---

## 2. Core Components

### What is the Live Refinery?
The **Live Refinery** is the core processing refinery. It identifies and redacts sensitive data using a tiered approach, ranging from high-speed dictionary lookups to deep-scan AI analysis using local Small Language Models (SLMs).

### In Tier 1, where can a client modify a regex, and should this be in the Dashboard?
Enterprise clients modify regex rules via the `config.yaml` file. Much like the dictionaries, we intend to move this into a "no-code" section of the Dashboard to allow for dynamic updates without restarting the service.

### In Tier 2, which SLM (Small Language Model) is being used?
We primarily utilize Qwen-1.5B or Phi-3-mini, quantized to GGUF format (~1.2 GB). These are chosen for their high performance-to-size ratio, allowing them to run on standard enterprise hardware without a massive GPU.

### Are the SLM (Small Language Model) weights shipped directly within the installation files? What is the purpose of the `models` folder?
No, we do not ship the heavy SLM weights directly in the installation files to avoid bloating the distribution. Instead, when you run the Enterprise Docker Compose stack, an `init-slm` container automatically downloads the necessary model (e.g., Qwen 1.5B) at runtime. The `models` folder serves two purposes: it holds tokenizer configuration files for specific domains, and it serves as the local vault where the downloaded `.gguf` weights are saved and loaded by the proxy.

### If the default model is tuned for specific regions (like French financial formats), can we train or fine-tune the model to recognize other EU formats?
Yes. OCULTAR uses an open-source, multilingual SLM architecture. Enterprise clients can fine-tune a base model on their specific regional documents (e.g., German tax IDs or Spanish invoices). Once fine-tuned, the model is quantized into a 4-bit `.gguf` file using our provided scripts and placed into the `models` folder. OCULTAR's native CGO `llama.cpp` integration will automatically load the new weights. You can also instantly add static region-specific formats to the Tier 1 Lead Shield via regex in `configs/config.yaml` without touching the AI model.

### What is the Sombra Gateway?
The **Sombra Gateway** is the intelligent orchestration layer above Ocultar. It provides multi-model AI routing, allowing you to direct queries to different providers (OpenAI, Gemini, local models) while ensuring consistent PII redaction and response re-hydration across all channels.

### Which AI Models does Sombra Gateway support?
Sombra natively routes to **OpenAI**, **Gemini**, **Claude**, and **Local AI** providers. 
Crucially, because Sombra supports the OpenAI-compatible API standard, clients can seamlessly route traffic to **Mistral, DeepSeek, Qwen**, or any other compatible Chinese or open-source model simply by defining them in `sombra.yaml` with the `openai` provider type and updating the `endpoint` URL.

### I see references to `llama.cpp`. I thought we were using the OpenAI SDK?
You are actually dealing with two different AI models serving two different purposes. Your client application uses the standard **OpenAI SDK** to make requests to your main upstream LLM (like GPT-4 or Claude). However, OCULTAR acts as a proxy intercepting those requests. To scrub the PII from your prompt *before* forwarding it to OpenAI, OCULTAR runs a completely local, ultra-fast privacy filter model directly inside the proxy using **`llama.cpp`**. You use the OpenAI SDK for your primary AI tasks, and `llama.cpp` locally for the zero-egress security filter.

### Can you explain the Proxy mode further?
The Proxy is a transparent reverse proxy that sits between the client application and the LLM API. It redacts PII on the way out (Vaulting) and restores it on the way back (Re-hydration). It is the "Aha!" deployment method because it requires zero changes to the client's existing code—they just change their API base URL to point to OCULTAR.

### In the Community version, do we still get the end-to-end interception and redaction shown in the Enterprise demo?
Yes, the core mechanism (Intercept → Redact → Forward → Rehydrate) absolutely still happens in the Community version. The substitute for the Sombra Gateway is simply called the **OCULTAR Proxy**. The difference is that the Community Proxy is a "one-to-one" proxy hardcoded to a single upstream provider (e.g., it only talks to OpenAI), whereas Sombra is an intelligent gateway that can route to multiple providers dynamically based on the request. Additionally, the Community Proxy only uses Regex and Dictionary detection (Tier 0 and 1) and will miss unstructured PII (like a name in a sentence) because it lacks the local Tier 2 AI engine.

### What is the Enterprise Dashboard?
The **Enterprise Dashboard** is a browser-based UI for monitoring and management. it provides:
- Real-time visualization of data refinement.
- **Risk Matrix**: Mapping leaks to regulatory categories.
- **ROI Analytics**: Quantifying financial savings and fine-avoidance.
- Identity Vault management and audit log review.

### Is there a UI to manage dictionaries and connectors?
Yes. The Dashboard includes a **Shield Manager** interface. This allows security officers to manage dictionaries and regex patterns directly via the UI without touching configuration files. Modifications are hot-reloaded and persisted to `configs/config.yaml`.

### What is the Identity Vault?
The **Identity Vault** is a local, encrypted database (DuckDB or PostgreSQL) that stores the mapping between original PII and its corresponding tokens. It allows the system to "re-hydrate" (restore) original values in the AI's response for the end-user, without ever exposing them to the AI provider.

### Which databases are included, and can a client pick their own?
OCULTAR includes DuckDB for Community/Single-node use and supports PostgreSQL for Enterprise HA clusters. Clients can choose their backend in `config.yaml`. DuckDB is a file (`vault.db`) sitting on the local host, while PostgreSQL can be an external, client-managed cluster.

### Should we eliminate the restrictions in DuckDB?
The main restriction in DuckDB is that it is single-process (not suitable for multi-node horizontal scaling). Rather than "eliminating" this inherent trait of DuckDB, we offer PostgreSQL as the upgrade path for clients requiring high availability and horizontal scalability.

### What is the Dictionary Shield?
The **Dictionary Shield (Tier 0)** is the first line of defense. it uses a mandatory list of protected terms (VIP names, internal project codes, proprietary keywords) to perform instant, exact-match redaction.

### In Tier 0, how can a client add its own dictionaries?
Currently, clients add terms by editing the `configs/protected_entities.json` file. This is a "Fail-Closed" dependency; if the file is missing or contains invalid JSON, the refinery will refuse to start to ensure no data is processed without the primary shield.

### Can we connect to external tools like LDAP, CRMs, or Databases for Tier 0?
Yes. OCULTAR Enterprise includes a **Live Identity Sync** worker. It natively polls external CRM/LDAP endpoints (e.g., Salesforce, Workday) to perform automated, real-time "Identity Ingestion" of protected names and VIPs into the Tier 0 shield.

---

## 3. Security & Privacy Principles

### What is the "Fail-Closed" guarantee?
Ocultar is built on a **Fail-Closed** principle: if any part of the security pipeline fails (e.g., a missing config file or an refinery error), the request is blocked and never forwarded to the upstream API. Security is prioritized over availability to prevent accidental data leaks.

### How is data encrypted in the vault?
Data is encrypted using **AES-256-GCM** with a master key (`OCU_MASTER_KEY`) that exists only in-process RAM during execution. Even if the vault file is compromised, the content is unreadable without the master key.

### Does Ocultar provide Privacy-by-Design?
Yes. Every component is architected to minimize data exposure. Clear trust boundaries ensure that plain-text PII never leaves the trusted zone, and even internal logs use tokens instead of raw data.

---

## 4. Target Users

### How does Ocultar benefit different roles?
- **CISO**: Provides a "Switzerland of Data" approach, ensuring neutral, local, and legally clean compliance oversight.
- **DevOps/SRE**: Offers a transparent sidecar proxy that integrates into existing CI/CD pipelines with minimal configuration.
- **Pilot Managers**: Facilitates rapid onboarding with industry-specific snapshots and clear ROI metrics.
- **App Developers**: Simplifies AI integration by handling all privacy concerns at the gateway level.

---

## 5. Typical Workflows

### What is the difference between Pilot and Enterprise workflows?
- **Pilot**: A 90-day structured onboarding focused on refinery deployment, PII pattern tuning, and weekly compliance reports.
- **Enterprise**: Full-scale production deployment with high-availability (PostgreSQL), SIEM integration, and ongoing ROI accounting.

### In production, does the end-user input their message directly into the public chatbox of ChatGPT, Mistral, or Claude?
No. If the user types directly into the public chatgpt.com web interface, the traffic bypasses OCULTAR entirely. The user must type their prompt into your internal enterprise application, internal chatbot, or an enterprise workspace that points its API requests to the OCULTAR Sombra Gateway.

### How can I update Refinery rules?
Rules can be updated via the `config.yaml` file (for Regex and Dictionaries) or automatically generated using the `Refinery Rule Generator` AI skill based on discovered edge cases.

---

## 6. Deployment & Packaging

### How are client packages delivered?
Ocultar is delivered as clean, versioned release artifacts (e.g., `.tar` or `.zip` archives) built by specialized AI agents that ensure all secrets are sanitized before delivery.

### Does Ocultar support air-gapped environments?
Yes. The Enterprise Tier is specifically designed for air-gapped support, allowing local SLM-based PII detection and on-premise vault management without any external internet requirements.

---

## 7. Industry-Specific Use Cases

### Which industries are supported out-of-the-box?
Ocultar includes pre-configured snapshots for:
- **Finance**: IBANs, SWIFT codes, PCI-DSS patterns, and proprietary ticker symbols.
- **Healthcare**: Patient IDs, HIPAA identifiers, and medical code detection (ICD-10).
- **GovTech**: SSNs, tax IDs, and classified project codenames.

---

## 8. AI Skills / Agent Roles

### What are "Specialized Agent Skills"?
Ocultar uses a decentralized network of AI agents to maintain system integrity. Key roles include:
- **Continuous AI Orchestrator**: Manages the 16-step protocol for every system change.
- **ROI Accountant**: Quantifies financial impact and potential fines avoided.
- **Red-Team Evasion Scanner**: Proactively tests the refinery for bypasses (e.g., Base64 or URL encoding attacks).
- **Documentation Updater**: Keeps all guides and FAQs in sync with the codebase.
- **Performance Benchmarker**: Monitors latency and suggests pipeline optimizations.

---

## 9. Performance and Latency

### What is the latency "tax" of using Ocultar?
Latency is minimal. Tier 0 and Tier 1 (Regex/Dictionary) run at disk speed. Tier 2 (Local SLM) adds a small overhead, which is monitored and optimized for real-time interactions. The `Sombra Performance Benchmarker` identifies and optimizes any bottlenecks.

### Is the system scalable?
Yes. Ocultar supports horizontal scaling in Enterprise mode using a PostgreSQL HA Vault, allowing multiple proxy instances to share the same identity mappings.

### How does the system handle slow AI models?
Ocultar Enterprise uses a **Fail-Closed** design for SLM scans. If the AI model (e.g., Qwen or Phi) takes longer than 5 seconds to respond, the refinery defaults to high-security mode (redacting chunks it cannot verify). We recommend using ultra-light models (< 1B parameters) and the **SLM AI Relay** for caching to maintain real-time performance.

---

## 10. Compliance & Audit Reporting

### How does the system handle audit logs?
All processing events are logged in a SIEM-compatible JSON format. These logs map every transaction to specific regulatory liabilities (e.g., "Health/Bio" or "Business Secrets") and risk levels, providing a clear path for compliance auditing.

---

## 11. Git & Workflow

### Why are `dist/*.zip` and `dist/*.tar.gz` not tracked in Git?
These are **generated artifacts**. Your pre-commit hook (`orchestrate.sh`) rebuilds them automatically during every commit attempt to ensure they are always up-to-date. If Git tracked them, the act of "building" them would immediately create a new "Modified" status, putting you in an endless loop where your branch is never clean. We exclude them from tracking but keep them in the `dist/` folder for easy distribution to clients.

### Where is the "Source of Truth" for my code?
The source of truth is the `/home/edu/dev/ocultar` directory. Other folders (like the now-deleted `ocultar-lab`) are for testing and local synchronization only. Always perform edits in the `dev/` folder to ensure they are propagated and committed correctly.

---

## 12. Troubleshooting

### Why does the demo show "Mock SLM Sidecar" and "Mock AI" in the service list?

The demo (`./demo/run_demo.sh`) deliberately starts two lightweight mock services so it can run end-to-end with no external dependencies.

**Mock AI** is only active when no real API key is found. If `GEMINI_API_KEY` or `OPENAI_API_KEY` is set in your shell, the demo auto-detects it and routes Sombra to the real model — Mock AI never starts. To record a demo with a live model:

```bash
GEMINI_API_KEY=your-key ./demo/run_demo.sh --record
```

**Mock SLM Sidecar** replaces the real Tier 2 NER engine (`apps/slm-engine`). The real sidecar loads a quantized NLP model (Qwen-1.5B or Phi-3-mini, ~1 GB) and takes 15–30 seconds to warm up — unsuitable for a sub-20-second demo startup. The mock returns empty scan results, meaning Tier 2 contributes no additional detections. However, the demo PII (email, SSN, IBAN, phone) is fully caught by Tiers 0–1.5 (regex, dictionary, phone, IBAN heuristics), so the core zero-egress proof is intact.

If you want Tier 2 to actively catch something — for example, a name buried in prose with no structured pattern — start the real SLM sidecar separately before running the demo:

```bash
go run ./apps/slm-engine   # starts on :8085
./demo/run_demo.sh
```

The demo will detect port 8085 is already occupied and skip starting the mock, routing the Enterprise Refinery to the real engine instead.

### Why does Sombra return "gemini: HTTP 404 ... is not found"?
The Google Gemini API requires specific programmatic model names (e.g., `gemini-flash-latest` instead of `gemini-1.5-flash`). If Sombra requests an invalid name, Google's API will return a 404 error. Ensure your `sombra.yaml` configuration uses the exact name found in Google AI Studio, and that your API or `curl` calls request this exact name.

---

> [!WARNING]
> **Security Guidance:** Never expose your `OCU_MASTER_KEY` or `OCU_LICENSE_KEY` in version control or logs. Use environment variables or secure vault managers.

> [!NOTE]
> For more technical details, refer to the [Architecture Reference](./ARCHITECTURE.md) and [Developer Guide](./DEVELOPER_GUIDE.md).
