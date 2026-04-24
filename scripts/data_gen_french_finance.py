import json
import random
import re

# Label mapping (BIOES)
LABELS = {
    "O": 0,
    "B-account_number": 1, "I-account_number": 2, "E-account_number": 3, "S-account_number": 4,
    "B-private_address": 5, "I-private_address": 6, "E-private_address": 7, "S-private_address": 8,
    "B-private_date": 9, "I-private_date": 10, "E-private_date": 11, "S-private_date": 12,
    "B-private_email": 13, "I-private_email": 14, "E-private_email": 15, "S-private_email": 16,
    "B-private_person": 17, "I-private_person": 18, "E-private_person": 19, "S-private_person": 20,
    "B-private_phone": 21, "I-private_phone": 22, "E-private_phone": 23, "S-private_phone": 24,
    "B-private_url": 25, "I-private_url": 26, "E-private_url": 27, "S-private_url": 28,
    "B-secret": 29, "I-secret": 30, "E-secret": 31, "S-secret": 32,
    "B-organization": 33, "I-organization": 34, "E-organization": 35, "S-organization": 36,
    "B-amount": 37, "I-amount": 38, "E-amount": 39, "S-amount": 40,
}

# Generators
def gen_iban():
    chars = "0123456789"
    bank = "".join(random.choices(chars, k=5))
    branch = "".join(random.choices(chars, k=5))
    account = "".join(random.choices(chars + "ABCDEFGHIJKLMNOPQRSTUVWXYZ", k=11))
    key = "".join(random.choices(chars, k=2))
    return f"FR76 {bank} {branch} {account} {key}"

def gen_siret():
    return "".join(random.choices("0123456789", k=14))

def gen_siren():
    return "".join(random.choices("0123456789", k=9))

def gen_phone():
    prefixes = ["01", "02", "03", "04", "05", "06", "07", "09"]
    p = random.choice(prefixes)
    rest = " ".join(["".join(random.choices("0123456789", k=2)) for _ in range(4)])
    return f"{p} {rest}"

def gen_amount():
    val = random.randint(10, 100000)
    fmt = random.choice(["€", "EUR", "Euros"])
    if random.random() > 0.5:
        return f"{val:,} {fmt}".replace(",", " ")
    return f"{val}.{random.randint(0, 99):02d} {fmt}"

def gen_cost_center():
    depts = ["EMEA", "APAC", "US", "FR", "CORP"]
    return f"{random.randint(1000, 9999)}-{random.choice(depts)}-{random.randint(100, 999)}"

def gen_company():
    names = ["Acme", "Logistics", "Services", "Finance", "Tech", "Global", "Solutions", "Industries"]
    suffix = random.choice(["SARL", "SAS", "SA", "EURL", "SCI"])
    return f"{random.choice(names)} {suffix}"

def gen_person():
    firsts = ["Jean", "Marie", "Pierre", "Sophie", "Lucas", "Emma", "Thomas", "Lea"]
    lasts = ["Dupont", "Martin", "Bernard", "Dubois", "Thomas", "Robert", "Richard", "Petit"]
    return f"{random.choice(firsts)} {random.choice(lasts)}"

# Templates
TEMPLATES = [
    "Virement de {amount} vers l'IBAN {iban} pour {person}.",
    "Facture {id} de {company} (SIRET {siret}).",
    "Paiement fournisseur {company}, compte {iban}, montant {amount}.",
    "Note de frais de {person}, centre de coût {cost_center}.",
    "Contactez {person} au {phone} pour le dossier {id}.",
    "Règlement de {amount} reçu de {company} via {iban}.",
    "Le SIREN de l'entreprise est {siren}.",
    "Transfert interne vers {cost_center} pour un montant de {amount}.",
]

def tokenize_and_tag(text, entities):
    # This is a simplified tokenizer. In real NER, you'd use the model's tokenizer.
    # We'll split by space and keep track of indices.
    tokens = text.split()
    ner_tags = [LABELS["O"]] * len(tokens)
    
    for entity_text, label in entities:
        entity_tokens = entity_text.split()
        # Find where this sequence of tokens appears in the main tokens list
        for i in range(len(tokens) - len(entity_tokens) + 1):
            if tokens[i:i+len(entity_tokens)] == entity_tokens:
                if len(entity_tokens) == 1:
                    ner_tags[i] = LABELS[f"S-{label}"]
                else:
                    ner_tags[i] = LABELS[f"B-{label}"]
                    for j in range(1, len(entity_tokens) - 1):
                        ner_tags[i+j] = LABELS[f"I-{label}"]
                    ner_tags[i+len(entity_tokens)-1] = LABELS[f"E-{label}"]
                break
    return tokens, ner_tags

def generate_example():
    t = random.choice(TEMPLATES)
    data = {
        "amount": gen_amount(),
        "iban": gen_iban(),
        "person": gen_person(),
        "company": gen_company(),
        "siret": gen_siret(),
        "siren": gen_siren(),
        "cost_center": gen_cost_center(),
        "phone": gen_phone(),
        "id": f"INV-{random.randint(2000, 2100)}-{random.randint(10000, 99999)}"
    }
    
    # Fill template
    text = t.format(**data)
    
    # Track entities
    entities = []
    if "{amount}" in t: entities.append((data["amount"], "amount"))
    if "{iban}" in t: entities.append((data["iban"], "account_number"))
    if "{person}" in t: entities.append((data["person"], "private_person"))
    if "{company}" in t: entities.append((data["company"], "organization"))
    if "{siret}" in t: entities.append((data["siret"], "account_number"))
    if "{siren}" in t: entities.append((data["siren"], "account_number"))
    if "{cost_center}" in t: entities.append((data["cost_center"], "secret"))
    if "{phone}" in t: entities.append((data["phone"], "private_phone"))
    
    tokens, ner_tags = tokenize_and_tag(text, entities)
    return {"tokens": tokens, "ner_tags": ner_tags, "text": text}

def save_dataset(filename, count):
    with open(filename, "w", encoding="utf-8") as f:
        for _ in range(count):
            ex = generate_example()
            f.write(json.dumps(ex, ensure_ascii=False) + "\n")

if __name__ == "__main__":
    import os
    os.makedirs("data/fine-tune", exist_ok=True)
    save_dataset("data/fine-tune/french_finance_train.jsonl", 200)
    save_dataset("data/fine-tune/french_finance_eval.jsonl", 50)
    print("Datasets generated successfully.")
