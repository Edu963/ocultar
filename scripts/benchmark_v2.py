"""
Tier 2 benchmark v2: openai/privacy-filter
- Runs the original 6-string English test for comparison
- Runs a 20-sample from the French finance eval JSONL
- Computes detection rate and avg latency for each suite
"""
import json
import time
import random
from pathlib import Path
from transformers import pipeline

# ── model ──────────────────────────────────────────────────────────────────
print("Loading openai/privacy-filter (cached)...")
t0 = time.time()
classifier = pipeline(
    "token-classification",
    model="openai/privacy-filter",
    aggregation_strategy="simple",
)
print(f"Loaded in {time.time() - t0:.2f}s\n")

# ── helpers ────────────────────────────────────────────────────────────────
NER_TAG_MAP = {
    # CoNLL-style tags used in french_finance_eval.jsonl
    1:  "IBAN_B", 2: "IBAN_I", 3: "IBAN_E",
    17: "PER_B",  19: "PER_E",
    21: "PHONE_B", 22: "PHONE_I", 23: "PHONE_E",
    32: "COST_CENTER",
    33: "ORG_B", 35: "ORG_E",
    37: "AMT_B", 38: "AMT_I", 39: "AMT_E",
}
SENSITIVE_TAGS = {1, 2, 3, 17, 19, 21, 22, 23, 32, 33, 35}  # excludes plain amounts

def scan(text: str) -> tuple[list[dict], float]:
    t = time.time()
    raw = classifier(text)
    ms = (time.time() - t) * 1000
    return [{"label": e["entity_group"], "value": e["word"], "score": float(e["score"])} for e in raw], ms

def has_any_sensitive(ner_tags: list[int]) -> bool:
    return any(t in SENSITIVE_TAGS for t in ner_tags)

# ── suite 1: original English strings ──────────────────────────────────────
ENGLISH_TESTS = [
    ("Transfer €84,293 from IBAN FR76 3000 6000 0112 3456 7890 189",
     ["FR76 3000 6000 0112 3456 7890 189"]),
    ("Vendor payment to Acme Corp, account 4532015112830366, approved by John Smith",
     ["Acme Corp", "4532015112830366", "John Smith"]),
    ("Cost center 4420-EMEA-CORP, GL account 6100, controller Sarah Chen",
     ["4420-EMEA-CORP", "Sarah Chen"]),
    ("Invoice INV-2026-00847 for Société Générale, SWIFT SOGEFRPP",
     ["Société Générale", "SOGEFRPP"]),
    ("Board resolution: CEO approved $2.3M acquisition, confidential until April 30",
     ["April 30"]),
    ("john.doe@company.fr called +33 6 12 34 56 78 re: Q1 close",
     ["john.doe@company.fr", "+33 6 12 34 56 78"]),
]

en_hits = 0
en_total = sum(len(exp) for _, exp in ENGLISH_TESTS)
en_latencies = []
en_rows = []

for text, expected in ENGLISH_TESTS:
    entities, ms = scan(text)
    en_latencies.append(ms)
    detected_values = " ".join(e["value"].strip() for e in entities).lower()
    row_hits = sum(1 for exp in expected if any(word.lower() in detected_values for word in exp.split()))
    en_hits += row_hits
    en_rows.append({
        "text": text,
        "expected": expected,
        "detected": entities,
        "hits": row_hits,
        "of": len(expected),
        "latency_ms": round(ms, 1),
    })

# ── suite 2: French finance eval (20 samples) ──────────────────────────────
eval_path = Path(__file__).parent.parent / "data/fine-tune/french_finance_eval.jsonl"
fr_samples = []
with eval_path.open() as f:
    for line in f:
        obj = json.loads(line)
        if has_any_sensitive(obj["ner_tags"]):
            fr_samples.append(obj)

random.seed(42)
fr_subset = random.sample(fr_samples, min(20, len(fr_samples)))

fr_detected = 0   # rows where model found ≥1 entity
fr_total = len(fr_subset)
fr_latencies = []
fr_rows = []

for obj in fr_subset:
    text = obj["text"]
    entities, ms = scan(text)
    fr_latencies.append(ms)
    found = len(entities) > 0
    if found:
        fr_detected += 1
    fr_rows.append({
        "text": text,
        "ner_tags": obj["ner_tags"],
        "detected": entities,
        "found_something": found,
        "latency_ms": round(ms, 1),
    })

# ── output ─────────────────────────────────────────────────────────────────
out = {
    "english_suite": {
        "detection_rate": f"{en_hits}/{en_total} ({100*en_hits//en_total}%)",
        "avg_latency_ms": round(sum(en_latencies) / len(en_latencies), 1),
        "rows": en_rows,
    },
    "french_finance_suite": {
        "rows_with_detection": f"{fr_detected}/{fr_total} ({100*fr_detected//fr_total}%)",
        "avg_latency_ms": round(sum(fr_latencies) / len(fr_latencies), 1),
        "rows": fr_rows,
    },
}

out_path = Path(__file__).parent.parent / "privacy_filter_results_v2.json"
with out_path.open("w") as f:
    json.dump(out, f, indent=2, ensure_ascii=False)

print("=" * 60)
print("ENGLISH SUITE")
print(f"  Detection rate : {out['english_suite']['detection_rate']}")
print(f"  Avg latency    : {out['english_suite']['avg_latency_ms']} ms")
print()
print("FRENCH FINANCE SUITE (20 samples, sensitive rows only)")
print(f"  Rows detected  : {out['french_finance_suite']['rows_with_detection']}")
print(f"  Avg latency    : {out['french_finance_suite']['avg_latency_ms']} ms")
print()

print("Detail — French rows:")
for r in fr_rows:
    status = "✓" if r["found_something"] else "✗"
    labels = ", ".join(f"{e['label']}:{e['value'].strip()}" for e in r["detected"]) or "(nothing)"
    print(f"  {status} {r['text'][:70]}")
    print(f"       → {labels}")

print(f"\nFull results → privacy_filter_results_v2.json")
