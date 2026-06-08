# OCULTAR | Setup & Usage Guide

This guide covers how to run OCULTAR locally, package it for a client, and use the dashboard.

---

## Quick Start (5 minutes)

The fastest path to a working demo. You only need [Docker](https://docs.docker.com/get-docker/) and Docker Compose v2 — no secrets manager, no Go toolchain.

```bash
git clone https://github.com/Edu963/ocultar
cd ocultar
docker compose up --build
```

First build takes ~5 minutes (compiling Go + DuckDB/CGO). Every subsequent start is instant.

**What's running:**

| Service | URL | What it does |
|---------|-----|--------------|
| `ocultar-proxy` | http://localhost:8081 | PII detection + redaction proxy |
| `echo-upstream` | http://localhost:8082 | Mock AI API (reflects the request back) |
| Prometheus metrics | http://localhost:9090/metrics | Live tier hit rates, latency, vault size |

**Test it:**

```bash
curl -s http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"My email is alice@example.com and SSN is 123-45-6789"}]}'
```

The echo response shows `[EMAIL_...]` and `[SSN_...]` tokens — the originals never reached the upstream.

**Add AI-powered Tier 2 NER** (downloads ~500 MB model on first run):

```bash
OCU_PILOT_MODE=0 docker compose --profile ai up --build
```

**Use real API keys:**

```bash
cp .env.example .env   # edit with your keys and upstream URL
docker compose up --build
```

> **Security note:** The default demo keys in `.env.example` are intentionally insecure. Always set `OCU_MASTER_KEY` and `OCU_SALT` to high-entropy values before connecting real data.

---

## Part 1: For the Sender (Packaging for a Client)

**Your Goal:** Package OCULTAR for a client deployment.

### Step 1 — Build the release

Open a terminal in the `ocultar/` folder and run:

```bash
./build_release.sh
```

This compiles and packages both editions. It creates:
- `dist/ocultar-community.zip` — Default binary (Community Tier)
- `dist/ocultar-enterprise.tar.gz` — Enterprise-ready binary (ready for OCU_LICENSE_KEY)

### Step 2 — Send the right file

Send the `ocultar-community.zip` to your client for a standard pilot.

---

## Part 2: For the Receiver (Running OCULTAR)

**Your Goal:** Start OCULTAR on your computer (Windows, Mac, or Linux) and test it.

OCULTAR is delivered as a Docker container — it requires zero programming knowledge and does not connect to the internet.

### Prerequisites

You need **Docker Desktop** installed:
- **Download here:** [https://www.docker.com/products/docker-desktop/](https://www.docker.com/products/docker-desktop/)
- Install it like any normal application and open it (look for the whale icon in your system tray or menu bar).

### Step 1 — Unzip the application

Extract `ocultar-community.zip` to your Desktop. Open the extracted folder.

### Step 2 — Run the setup launcher

Inside the folder, open the `scripts/` folder and run the launcher for your system.

**Windows (Community):**
1. Right-click `scripts\setup-community.ps1`
2. Select **Run with PowerShell**
3. If a blue box pops up saying "Windows protected your PC", click **More info** then **Run anyway**. If prompted about execution policy, type `Y` and press Enter.
4. The script automatically generates your encryption keys and starts the refinery.

**Mac / Linux:**
1. Open your Terminal.
2. Drag and drop `scripts/setup-community.sh` into the Terminal and press **Enter**.
3. For **Enterprise Tier** activation, proceed to Part 2.5 after setup.

> **First run note:** The setup pulls a local AI model (~1.2 GB) and builds the local binary. This takes a few minutes once. Every subsequent run starts instantly.

### Step 3 — Open the Dashboard

Once setup completes, open your web browser and go to:

👉 **http://localhost:3030**

You should see the OCULTAR Live Dashboard with the input panel on the left.

---

## Part 2.5: Enterprise Licensing & Pilot Onboarding

**Your Goal:** Activate your Enterprise trial and run a 14-day technical pilot.

### 1. License Activation
OCULTAR uses an **Offline Ed25519** model. To activate:
1. Locate your `OCU_LICENSE_KEY` (provided by the Ocultar account team).
2. Paste it into your `.env` file: `OCU_LICENSE_KEY=Signature.Payload`.
3. Restart the refinery. The dashboard will now unlock the **Risk Matrix** and **ROI Analytics**.


---

## Part 3: Using the Dashboard

### The Interface

| Area | What it does |
|---|---|
| **Sidebar** | Navigate between Live Refinery, Identity Vault (future), Audit Logs (future) |
| **System Status** (bottom-left) | Shows Lead Shield (regex refinery) status and refinery version |
| **Metrics Bar** | Live count of entities scrubbed, Privacy ROI, Vault reuse rate, Deep Scan health |

### The Refinery Controls

- **Lead Shield (Regex)** — Toggle for structural PII detection (emails, phones, IBANs, addresses, URLs).
- **Deep AI Scrub (SLM)** — Toggle for contextual NER (names in prose, company names). Requires local AI to be running.
- **Operational Controls** — Direct access to Regex Enforcement, Identity Dictionaries, and Risk Compliance data.

### Step-by-Step: Run a Test

1. In the **Raw Input (Liability)** box on the left, paste any text containing personal data. Example:

   ```
   From: sarah.connor@cyberdyne.com
   Tel: +33 6 12 34 56 78
   IBAN: DE89370400440532013000
   Regards, Sarah Connor
   ```

2. Click **Execute Redaction**.

3. The **Clean Asset** panel on the right will instantly show redacted output:

   ```
   From: [EMAIL_a1b2c3d4]
   Tel: [PHONE_9f8e7d6c]
   IBAN: [IBAN_12ab34cd]
   Regards, [PERSON_5e4f3a2b]
   ```

4. The **Detection Attribution** panel appears below the output. For each entity found it shows:

   | Column | What it means |
   |--------|--------------|
   | Entity type | `EMAIL`, `PHONE`, `SSN`, `PERSON`, etc. |
   | Tier · Method | Which pipeline tier caught it and how (e.g. `Tier 1 · Rule Engine`, `Tier 1.5 · Greeting Shield`, `Tier 2 · AI NER`) |
   | `@offset` | Character position in the original text |
   | Confidence | Detection certainty (100% for rule-based, lower for contextual) |

   Example output for the text above:

   | Entity | Tier | Method | Location | Confidence |
   |--------|------|--------|----------|------------|
   | EMAIL | Tier 1 | Rule Engine | @6-30 | 100% |
   | PHONE | Tier 1.1 | Phone Shield | @36-55 | 100% |
   | IBAN | Tier 1 | Rule Engine · Validated | @62-84 | 100% |
   | PERSON | Tier 1.5 | Greeting Shield | @94-106 | 100% |

5. The **Secure Audit Trail** (right panel) shows the live Ed25519-signed log of vault events.

### Testing with Large Files

To process a large `.csv` or `.json` file directly through the refinery (bypassing the browser):

```bash
curl -F "file=@my_large_data.csv" http://localhost:8081/api/refine/file > cleaned_data.csv
```

OCULTAR streams the cleaned output directly into `cleaned_data.csv`.

---

## Part 4: Shut Down

When finished, run from your terminal (inside the `ocultar-community` folder):

```bash
docker compose down
```

This safely stops all containers and frees resources.

---

## Troubleshooting

| Symptom | Fix |
|---|---|
| Browser shows nothing at `localhost:3030` | Ensure Docker Desktop is running. Wait 30 seconds after setup, then refresh. |
| Dashboard loads but refinement fails | Check that the local AI container started: `docker compose logs ocultar-ai` |
| Windows PowerShell script blocked | Open PowerShell as Administrator, run `Set-ExecutionPolicy RemoteSigned`, then retry. |
### SharePoint Connector Environment Variables
> [!IMPORTANT]
> Pro Connectors (SharePoint/Teams, Slack) require an **Enterprise License** with the appropriate bitmask enabled.

- `MS_TENANT_ID`: Microsoft Entra (Azure AD) Tenant ID.
- `MS_CLIENT_ID`: Application (client) ID.
- `MS_CLIENT_SECRET`: Client secret.
- `MS_SHAREPOINT_SITE_ID`: (Optional) Target SharePoint site ID.
- `OCU_SALT`: (Required for Production) 16-uint8 hex salt used for HKDF key derivation.
