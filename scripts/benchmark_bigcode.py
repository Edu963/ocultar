import os
import json
import requests
import time
from datasets import load_dataset
from dotenv import load_dotenv
from tqdm import tqdm

# Load environment variables
load_dotenv()
HF_TOKEN = os.getenv("HF_TOKEN")
REFINERY_URL = "http://localhost:8090/api/refine"

# Mapping BigCode categories to OCULTAR entities
CATEGORY_MAP = {
    "Emails": "EMAIL",
    "IP addresses": "IP_ADDRESS",
    "Keys": "SECRET",
    "Passwords": "CREDENTIAL",
    "IDs": "SSN", # Approximation
    "Names": "PERSON",
    "Usernames": "PERSON"
}

def refine_text(text):
    try:
        resp = requests.post(REFINERY_URL, json={"text": text}, timeout=10)
        if resp.status_code == 200:
            return resp.json()
    except Exception as e:
        print(f"Error calling refinery: {e}")
    return None

def evaluate():
    print("Loading dataset...")
    # Use a small subset for testing
    ds = load_dataset("bigcode/bigcode-pii-dataset", split="train", streaming=True, token=HF_TOKEN)
    
    samples = []
    count = 0
    max_samples = 50
    for item in ds:
        samples.append(item)
        count += 1
        if count >= max_samples:
            break
            
    print(f"Processing {len(samples)} samples...")
    results = []
    
    overall_metrics = {cat: {"tp": 0, "fp": 0, "fn": 0} for cat in CATEGORY_MAP.keys()}
    
    for i, sample in enumerate(tqdm(samples)):
        text = sample["text"]
        ground_truth = sample["fragments"]
        
        # Call OCULTAR
        refine_resp = refine_text(text)
        if not refine_resp:
            continue
            
        detected_hits = refine_resp.get("report", {}).get("pii_hits", [])
        
        # Evaluation logic
        # We check if each ground truth fragment was detected by OCULTAR
        for gt in ground_truth:
            gt_cat = gt["category"]
            gt_val = gt["value"]
            gt_start = gt["position"][0]
            gt_end = gt["position"][1]
            
            if gt_cat not in CATEGORY_MAP:
                continue
                
            ocu_entity = CATEGORY_MAP[gt_cat]
            
            # Look for a match in detected_hits
            found = False
            for hit in detected_hits:
                # OCULTAR location is "start-end"
                try:
                    loc = hit["location"].split("-")
                    hit_start = int(loc[0])
                    hit_end = int(loc[1])
                    
                    # Check for overlap
                    if max(gt_start, hit_start) < min(gt_end, hit_end):
                        # Matching entity type (approximate)
                        if hit["entity"] == ocu_entity or (ocu_entity == "SECRET" and hit["entity"] == "CREDENTIAL"):
                            found = True
                            break
                except:
                    continue
            
            if found:
                overall_metrics[gt_cat]["tp"] += 1
            else:
                overall_metrics[gt_cat]["fn"] += 1
                
        # Simple FP check: detected hits that don't overlap with any ground truth
        for hit in detected_hits:
            try:
                loc = hit["location"].split("-")
                hit_start = int(loc[0])
                hit_end = int(loc[1])
                
                is_tp = False
                for gt in ground_truth:
                    gt_start = gt["position"][0]
                    gt_end = gt["position"][1]
                    if max(gt_start, hit_start) < min(gt_end, hit_end):
                        is_tp = True
                        break
                
                if not is_tp:
                    # Find which category this hit would belong to
                    for cat, ocu_type in CATEGORY_MAP.items():
                        if hit["entity"] == ocu_type:
                            overall_metrics[cat]["fp"] += 1
                            break
            except:
                continue

    # Calculate final metrics
    final_report = {}
    for cat, m in overall_metrics.items():
        tp, fp, fn = m["tp"], m["fp"], m["fn"]
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

    # Save results to temp (git ignored)
    os.makedirs("temp", exist_ok=True)
    with open("temp/bigcode_results.json", "w") as f:
        json.dump(final_report, f, indent=2)
        
    print("\n--- BigCode PII Benchmark Results ---")
    print(f"{'Category':<15} | {'Precision':<10} | {'Recall':<10} | {'F1':<10}")
    print("-" * 55)
    for cat, metrics in final_report.items():
        print(f"{cat:<15} | {metrics['precision']:<10} | {metrics['recall']:<10} | {metrics['f1_score']:<10}")
    
    print(f"\nFull results saved to temp/bigcode_results.json")

if __name__ == "__main__":
    evaluate()
