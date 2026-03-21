import json
import re
import sys

filename = sys.argv[1] if len(sys.argv) > 1 else "CLEANED_emails_test.json"
with open(filename, "r") as f:
    data = json.load(f)


# Extract all string values to search, or just traverse the whole dict/list
def get_all_strings(obj):
    strings = []
    if isinstance(obj, dict):
        for k, v in obj.items():
            strings.extend(get_all_strings(v))
    elif isinstance(obj, list):
        for item in obj:
            strings.extend(get_all_strings(item))
    elif isinstance(obj, str):
        strings.append(obj)
    return strings


texts = get_all_strings(data)

# Patterns
founder_refs = re.compile(r"(h[ée]ctor|eduardo|trejos)", re.IGNORECASE)
phone_refs = re.compile(r"(\+57|\+54|\+33)")
address_refs = re.compile(
    r"(grenoble|bogot[áa]|buenos aires|mallifaud|cra 13)", re.IGNORECASE
)

# Greetings/Signatures
greeting_refs = re.compile(
    r"(?:Regards|Best|Cheers|Bonjour|Hello|Hi|Dear|Sincerely|Cordialement)(?:,?)\s+([A-Z][a-z]+)",
    re.IGNORECASE,
)

# Tracking URLs (Letsignit, Calendly) query params
url_refs = re.compile(
    r"(letsignit|calendly)\.[a-z]+(/[^\s]*)?\?[^\s]*([a-zA-Z0-9_\-\.]+=[a-zA-Z0-9_\-\.]+)",
    re.IGNORECASE,
)
PII_PARAMS = {"name", "email", "user", "id", "fname", "lname", "firstname", "lastname"}

findings = {"founder": [], "phone": [], "address": [], "identity": [], "metadata": []}

for i, text in enumerate(texts):
    if founder_refs.search(text):
        for match in founder_refs.finditer(text):
            findings["founder"].append((i, match.group(0), text.replace("\n", " ")))

    if phone_refs.search(text):
        for match in phone_refs.finditer(text):
            findings["phone"].append((i, match.group(0), text.replace("\n", " ")))

    if address_refs.search(text):
        for match in address_refs.finditer(text):
            findings["address"].append((i, match.group(0), text.replace("\n", " ")))

    # Check greetings
    for match in greeting_refs.finditer(text):
        name = match.group(1)
        # Check if the name got replaced or is left unmasked.
        # A masked name might be something like NAME_ABC. But we look for unmasked.
        if name.upper() != name:  # Not fully uppercase (like a token)
            findings["identity"].append((i, name, text.replace("\n", " ")))

    # Check URLs
    for match in url_refs.finditer(text):
        url = match.group(0)
        # Parse query string for obvious PII keys or email addresses
        if "@" in url or any(p in url.lower() for p in ["name=", "email="]):
            findings["metadata"].append((i, url, text.replace("\n", " ")))

total_leaks = 0
for k, v in findings.items():
    print(f"=== {k.upper()} LEAKS: {len(v)} ===")
    for leak in v:
        print(leak)
    if k != "metadata":
        total_leaks += len(v)

if total_leaks > 0:
    print(f"FAIL: {total_leaks} critical leaks detected!")
    sys.exit(1)
else:
    print("PASS: 0 critical leaks detected.")
    sys.exit(0)
