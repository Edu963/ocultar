import json
import requests
import sys

json_file = "/home/edu/Documents/taxes/IB_1099_2025_Data.json"
api_url = "http://localhost:9091/api/refine"

print(f"Loading JSON from {json_file}...")
with open(json_file, "r") as f:
    data = json.load(f)

print("Sending JSON to Enterprise Refinery...")
# Send the data directly as a JSON object
response = requests.post(api_url, json=data)

if response.status_code != 200:
    print(f"Error from refinery: {response.status_code} - {response.text}")
    sys.exit(1)

result = response.json()

print("\n--- REFINED OUTPUT (Snippet) ---\n")
refined = result.get("refined", {})
# If refined is a JSON string (as we saw before with ProcessInterface), parse it for display
if isinstance(refined, str):
    try:
        refined_obj = json.loads(refined)
        print(json.dumps(refined_obj, indent=4)[:1000] + "...")
    except:
        print(refined[:1000] + "...")
else:
    print(json.dumps(refined, indent=4)[:1000] + "...")

print("\n--- DETECTED ENTITIES ---")
report = result.get("report", {})
hits = report.get("hits", [])
if not hits:
    # Try pii_hits which we saw in the raw JSON output earlier
    hits = report.get("pii_hits", [])

if not hits:
    print("No entities detected.")
else:
    counts = {}
    for hit in hits:
        etype = hit.get("entity", "UNKNOWN")
        counts[etype] = counts.get(etype, 0) + 1
    
    for etype, count in sorted(counts.items(), key=lambda x: x[1], reverse=True):
        print(f"  {etype}: {count} instances")

# Save results
with open("/tmp/ib_1099_results.json", "w") as f:
    json.dump(result, f, indent=2)

print(f"\nTotal entities detected: {len(hits)}")
print("Detailed results saved to /tmp/ib_1099_results.json")
