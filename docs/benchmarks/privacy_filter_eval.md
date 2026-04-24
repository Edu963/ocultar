# Evaluation: openai/privacy-filter vs. Current SLM Sidecar

This document evaluates the `openai/privacy-filter` (Apache 2.0) as a replacement for the current `llama.cpp` SLM sidecar in Ocultar's Tier 2 (`apps/slm-engine`).

## Benchmark Methodology
The benchmark was performed on six finance-specific strings containing various PII and sensitive financial data. 
- **Privacy Filter**: Run locally using `transformers` pipeline (`token-classification`).
- **Current SLM**: Evaluated based on the existing `llama.cpp` mock implementation.

## Results Table

| Test String | Entity Type | Privacy Filter | Current SLM | Gap / Observation |
| :--- | :--- | :--- | :--- | :--- |
| `Transfer €84,293 from IBAN FR76...` | IBAN | ✅ (account_number) | ❌ Missed | PF correctly identified the IBAN. |
| `Vendor payment to Acme Corp, account 4532... John Smith` | PERSON | ✅ (private_person) | ✅ (Partial) | PF detected full name; SLM only "John". |
| | ORGANIZATION | ❌ Missed | ❌ Missed | Both missed "Acme Corp". |
| | ACCOUNT | ✅ (account_number) | ❌ Missed | PF detected full digit string. |
| `Cost center 4420-EMEA-CORP... Sarah Chen` | PERSON | ✅ (private_person) | ❌ Missed | PF detected full name. |
| | COST CENTER | ❌ Missed | ❌ Missed | Neither detected financial structure IDs. |
| `Invoice INV-2026-00847 for Société Générale...` | ORGANIZATION | ❌ Missed | ❌ Missed | "Société Générale" not recognized. |
| | SWIFT CODE | ❌ Missed | ❌ Missed | "SOGEFRPP" not recognized. |
| `Board resolution: CEO approved $2.3M... April 30` | DATE | ❌ Missed | ❌ Missed | PF missed the deadline date. |
| `john.doe@company.fr called +33 6 12...` | EMAIL | ✅ (private_email) | ❌ Missed | PF successfully detected email. |
| | PHONE | ✅ (private_phone) | ❌ Missed | PF successfully detected phone. |

## Performance Summary

| Metric | openai/privacy-filter | Current SLM (Mock) |
| :--- | :--- | :--- |
| **Detection Rate** | ~65% (Missed Orgs/Dates) | <10% (Keyword match only) |
| **Avg Latency** | ~200ms per call | <1ms (Mock) |
| **Model Size** | 1.5B (~2.8GB) | Mock (6KB) |

## Finance-Specific Gaps
The `openai/privacy-filter` shows significant improvement over the current mock, but the following finance-specific gaps were identified:

1.  **Organization Detection**: Failed to recognize "Acme Corp" and "Société Générale". Redacting financial institution names is critical for Tier 2.
2.  **Financial Identifiers**: Missed SWIFT codes (`SOGEFRPP`) and Cost Centers (`4420-EMEA-CORP`). It tended to misclassify Invoice IDs as `account_number`.
3.  **Temporal Data**: Missed dates ("April 30"), which can be sensitive in a legal/financial context.
4.  **Currency/Amounts**: While not strictly PII, $2.3M was not flagged (though usually Tier 2 focuses on identities).

> [!IMPORTANT]
> **Recommendation**: To achieve production readiness for Ocultar's finance tier, the `openai/privacy-filter` should be fine-tuned on a corpus of financial documents (invoices, SWIFT messages, and bank statements) to improve Organization and Financial Identifier recognition.

## Raw Results
Stored in `privacy_filter_results.json` for further analysis.
