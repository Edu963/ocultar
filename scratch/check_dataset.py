from datasets import load_dataset
import sys

try:
    # Try loading a small streaming subset to check access
    ds = load_dataset("bigcode/bigcode-pii-dataset", split="train", streaming=True)
    item = next(iter(ds))
    print("Success! Dataset is accessible.")
    print(item.keys())
except Exception as e:
    print(f"Error: {e}")
    sys.exit(1)
