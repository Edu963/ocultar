import os
import torch
import json
from datasets import load_dataset
from transformers import (
    AutoTokenizer,
    AutoModelForTokenClassification,
    TrainingArguments,
    Trainer,
    DataCollatorForTokenClassification,
)
import evaluate
import numpy as np

# Configuration
MODEL_ID = "openai/privacy-filter"
TRAIN_FILE = "data/fine-tune/french_finance_train.jsonl"
EVAL_FILE = "data/fine-tune/french_finance_eval.jsonl"
OUTPUT_DIR = "models/privacy-filter-fr-finance"

# Label mapping (must match data_gen_french_finance.py)
ID2LABEL = {
    0: 'O', 
    1: 'B-account_number', 2: 'I-account_number', 3: 'E-account_number', 4: 'S-account_number', 
    5: 'B-private_address', 6: 'I-private_address', 7: 'E-private_address', 8: 'S-private_address', 
    9: 'B-private_date', 10: 'I-private_date', 11: 'E-private_date', 12: 'S-private_date', 
    13: 'B-private_email', 14: 'I-private_email', 15: 'E-private_email', 16: 'S-private_email', 
    17: 'B-private_person', 18: 'I-private_person', 19: 'E-private_person', 20: 'S-private_person', 
    21: 'B-private_phone', 22: 'I-private_phone', 23: 'E-private_phone', 24: 'S-private_phone', 
    25: 'B-private_url', 26: 'I-private_url', 27: 'E-private_url', 28: 'S-private_url', 
    29: 'B-secret', 30: 'I-secret', 31: 'E-secret', 32: 'S-secret',
    33: 'B-organization', 34: 'I-organization', 35: 'E-organization', 36: 'S-organization',
    37: 'B-amount', 38: 'I-amount', 39: 'E-amount', 40: 'S-amount'
}
LABEL2ID = {v: k for k, v in ID2LABEL.items()}

def train():
    print(f"Loading tokenizer and model: {MODEL_ID}")
    tokenizer = AutoTokenizer.from_pretrained(MODEL_ID)
    
    # Load model with updated label count
    model = AutoModelForTokenClassification.from_pretrained(
        MODEL_ID,
        num_labels=len(ID2LABEL),
        id2label=ID2LABEL,
        label2id=LABEL2ID,
        ignore_mismatched_sizes=True # Crucial since we added labels
    )

    print("Loading datasets...")
    dataset = load_dataset("json", data_files={"train": TRAIN_FILE, "test": EVAL_FILE})

    def tokenize_and_align_labels(examples):
        tokenized_inputs = tokenizer(
            examples["tokens"], truncation=True, is_split_into_words=True, padding="max_length", max_length=128
        )
        labels = []
        for i, label in enumerate(examples["ner_tags"]):
            word_ids = tokenized_inputs.word_ids(batch_index=i)
            previous_word_idx = None
            label_ids = []
            for word_idx in word_ids:
                if word_idx is None:
                    label_ids.append(-100)
                elif word_idx != previous_word_idx:
                    label_ids.append(label[word_idx])
                else:
                    # In BIOES, sub-tokens should technically have specific tags,
                    # but for simplicity in fine-tuning, we can repeat or use -100.
                    # Here we use -100 to only train on the first sub-token.
                    label_ids.append(-100)
                previous_word_idx = word_idx
            labels.append(label_ids)
        tokenized_inputs["labels"] = labels
        return tokenized_inputs

    tokenized_datasets = dataset.map(tokenize_and_align_labels, batched=True)

    data_collator = DataCollatorForTokenClassification(tokenizer=tokenizer)
    seqeval = evaluate.load("seqeval")

    def compute_metrics(p):
        predictions, labels = p
        predictions = np.argmax(predictions, axis=2)

        true_predictions = [
            [ID2LABEL[p] for (p, l) in zip(prediction, label) if l != -100]
            for prediction, label in zip(predictions, labels)
        ]
        true_labels = [
            [ID2LABEL[l] for (p, l) in zip(prediction, label) if l != -100]
            for prediction, label in zip(predictions, labels)
        ]

        results = seqeval.compute(predictions=true_predictions, references=true_labels)
        return {
            "precision": results["overall_precision"],
            "recall": results["overall_recall"],
            "f1": results["overall_f1"],
            "accuracy": results["overall_accuracy"],
        }

    training_args = TrainingArguments(
        output_dir="./results",
        eval_strategy="epoch",
        learning_rate=2e-5,
        per_device_train_batch_size=8,
        per_device_eval_batch_size=8,
        num_train_epochs=5,
        weight_decay=0.01,
        save_total_limit=2,
        logging_dir='./logs',
        push_to_hub=False,
    )

    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=tokenized_datasets["train"],
        eval_dataset=tokenized_datasets["test"],
        processing_class=tokenizer,
        data_collator=data_collator,
        compute_metrics=compute_metrics,
    )

    print("Starting training...")
    trainer.train()
    
    print(f"Saving model to {OUTPUT_DIR}...")
    model.save_pretrained(OUTPUT_DIR)
    tokenizer.save_pretrained(OUTPUT_DIR)
    print("Training complete.")

if __name__ == "__main__":
    train()
