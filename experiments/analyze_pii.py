"""
Multi-Framework Compliance Exposure Report — sdo_report.json (Sovereign Data Object)
Frameworks: GDPR · AI Act · BSI C5 · NIS2 · HIPAA · ISO 27001 · DTX
"""

import re
from collections import defaultdict

# ── Load & repair ──────────────────────────────────────────────────────────────
with open("sdo_report.json", "rb") as f:
    raw = f.read()

raw = re.sub(rb"\\r", b" ", raw)
raw = re.sub(rb"\\n", b" ", raw)
raw = re.sub(rb"\\([^\"\\\/bfnrtu0-9])", rb"\1", raw)
text = raw.decode("utf-8", errors="replace")

# ── Token placeholder ──────────────────────────────────────────────────────────
TOKEN = re.compile(r"\[[A-Z_]+_[0-9a-f]{4,}\]")


def is_token(s):
    return bool(TOKEN.fullmatch(s.strip()))


# ── PII patterns ───────────────────────────────────────────────────────────────
PATTERNS = {
    "Email address": re.compile(r"[a-zA-Z0-9._%+\-]{2,}@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}"),
    "Phone (intl)": re.compile(
        r"\+\d{1,3}[\s.\-]?\(?\d{1,4}\)?[\s.\-]?\d{2,4}[\s.\-]?\d{2,4}(?:[\s.\-]?\d{2,4})?"
    ),
    "Phone (local FR)": re.compile(r"(?<!\d)0[1-9](?:[\s.\-]?\d{2}){4}(?!\d)"),
    "URL / web link": re.compile(
        r'https?://[^\s"<>{}\[\]\\]+|(?<![a-zA-Z])www\.[a-zA-Z0-9\-]+\.[a-zA-Z]{2,}[^\s"<>{}\[\]\\]*'
    ),
    "LinkedIn / social": re.compile(
        r'linkedin\.com/[^\s"<>{}\[\]\\]+|twitter\.com/[^\s"<>{}\[\]\\]+|facebook\.com/[^\s"<>{}\[\]\\]+',
        re.I,
    ),
    "Street address": re.compile(
        r"\b\d{1,5}[\s,]+(?:rue|avenue|ave|boulevard|blvd|street|st\.|road|rd\.|calle|carrera|piso|paseo|chemin|all[eé]e|impasse|square|plaza|via\b)\b.{0,80}",
        re.I,
    ),
    "Postal code (FR/ES/AR)": re.compile(
        r"(?<!\d)(?:0[1-9]|[1-8]\d|9[0-8])\d{3}(?!\d)"
    ),
    "Full name (bare)": re.compile(
        r"\b[A-ZÁÉÍÓÚÜÑ][a-záéíóúüñ]{2,}(?:\s+[A-ZÁÉÍÓÚÜÑ][a-záéíóúüñ]{1,}){1,2}\b"
    ),
    "Company (legal form)": re.compile(
        r"\b[A-ZÁÉÍÓÚÜÑ][A-Za-záéíóúüñ&\s]{2,35}(?:S\.A\.|S\.L\.|SRL|SARL|GmbH|LLC|Ltd\.?|Inc\.?|SAS|B\.V\.|N\.V\.|PLC|SpA)\b"
    ),
    "IP address": re.compile(r"\b(?:\d{1,3}\.){3}\d{1,3}\b"),
    "IBAN": re.compile(r"\b[A-Z]{2}\d{2}[A-Z0-9]{4}\d{7}[A-Z0-9]{0,16}\b"),
    "Credit card": re.compile(r"(?<!\d)(?:\d{4}[\s\-]?){3}\d{4}(?!\d)"),
    "National ID / passport": re.compile(r"\b[A-Z]{1,2}\d{7,9}\b"),
    "Health / medical data": re.compile(
        r"\b(?:diagnosis|diagnóstico|patient|paciente|medical|clínica|hospital|treatment|tratamiento|prescription|receta|ICD-?\d{2}|DSM-?\d)\b",
        re.I,
    ),
}

NAME_SKIP = re.compile(
    r"^(Bonjour|Cordialement|Bonsoir|Regards|Thanks|Hello|Dear|Salut|Hola|Buenos|Gracias|"
    r"Merci|Bien|Votre|Notre|Pour|Avec|Objet|Subject|Date|From|Reply|Sent|Message|"
    r"Confidential|Please|Also|Best|Kind|Warm|Good|Monday|Tuesday|Wednesday|Thursday|"
    r"Friday|Saturday|Sunday|January|February|March|April|May|June|July|August|September|"
    r"October|November|December|Lunes|Martes|Miercoles|Jueves|Viernes)",
    re.I,
)

# ── Compliance mapping ─────────────────────────────────────────────────────────
# For each PII type, list which framework articles / controls are implicated
COMPLIANCE = {
    "Email address": {
        "GDPR": "Art. 4(1) personal data · Art. 5 purpose limitation",
        "AI Act": "Art. 10 data governance (training/test data)",
        "C5": "OIS-04 data classification",
        "NIS2": "Art. 21 security of processing",
        "HIPAA": "§164.514 de-identification (if health context)",
        "ISO27001": "A.8.11 data masking · A.5.34 privacy",
        "DTX": "§6 data minimisation obligation",
    },
    "Phone (intl)": {
        "GDPR": "Art. 4(1) · Art. 9 (if health-related)",
        "C5": "OIS-04",
        "NIS2": "Art. 21",
        "ISO27001": "A.8.11",
        "DTX": "§6",
    },
    "Phone (local FR)": {
        "GDPR": "Art. 4(1) · Art. 5",
        "C5": "OIS-04",
        "ISO27001": "A.8.11",
        "DTX": "§6",
    },
    "Full name (bare)": {
        "GDPR": "Art. 4(1) direct identifier · Art. 17 right to erasure",
        "AI Act": "Art. 10 · Recital 47 (profiling risk)",
        "C5": "OIS-04",
        "NIS2": "Art. 21",
        "HIPAA": "§164.514(b)(2) Safe Harbor — name is listed identifier",
        "ISO27001": "A.8.11 · A.5.34",
        "DTX": "§4 data subject rights · §6",
    },
    "URL / web link": {
        "GDPR": "Art. 4(1) indirect identifier (company/person URL)",
        "C5": "OIS-04 · SIM-01 supply chain",
        "NIS2": "Art. 21 access control",
        "ISO27001": "A.8.23 web filtering",
    },
    "LinkedIn / social": {
        "GDPR": "Art. 4(1) direct identifier · Art. 9(2) (political/religious risk)",
        "AI Act": "Art. 10 · Annex III prohibited biometric categorisation risk",
        "C5": "OIS-04",
        "ISO27001": "A.5.34 · A.8.11",
        "DTX": "§4 · §6",
    },
    "Street address": {
        "GDPR": "Art. 4(1) · Art. 4(15) location data",
        "C5": "OIS-04",
        "NIS2": "Art. 21",
        "HIPAA": "§164.514(b)(2) — geographic data smaller than state",
        "ISO27001": "A.8.11 · A.5.34",
        "DTX": "§6",
    },
    "Postal code (FR/ES/AR)": {
        "GDPR": "Art. 4(1) (when combined with other data)",
        "HIPAA": "§164.514(b)(2) — 3-digit geographic unit rule",
        "ISO27001": "A.8.11",
    },
    "Company (legal form)": {
        "GDPR": "Art. 4(1) — if data pertains to a natural person in context",
        "C5": "SIM-01 supply-chain / third-party disclosure",
        "NIS2": "Art. 21 third-party risk",
        "ISO27001": "A.5.19 supplier relationships",
    },
    "IP address": {
        "GDPR": "Art. 4(1) — ECJ C-582/14 (Breyer): IP = personal data",
        "C5": "OIS-04 · BCM-01",
        "NIS2": "Art. 21",
        "ISO27001": "A.8.15 logging · A.8.11",
    },
    "IBAN": {
        "GDPR": "Art. 4(1) financial personal data",
        "C5": "OIS-04 · COS-01",
        "HIPAA": "§164.514(b)(2) — account numbers listed",
        "ISO27001": "A.8.11 · A.5.34",
        "DTX": "§6 · §8 financial data category",
    },
    "Credit card": {
        "GDPR": "Art. 4(1) financial personal data",
        "C5": "COS-01",
        "HIPAA": "§164.514(b)(2) — account numbers",
        "ISO27001": "A.8.11 PCI-DSS alignment",
        "DTX": "§8",
    },
    "National ID / passport": {
        "GDPR": "Art. 87 national ID numbers — Member State law",
        "C5": "OIS-04",
        "HIPAA": "§164.514(b)(2) — listed identifier",
        "ISO27001": "A.8.11 · A.5.34",
        "DTX": "§6 · §9 special category",
    },
    "Health / medical data": {
        "GDPR": "Art. 9(1) special category — EXPLICIT CONSENT required",
        "AI Act": "Art. 6 high-risk AI system (health) · Annex III",
        "C5": "OIS-04 high sensitivity",
        "NIS2": "Art. 21 — healthcare sector essential entity",
        "HIPAA": "§164.502 PHI — ALL 18 identifiers must be stripped",
        "ISO27001": "A.8.11 · A.5.34 · healthcare annex",
        "DTX": "§9 sensitive data category — enhanced obligations",
    },
}

FRAMEWORKS = ["GDPR", "AI Act", "C5", "NIS2", "HIPAA", "ISO27001", "DTX"]

# ── Scan ──────────────────────────────────────────────────────────────────────
STRING_RE = re.compile(r'"(?:[^"\\]|\\.)*"')
leaks = defaultdict(set)
contexts = defaultdict(dict)

for m in STRING_RE.finditer(text):
    val = m.group(0)[1:-1]
    val = re.sub(r"\\[rn]", " ", val)
    val = re.sub(r"\\(.)", r"\1", val)
    for label, pat in PATTERNS.items():
        for hit in pat.finditer(val):
            found = hit.group(0).strip()
            if not found or len(found) < 3 or is_token(found):
                continue
            if label == "Full name (bare)" and NAME_SKIP.match(found):
                continue
            if label.startswith("Postal") and re.fullmatch(r"(19|20)\d{2}", found):
                continue
            leaks[label].add(found)
            if found not in contexts[label]:
                s = max(0, hit.start() - 40)
                contexts[label][found] = val[s : hit.end() + 40].replace("\n", " ")

# ── Report ────────────────────────────────────────────────────────────────────
total = sum(len(v) for v in leaks.values())
RISK = {  # severity per category
    "Health / medical data": "🔴 CRITICAL",
    "IBAN": "🔴 CRITICAL",
    "Credit card": "🔴 CRITICAL",
    "National ID / passport": "🔴 CRITICAL",
    "Full name (bare)": "🟠 HIGH",
    "Email address": "🟠 HIGH",
    "Street address": "🟠 HIGH",
    "LinkedIn / social": "🟠 HIGH",
    "Phone (intl)": "🟡 MEDIUM",
    "Phone (local FR)": "🟡 MEDIUM",
    "Company (legal form)": "🟡 MEDIUM",
    "URL / web link": "🟡 MEDIUM",
    "Postal code (FR/ES/AR)": "🟢 LOW",
    "IP address": "🟡 MEDIUM",
}

print("=" * 72)
print("  MULTI-FRAMEWORK COMPLIANCE EXPOSURE REPORT — sdo_report.json (SDO)")
print(f"  Frameworks: {' · '.join(FRAMEWORKS)}")
print("=" * 72)
print(f"  Leak categories : {len(leaks)}   |   Unique PII values : {total}")
print("=" * 72)

# Summary matrix
print("\n  RISK MATRIX\n")
header = f"  {'Category':<30}" + "".join(f"{fw:>9}" for fw in FRAMEWORKS)
print(header)
print("  " + "-" * (30 + 9 * len(FRAMEWORKS)))
for cat, vals in sorted(leaks.items(), key=lambda x: -len(x[1])):
    cm = COMPLIANCE.get(cat, {})
    row = f"  {cat:<30}"
    for fw in FRAMEWORKS:
        row += f"{'  ⚠ ':>9}" if fw in cm else f"{'  -':>9}"
    print(row + f"   ({len(vals)} hits)")

# Detail per category
PRIORITY_ORDER = [
    "Health / medical data",
    "IBAN",
    "Credit card",
    "National ID / passport",
    "Full name (bare)",
    "Email address",
    "Street address",
    "LinkedIn / social",
    "Phone (intl)",
    "Phone (local FR)",
    "IP address",
    "Company (legal form)",
    "URL / web link",
    "Postal code (FR/ES/AR)",
]
ordered = [c for c in PRIORITY_ORDER if c in leaks] + [
    c for c in sorted(leaks) if c not in PRIORITY_ORDER
]

for cat in ordered:
    vals = sorted(leaks[cat])
    risk = RISK.get(cat, "🟡 MEDIUM")
    cm = COMPLIANCE.get(cat, {})
    print(f"\n{'─'*72}")
    print(f"  {risk}  {cat}  [{len(vals)} unique value(s)]")
    print(f"{'─'*72}")
    # Compliance references
    for fw in FRAMEWORKS:
        if fw in cm:
            print(f"  [{fw}]  {cm[fw]}")
    print()
    # Sample values
    for v in vals[:15]:
        ctx = contexts[cat].get(v, "")
        print(f"  • {v!r}")
        if ctx:
            print(f"    ↳ …{ctx}…")
    if len(vals) > 15:
        print(f"\n  [+{len(vals)-15} more unique values not shown]")

print(f"\n{'='*72}")
print("  END OF REPORT")
print(f"{'='*72}\n")
