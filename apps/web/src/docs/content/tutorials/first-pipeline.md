# Your First Refinery Pipeline

This tutorial will guide you through creating a deterministic PII redaction pipeline using OCULTAR's Tier 0 architecture.

## Overview
A "Pipeline" in OCULTAR is a set of rules applied in sequence to an incoming data stream. For your first pipeline, we will focus on matching known internal identifiers (like employee names) and redacting them before they egress to an AI model.

## 1. Define Your Protected Entities
Create a file named `data/protected_entities.json` if it doesn't exist. This acts as your deterministic dictionary.

```json
[
  "Héctor Eduardo Trejos",
  "Project Phoenix",
  "Ouroboros Protocol"
]
```

## 2. Configure the Refinery
Update your `config.yaml` to enable the dictionary matching.

```yaml
settings:
  dictionaries:
    - type: "PROTECTED_ENTITY"
      terms: ["Héctor Eduardo Trejos", "Project Phoenix"]
  slm_confidence: 0.6
```

## 3. Test the Redaction
Send a prompt to the `/refine` endpoint.

```bash
curl -X POST http://localhost:3030/refine \
  -d '{
    "messages": [
      {"role": "user", "content": "Update the Project Phoenix roadmap for Héctor Eduardo."}
    ]
  }'
```

### Expected Output
The refinery will identify both the project name and the person's name using Tier 0 logic.

```json
{
  "content": "Update the [PROTECTED_ENTITY_1] roadmap for [PROTECTED_ENTITY_2]."
}
```

## How it Works
The Tier 0 engine uses a **Bloom Filter** for high-speed lookup. When a match is found:
1. The raw value is stored in the **Secure Vault**.
2. A random token is generated.
3. The token replaces the raw value in the egress payload.

---

## Next Steps
- [Configure Regex Rules](/docs/guides/pii-rules)
- [Set up the Sombra Proxy](/docs/guides/sombra)
