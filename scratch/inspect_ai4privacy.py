from datasets import load_dataset
import os
from dotenv import load_dotenv

load_dotenv()
token = os.getenv("HF_TOKEN")

try:
    ds = load_dataset("ai4privacy/pii-masking-300k", split="train", streaming=True)
    item = next(iter(ds))
    print(item.keys())
    # print(item)
except Exception as e:
    print(f"Error: {e}")
