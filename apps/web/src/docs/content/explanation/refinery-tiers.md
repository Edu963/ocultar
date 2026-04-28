# Refinery Tiers

The OCULTAR Refinery operates on a three-tier lifecycle to transform raw, potentially hazardous data into safe, technical assets.

## Tier 0: Direct Identification (Deterministic)
The first layer of defense uses high-performance Bloom filters and local dictionary matching.

- **Objective**: Immediate redaction of known entities.
- **Mechanism**: Match against a local encrypted database of system-wide PII (Names, Employee IDs, IPs).
- **Latency**: < 0.5ms.

## Tier 1: Pattern Recognition (Heuristic)
The second layer applies regional "Sovereign Packs" to identify structured PII.

- **Objective**: Identifying SSNs, Credit Cards, Passport numbers.
- **Mechanism**: Regex + Luhn validation + Contextual window analysis.
- **Latency**: < 2ms.

## Tier 2: Deep Context (AI-Powered)
The final layer uses a Small Language Model (SLM) running on-premise to identify conversational PII.

- **Objective**: Redacting names and entities that are only identifiable via context.
- **Mechanism**: Attention-based sequence labeling.
- **Latency**: 50ms - 200ms (Configurable via `slm_confidence`).

---

## The Transformation Matrix

| Tier | Sensitivity | Method | Action |
| :--- | :--- | :--- | :--- |
| **0** | Critical | Dictionary | HASH + VAULT |
| **1** | High | Heuristic | TOKENIZE |
| **2** | Contextual | SLM | REDACT |

> [!NOTE]
> All transformations are deterministic within a single session. This ensures that if "Alice" is tokenized as `TOKEN_1` in the first paragraph, she remains `TOKEN_1` throughout the entire prompt, maintaining LLM coherence.
