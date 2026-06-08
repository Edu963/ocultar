import os
import json
import re
import requests
import time
from datasets import load_dataset
from dotenv import load_dotenv
from tqdm import tqdm

# Load environment variables
load_dotenv()
HF_TOKEN = os.getenv("HF_TOKEN")
REFINERY_URL = "http://localhost:8090/api/refine"

# Mapping ai4privacy categories to OCULTAR entities
CATEGORY_MAP = {
    "EMAIL": "EMAIL",
    "SOCIALNUMBER": "SSN",
    "IDCARD": "SSN",
    "PASS": "SSN",
    "FIRSTNAME": "PERSON",
    "LASTNAME1": "PERSON",
    "LASTNAME2": "PERSON",
    "USERNAME": "PERSON",
    "STREET": "ADDRESS",
    "CITY": "ADDRESS",
    "STATE": "ADDRESS",
    "COUNTRY": "ADDRESS",
    "BUILDING": "ADDRESS",
    "POSTCODE": "ADDRESS",
    "IP_ADDRESS": "IP_ADDRESS",
    "CREDIT_CARD": "CREDIT_CARD",
    "API_KEY": "SECRET",
    "PASSWORD": "CREDENTIAL",
    "URL": "URL",
    "PHONE": "PHONE",
    "DATE": "DATE",
    "TIME": "TIME",
}

# Reverse map: OCULTAR entity -> set of ai4privacy categories
REVERSE_MAP = {}
for cat, ocu in CATEGORY_MAP.items():
    REVERSE_MAP.setdefault(ocu, set()).add(cat)

# Token pattern used by OCULTAR for redaction
TOKEN_RE = re.compile(r'\[([A-Z_]+)_[0-9a-f]{8}\]')

def refine_text(text):
    try:
        resp = requests.post(REFINERY_URL, json={"text": text}, timeout=30)
        if resp.status_code == 200:
            return resp.json()
    except Exception as e:
        print(f"Error calling refinery: {e}")
    return None

def evaluate():
    print("Loading ai4privacy/pii-masking-300k dataset (English only)...")
    ds = load_dataset("ai4privacy/pii-masking-300k", split="train", streaming=True)
    
    samples = []
    max_samples = 200
    for item in ds:
        lang = item.get("language", "")
        if lang.lower().startswith("en") or lang.lower() == "english":
            samples.append(item)
        if len(samples) >= max_samples:
            break

            
    print(f"Processing {len(samples)} English samples...")
    
    overall_metrics = {cat: {"tp": 0, "fp": 0, "fn": 0} for cat in CATEGORY_MAP.keys()}
    debug_misses = []  # Track missed detections for analysis
    
    for i, sample in enumerate(tqdm(samples)):
        text = sample["source_text"]
        
        # Parse ground truth
        try:
            if isinstance(sample["span_labels"], str):
                ground_truth = json.loads(sample["span_labels"])
            else:
                ground_truth = sample["span_labels"]
        except Exception as e:
            continue

        # Call OCULTAR
        refine_resp = refine_text(text)
        if not refine_resp:
            continue
        
        refined_text = refine_resp.get("refined", "")
        # The refined text is wrapped in JSON, extract it
        try:
            refined_inner = json.loads(refined_text)
            if isinstance(refined_inner, dict) and "text" in refined_inner:
                refined_text = refined_inner["text"]
        except:
            pass
            
        detected_hits = refine_resp.get("report", {}).get("pii_hits", [])
        
        # --- TEXT-BASED MATCHING ---
        # For each ground truth span, check if its text was redacted in the output
        gt_matched = set()
        
        for gt_idx, gt in enumerate(ground_truth):
            gt_start, gt_end, gt_cat = gt
            
            if gt_cat not in CATEGORY_MAP:
                continue
            
            gt_text = text[gt_start:gt_end].strip()
            if len(gt_text) < 2:
                continue
                
            # Check if the original text was replaced by a token in the refined output
            was_redacted = gt_text not in refined_text
            
            if was_redacted:
                overall_metrics[gt_cat]["tp"] += 1
                gt_matched.add(gt_idx)
            else:
                overall_metrics[gt_cat]["fn"] += 1
                if i < 10:  # Debug first 10 samples
                    debug_misses.append({
                        "sample": i,
                        "category": gt_cat,
                        "text": gt_text,
                        "ocu_entity": CATEGORY_MAP[gt_cat]
                    })
        
        # --- FALSE POSITIVE CHECK ---
        # Count tokens in refined text that don't correspond to any ground truth span
        found_tokens = TOKEN_RE.findall(refined_text)
        for token_entity in found_tokens:
            # Check if this token entity type maps to any ground truth category
            is_tp = False
            if token_entity in REVERSE_MAP:
                for gt in ground_truth:
                    gt_start, gt_end, gt_cat = gt
                    if gt_cat in REVERSE_MAP.get(token_entity, set()):
                        gt_text = text[gt_start:gt_end].strip()
                        if gt_text not in refined_text:
                            is_tp = True
                            break
            
            if not is_tp:
                # This is a false positive — find the first matching category
                for cat, ocu_type in CATEGORY_MAP.items():
                    if token_entity == ocu_type:
                        overall_metrics[cat]["fp"] += 1
                        break

    # Calculate final metrics
    final_report = {}
    total_tp = total_fp = total_fn = 0
    
    for cat, m in overall_metrics.items():
        tp, fp, fn = m["tp"], m["fp"], m["fn"]
        if tp == 0 and fp == 0 and fn == 0:
            continue
        
        total_tp += tp
        total_fp += fp
        total_fn += fn
            
        precision = tp / (tp + fp) if (tp + fp) > 0 else 0
        recall = tp / (tp + fn) if (tp + fn) > 0 else 0
        f1 = 2 * (precision * recall) / (precision + recall) if (precision + recall) > 0 else 0
        
        final_report[cat] = {
            "precision": round(precision, 4),
            "recall": round(recall, 4),
            "f1_score": round(f1, 4),
            "tp": tp,
            "fp": fp,
            "fn": fn
        }
    
    # Overall weighted metrics
    overall_precision = total_tp / (total_tp + total_fp) if (total_tp + total_fp) > 0 else 0
    overall_recall = total_tp / (total_tp + total_fn) if (total_tp + total_fn) > 0 else 0
    overall_f1 = 2 * (overall_precision * overall_recall) / (overall_precision + overall_recall) if (overall_precision + overall_recall) > 0 else 0

    # Save results to temp (git ignored)
    os.makedirs("temp", exist_ok=True)
    output = {
        "per_category": final_report,
        "overall": {
            "precision": round(overall_precision, 4),
            "recall": round(overall_recall, 4),
            "f1_score": round(overall_f1, 4),
            "total_tp": total_tp,
            "total_fp": total_fp,
            "total_fn": total_fn,
        },
        "debug_misses_sample": debug_misses[:20],
        "config": {
            "dataset": "ai4privacy/pii-masking-300k",
            "samples": len(samples),
            "language": "en",
            "model": "openai/privacy-filter (base)",
            "tier": "enterprise",
            "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
        }
    }
    with open("temp/ai4privacy_results.json", "w") as f:
        json.dump(output, f, indent=2)
        
    print("\n" + "=" * 65)
    print("  OCULTAR PII Benchmark — ai4privacy/pii-masking-300k")
    print("=" * 65)
    print(f"  Samples: {len(samples)} | Language: EN | Tier: Enterprise")
    print(f"  Model: openai/privacy-filter (base)")
    print("-" * 65)
    print(f"{'Category':<15} | {'Prec':>8} | {'Recall':>8} | {'F1':>8} | {'TP':>4} | {'FP':>4} | {'FN':>4}")
    print("-" * 65)
    for cat, metrics in sorted(final_report.items(), key=lambda x: x[1]['f1_score'], reverse=True):
        print(f"{cat:<15} | {metrics['precision']:>8.4f} | {metrics['recall']:>8.4f} | {metrics['f1_score']:>8.4f} | {metrics['tp']:>4} | {metrics['fp']:>4} | {metrics['fn']:>4}")
    print("-" * 65)
    print(f"{'OVERALL':<15} | {overall_precision:>8.4f} | {overall_recall:>8.4f} | {overall_f1:>8.4f} | {total_tp:>4} | {total_fp:>4} | {total_fn:>4}")
    print("=" * 65)
    
    if debug_misses:
        print(f"\n--- Sample Misses (first {min(10, len(debug_misses))}) ---")
        for m in debug_misses[:10]:
            print(f"  [{m['category']}] \"{m['text']}\" (expected: {m['ocu_entity']})")
    
    print(f"\nFull results saved to temp/ai4privacy_results.json")

if __name__ == "__main__":
    evaluate()
