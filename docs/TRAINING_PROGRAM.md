# OCULTAR Internal Training Program: Zero-Egress Data Mastery

## 1. Course Overview

**Purpose of the Training**
This course is designed to equip all internal team members—sales, marketing, developers, support, and operations—with a robust understanding of the OCULTAR platform and the Sombra Gateway. You will learn what the platform does, how it works, how to run it, and how to effectively explain its value to enterprise clients.

**Expected Outcomes**
By the end of this course, you will be able to:
1. Explain the fundamental problem OCULTAR solves and why our Zero-Egress architecture is unique.
2. Understand the components of the platform, including the core engine, the Enterprise Dashboard, and the Sombra Gateway.
3. Install and run OCULTAR and Sombra locally or via Docker.
4. Test the system using actual sensitive data (PII) to see the live redaction and rehydration.
5. Troubleshoot common setup and connectivity issues.
6. Speak confidently about the product to both technical and non-technical stakeholders.

---

## 2. Conceptual Foundations

### The Problem
When enterprises use AI models like ChatGPT, Gemini, or Claude, they often send customer names, financial records, and health data to external servers. Under GDPR (Article 28), HIPAA, and other frameworks, exposing this Personally Identifiable Information (PII) to third-party clouds is a reportable liability, carrying massive fines.

### The Solution: OCULTAR
OCULTAR is a **Zero-Egress Data Refinery**. "Zero Egress" means sensitive data never leaves the client's infrastructure. OCULTAR acts as a secure bridge, catching PII in real-time, locking it in a local encrypted vault, and sending meaningless tokens (e.g., `[EMAIL_3a9f2b01]`) to the AI. When the AI responds, OCULTAR rehydrates the tokens back into the original data before returning it to the user.

### Why it Matters
Unlike tools like AWS Macie or Google DLP, which require sending your data to *their* cloud to be scanned, OCULTAR operates natively on the client's hardware. It provides **Privacy-by-Design**, converting regulatory liabilities into audit-safe assets.

---

## 3. System Architecture

OCULTAR is composed of several independent but cooperating components:

1. **The Live Refinery (Engine):** The core processor that redacts PII using three tiers of defense:
   - *Tier 0 (Dictionary Shield):* Exact matches for VIPs and custom internal terms.
   - *Tier 1 (Lead Shield):* High-speed Regex (Phones, URLs, IBANs, Emails).
   - *Tier 2 (Deep Scan):* Local SLM (Small Language Model) for contextual NER (Named Entity Recognition).
2. **The Identity Vault:** An encrypted local database (DuckDB or PostgreSQL) that securely maps raw PII to tokens.
3. **The OCULTAR Proxy:** A transparent HTTP proxy that sits between internal tools and external LLMs, stripping PII on the fly.
4. **Sombra Gateway:** The intelligent, agentic layer *above* the proxy. It features:
   - Multi-Model Routing (route requests to OpenAI, Gemini, or local models).
   - Connectors (ingest local files or APIs before querying the LLM).
   - The single `/query` endpoint orchestrating the ingest → redact → route → respond → rehydrate workflow.
5. **Enterprise Dashboard:** A visual interface for real-time monitoring, ROI calculation, and compliance mapping (Risk Matrix).

**Data Flow:**
Data/Prompt → Sombra/Proxy → Tier 0/1/2 Redaction → Encrypted Vault Storage → Tokenized Prompt sent to LLM → LLM returns Tokenized Response → Sombra/Proxy reads Vault → Rehydrated Plaintext returned to User.

---

## 4. Installation and Setup

### Local Installation (Community/Developer Mode)
*Prerequisites: Go 1.22+ installed.*
1. **Clone the Repo:** 
   ```bash
   git clone https://github.com/Edu963/ocultar.git
   cd ocultar
   ```
2. **Setup Dictionary Shield:** Ensure `configs/protected_entities.json` exists with at least an empty JSON array `[]` or dummy terms `["Project Nightshade"]`. OCULTAR fails safely (Fail-Closed) and won't start without it.
3. **Set Master Key:** 
   ```bash
   export OCU_MASTER_KEY="local-dev-key-123"
   ```
4. **Start the Proxy:**
   ```bash
   go run ./cmd/proxy
   ```

### Running the Full Stack (Docker)
For a complete environment test including the Database and Dashboard:
```bash
docker compose -f docker-compose.proxy.yml up -d
```
You can access the Dashboard at `http://localhost:8080` (or the port defined in your `.env`).

### Setting up Sombra Gateway
1. Clone Sombra alongside Ocultar:
   ```bash
   cd ..
   git clone https://github.com/Edu963/sombra.git
   cd ocultar
   go work use ../sombra
   ```
2. Configure `sombra.yaml` with your `OPENAI_API_KEY` or `GEMINI_API_KEY`.
3. Start Sombra (runs on port 8081 by default).

---

## 5. Hands-On Labs

### Lab 1: Seeing the Proxy in Action
1. Ensure the OCULTAR Proxy is running.
2. Send a request with your personal phone number and email:
   ```bash
   # For Local Developer Proxy (port 8080)
   curl -X POST http://localhost:8080/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{"messages": [{"role": "user", "content": "My email is eduardo@test.com and phone is +33 6 12 34 56 78."}]}'

   # For Enterprise Engine Testing (port 9090)
   curl -X POST http://localhost:9090/api/refine \
     -d '{"messages": [{"role": "user", "content": "My email is eduardo@test.com and phone is +33 6 12 34 56 78."}]}'
   ```
3. Look at your proxy console logs: You will see OCULTAR redacting the data into `[EMAIL_...]` and `[PHONE_...]` tokens before passing it upstream (or returning it directly in the Enterprise API).

### Lab 2: Multi-Model Routing with Sombra
1. Ensure Sombra Gateway is running.
2. Route a request explicitly to OpenAI, then to Gemini:
   ```bash
   curl -X POST http://localhost:8081/query \
     -F "connector=file" \
     -F "model=gpt-4o" \
     -F "prompt=Analyse these transactions." \
     -F "file=@test_statement.csv"
   ```
3. Check the response. Sombra ensures that no matter which LLM is selected, data policies (such as redacting SSNs or Account Numbers) are strictly enforced before routing.

### Lab 3: Creating a Custom Dictionary Rule
1. Open `configs/protected_entities.json`.
2. Add your own name: `["Carlos Segura"]`.
3. Restart the Proxy/Sombra.
4. Send a prompt containing "Carlos Segura" and verify it is redacted.

---

## 6. Troubleshooting

**"Fatal: protected_entities.json is missing"**
*Cause:* OCULTAR's Fail-Closed security requires the Tier 0 config to exist. 
*Fix:* Create the file `configs/protected_entities.json` containing `[]`.

**"401 Unauthorized" from the LLM**
*Cause:* Your upstream API keys are missing.
*Fix:* Export `OPENAI_API_KEY` or `GEMINI_API_KEY` in the terminal before running Sombra or the Proxy.

**Raw tokens `[EMAIL_...]` appearing in the final chat response**
*Cause:* Rehydration failed. Sombra and OCULTAR are likely using different vault files or different `OCU_MASTER_KEY` / `OCU_SALT` values.
*Fix:* Ensure `vault_path` in `sombra.yaml` matches the proxy's `OCU_VAULT_PATH` and standardise your environment variables.

**Sombra Error: "ai model request failed: local: HTTP 404"**
*Cause:* Sombra is attempting to route the scrubbed prompt to the Local AI model, but the `endpoint` configured in `sombra.yaml` doesn't exist or isn't serving an LLM API. 
*Fix:* Check `configs/sombra.yaml` and verify the `endpoint` for your `local` provider. It should point to the port of your actual LLM (e.g., `http://localhost:8080` for llama.cpp), *not* the OCULTAR API port (9090).

**Sombra Error: "gemini: HTTP 404 ... is not found"**
*Cause:* The Google Gemini API requires specific model names that may differ from their marketing names. If you request an invalid model name, Google returns a 404.
*Fix:* Ensure the model `name` in `sombra.yaml` uses the exact programmatic name found in Google AI Studio, such as `gemini-flash-latest`.

---

## 7. Real-World Scenarios

- **Finance Sector:** A bank needs to summarize daily transaction notes containing SWIFT codes, IBANs, and client account balances. OCULTAR strips all financial identifiers at Tier 1, queries Claude-3 for the summary, and rehydrates the PDF locally. Target: Bank CISOs.
- **Healthcare & Pharma:** Summarizing patient medical histories (containing names, SSNs, ICD-10 codes) into research datasets without violating HIPAA.
- **DevOps (The Git Lead Shield):** Developers constantly paste database connection strings or AWS keys into Jira/Git. OCULTAR acts as a pre-commit hook that blocks the commit if PII/secrets are detected.

---

## 8. Operational Skills

1. **Monitoring the Dashboard:** Navigate to the local Enterprise Dashboard. Watch the Risk Matrix map real-time data onto regulatory liabilities (e.g., GDPR, EU AI Act).
2. **Viewing Audit Logs:** Understand the SIEM-compatible JSON logs generated by OCULTAR Enterprise regarding every redaction transaction.
3. **Configuration Management:** Recognize how `configs/config.yaml` can be used to add new Regexes seamlessly without recompiling Go code.

---

## 9. Evaluation Framework

Traditional exams are not used here. Instead, employees will prove operational competency:
- **Task 1:** Successfully run OCULTAR proxy and Sombra locally.
- **Task 2:** Write a prompt containing 3 distinct types of PII (a name, an email, and an IBAN).
- **Task 3:** Send the prompt through Sombra, capture the output, and show the terminal log proving it was tokenised in transit.
- **Task 4:** Add a custom word to the Dictionary Shield and prove it redacts.
- **Task 5:** Document the solution to one intentionally broken config state (e.g., mismatched Vault Path).

---

## 10. Certification Milestones

- **Explorer:** Completed the reading, passed Task 1. Can explain Zero-Egress in plain terms.
- **Operator:** Passed Tasks 1-4. Comfortable running demonstrations of the product for potential clients, showing live redaction.
- **Advanced Operator:** Passed Task 5. Capable of deploying via Docker, modifying configuration files, diagnosing Vault rehydration issues, and interpreting SIEM logs.

---

## 11. Recommended Documentation and Learning Materials

After completing this course, reference the following materials for in-depth technical knowledge:
1. **PITCH.md**: For sales and marketing messaging on Value Proposition.
2. **FAQ.md**: The definitive "cheat sheet" for answering complex client questions on zero-egress, pricing, and latency.
3. **ARCHITECTURE.md**: For developers wanting to understand the Go Engine structure.
4. **SOMBRA_GUIDE.md**: Detailed instructions on configuring Multi-Model routing and connectors.
5. **DEVELOPER_GUIDE.md**: Internal code standards, CI/CD checklist, and engine extension tutorials.
