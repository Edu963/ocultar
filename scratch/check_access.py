import os
from huggingface_hub import HfApi
from dotenv import load_dotenv

load_dotenv()
token = os.getenv("HF_TOKEN")
api = HfApi(token=token)

try:
    info = api.dataset_info("bigcode/bigcode-pii-dataset")
    print(f"Dataset info: {info.id}")
    print(f"Is gated: {info.gated}")
    if info.gated:
        print("Dataset is gated. You need to accept the terms on the HF website.")
except Exception as e:
    print(f"Error: {e}")
