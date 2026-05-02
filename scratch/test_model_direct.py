
import torch
from transformers import AutoTokenizer, AutoModelForTokenClassification

model_path = "/home/edu/ocultar/models/privacy-filter-fr-finance"
tokenizer = AutoTokenizer.from_pretrained(model_path)
model = AutoModelForTokenClassification.from_pretrained(model_path, trust_remote_code=True)

text = "Le virement de 5000€ doit être effectué vers l'IBAN FR76 3000 6000 0112 3456 7890 123 pour la société OCULTAR SAS (SIRET 12345678901234)."
inputs = tokenizer(text, return_tensors="pt")

with torch.no_grad():
    outputs = model(**inputs)
    logits = outputs.logits
    predictions = torch.argmax(logits, dim=-1)

print("Tokens:", tokenizer.convert_ids_to_tokens(inputs["input_ids"][0]))
print("Predictions:", [model.config.id2label[str(p.item())] for p in predictions[0]])
