import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from transformers import pipeline
import uvicorn

app = FastAPI()

# Configuration
MODEL_PATH = os.getenv("PRIVACY_FILTER_MODEL_PATH", "openai/privacy-filter")
DEVICE = -1 # Default to CPU

print(f"Loading model from {MODEL_PATH}...")
try:
    classifier = pipeline(
        "token-classification",
        model=MODEL_PATH,
        aggregation_strategy="simple",
        device=DEVICE
    )
except Exception as e:
    print(f"Error loading model: {e}")
    # Fallback to base model if fine-tuned fails
    print("Falling back to openai/privacy-filter")
    classifier = pipeline(
        "token-classification",
        model="openai/privacy-filter",
        aggregation_strategy="simple",
        device=DEVICE
    )

class ScanRequest(BaseModel):
    text: str

@app.post("/scan")
async def scan(request: ScanRequest):
    if not request.text:
        return {}
    
    try:
        entities = classifier(request.text)
        result = {}
        for e in entities:
            label = e["entity_group"]
            value = e["word"]
            if label not in result:
                result[label] = []
            if value not in result[label]:
                result[label].append(value)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health():
    return {"status": "ok", "model": MODEL_PATH}

if __name__ == "__main__":
    port = int(os.getenv("PORT", 8086))
    uvicorn.run(app, host="0.0.0.0", port=port)
