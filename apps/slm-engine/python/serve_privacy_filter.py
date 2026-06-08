import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from transformers import pipeline
import uvicorn

app = FastAPI()

MODEL_PATH = os.getenv("PRIVACY_FILTER_MODEL_PATH", "openai/privacy-filter")
DEVICE = -1

# MODEL_SCHEMA controls label normalization.
# "privacy-filter" (default): openai/privacy-filter and its fine-tunes (fr-finance, etc.)
# "piiranha": iiiorg/piiranha-v1-detect-personal-information
MODEL_SCHEMA = os.getenv("MODEL_SCHEMA", "privacy-filter")

# Canonical OCULTAR label set. All model outputs are normalized to these.
_PRIVACY_FILTER_LABELS = {
    "private_person": "PERSON",
    "private_email": "EMAIL",
    "private_phone": "PHONE",
    "private_address": "ADDRESS",
    "private_date": "DATE",
    "private_url": "URL",
    "secret": "SECRET",
    "organization": "ORGANIZATION",
    "account_number": "ACCOUNT_NUMBER",
    "amount": "FINANCIAL_AMOUNT",
}

_PIIRANHA_LABELS = {
    "PERSON": "PERSON",
    "EMAIL_ADDRESS": "EMAIL",
    "PHONE_NUM": "PHONE",
    "STREET_ADDRESS": "ADDRESS",
    "LOCATION": "LOCATION",
    "ORGANIZATION": "ORGANIZATION",
    "DATE_TIME": "DATE",
    "AGE": "AGE",
    "NRP": "SENSITIVE_ATTRIBUTE",
    "URL_PERSONAL": "URL",
    "IP_ADDRESS": "IP_ADDRESS",
    "CREDIT_CARD": "CREDIT_CARD",
    "US_SSN": "SSN",
    "US_BANK_NUMBER": "ACCOUNT_NUMBER",
    "IBAN_CODE": "IBAN",
    "US_PASSPORT": "PASSPORT",
    "US_DRIVER_LICENSE": "LICENSE",
    "MEDICAL_LICENSE": "MEDICAL_LICENSE",
    "USERNAME": "USERNAME",
    "PASSWORD": "CREDENTIAL",
    "TITLE": "TITLE",
}

_LABEL_MAP = _PIIRANHA_LABELS if MODEL_SCHEMA == "piiranha" else _PRIVACY_FILTER_LABELS


def _normalize(raw_label: str) -> str:
    return _LABEL_MAP.get(raw_label, raw_label.upper())


print(f"Loading model from {MODEL_PATH} (schema: {MODEL_SCHEMA})...")
try:
    classifier = pipeline(
        "token-classification",
        model=MODEL_PATH,
        aggregation_strategy="simple",
        device=DEVICE,
        trust_remote_code=True,
    )
except Exception as e:
    print(f"Error loading model: {e}")
    print("Falling back to openai/privacy-filter")
    classifier = pipeline(
        "token-classification",
        model="openai/privacy-filter",
        aggregation_strategy="simple",
        device=DEVICE,
        trust_remote_code=True,
    )
    MODEL_SCHEMA = "privacy-filter"
    _LABEL_MAP = _PRIVACY_FILTER_LABELS


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
            label = _normalize(e["entity_group"])
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
    return {"status": "ok", "model": MODEL_PATH, "schema": MODEL_SCHEMA}


if __name__ == "__main__":
    port = int(os.getenv("PORT", 8086))
    uvicorn.run(app, host="0.0.0.0", port=port)
