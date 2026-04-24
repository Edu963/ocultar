import time
import torch
import json
from transformers import pipeline

test_strings = [
    'Transfer €84,293 from IBAN FR76 3000 6000 0112 3456 7890 189',
    'Vendor payment to Acme Corp, account 4532015112830366, approved by John Smith',
    'Cost center 4420-EMEA-CORP, GL account 6100, controller Sarah Chen',
    'Invoice INV-2026-00847 for Société Générale, SWIFT SOGEFRPP',
    'Board resolution: CEO approved $2.3M acquisition, confidential until April 30',
    'john.doe@company.fr called +33 6 12 34 56 78 re: Q1 close'
]

# Rough expectation for evaluation
expected_entities = [
    ['FR76 3000 6000 0112 3456 7890 189'], # IBAN
    ['Acme Corp', '4532015112830366', 'John Smith'],
    ['4420-EMEA-CORP', 'Sarah Chen'],
    ['Société Générale', 'SOGEFRPP'],
    ['April 30'],
    ['john.doe@company.fr', '+33 6 12 34 56 78']
]

def run_privacy_filter():
    print("Loading openai/privacy-filter...")
    start_load = time.time()
    classifier = pipeline('token-classification', 
                         model='openai/privacy-filter', 
                         aggregation_strategy='simple')
    print(f"Model loaded in {time.time() - start_load:.2f}s")

    results = []
    for text in test_strings:
        start_time = time.time()
        entities = classifier(text)
        latency = (time.time() - start_time) * 1000
        
        detected = []
        for e in entities:
            detected.append({
                'label': e['entity_group'],
                'value': e['word'],
                'score': float(e['score'])
            })
        
        results.append({
            'text': text,
            'entities': detected,
            'latency': latency
        })
    return results

def get_current_slm_mock_results():
    # Based on llama.c logic
    results = []
    for text in test_strings:
        start_time = time.time()
        # Mock logic from llama.c
        detected = []
        if "John" in text:
            detected.append({'label': 'PERSON', 'value': 'John', 'score': 1.0})
        # Note: llama.c uses strstr on the prompt, but it's very limited.
        
        latency = (time.time() - start_time) * 1000 # Near zero but let's be realistic
        results.append({
            'text': text,
            'entities': detected,
            'latency': latency
        })
    return results

if __name__ == "__main__":
    pf_results = run_privacy_filter()
    slm_results = get_current_slm_mock_results()
    
    with open('privacy_filter_results.json', 'w') as f:
        json.dump({
            'privacy_filter': pf_results,
            'current_slm': slm_results
        }, f, indent=2)
    
    print("Benchmarking complete. Results saved to privacy_filter_results.json")
