# OCULTAR | Setup & Usage Guide

This guide covers how to package OCULTAR for a client, how to run it locally, and how to use the dashboard. Written for non-technical users and hands-on evaluators alike.

---

## Part 1: For the Sender (Packaging for a Client)

**Your Goal:** Package OCULTAR's Community Edition and send it to a client.

### Step 1 — Build the release

Open a terminal in the `ocultar/` folder and run:

```bash
./build_release.sh
```

This compiles and packages both editions. It creates:
- `dist/ocultar-community.zip` — for Community clients
- `dist/ocultar-enterprise.tar.gz` — for Enterprise clients

### Step 2 — Send the right file

Send only `ocultar-community.zip` to your Community client. Use WeTransfer, Google Drive, or SharePoint if the file is too large for email.

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

**Mac / Linux (Community & Enterprise):**
1. Open your Terminal.
2. Drag and drop `scripts/setup-community.sh` (or `setup-enterprise.sh` for the Enterprise edition) into the Terminal and press **Enter**.
3. The script handles all keys, networking, and build steps for you.

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

### 2. Pilot "Proof of Value" (PoV)
Every Enterprise pilot follows a structured 14-day lifecycle (managed by the `Pilot-Manager` skill):
- **Days 1-3 (Onboarding)**: Auto-rotation of `OCU_MASTER_KEY` and sync of local VIP dictionaries.
- **Days 7-10 (Quantification)**: Running the **ROI Accountant** against your sample audit logs to map blocked PII to financial risk.
- **Day 14 (Hardening)**: Generating the verifiable **Auditor Report** for your SOC2/HIPAA stakeholders.

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

4. The **Egress Feed** (the live log panel) shows internal refinery events:
   - `REF_MATCH` — a PII type was intercepted
   - `VAULT_SAVE` — a new encrypted token was stored
   - `VAULT_LOCK` — the vault security handshake

### Testing with Large Files

To process a large `.csv` or `.json` file directly through the refinery (bypassing the browser):

```bash
curl -F "file=@my_large_data.csv" http://localhost:8080/api/refine/file > cleaned_data.csv
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
| Browser shows nothing at `localhost:3000` | Ensure Docker Desktop is running. Wait 30 seconds after setup, then refresh. |
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
