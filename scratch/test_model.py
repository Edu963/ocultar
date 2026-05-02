
import os
from transformers import pipeline

model_path = "/home/edu/ocultar/models/privacy-filter-fr-finance"
print(f"Testing model at {model_path}")

try:
    classifier = pipeline(
        "token-classification",
        model=model_path,
        aggregation_strategy="none",
        trust_remote_code=True
    )
    text = "Le virement de 5000€ doit être effectué vers l'IBAN FR76 3000 6000 0112 3456 7890 123 pour la société OCULTAR SAS (SIRET 12345678901234)."
    results = classifier(text)
    print("Results:", results)
except Exception as e:
    print(f"Error: {e}")
