# Zero-Egress Architecture

Zero-Egress is the foundational security guarantee of OCULTAR. Unlike traditional security perimeters that focus on "Who is entering the network," Zero-Egress focuses on **"What is leaving the refinery."**

## The Problem: Data Exfiltration via AI
As organizations adopt LLMs, they inadvertently create massive egress points. Every prompt sent to an external API (OpenAI, Anthropic, Google) is a potential leak of PII, internal IP, or customer data.

## The Solution: The OCULTAR Shield
OCULTAR acts as a "unidirectional valve" for your data:

1.  **Intercept**: All outbound traffic to AI models is routed through the **Sombra Proxy**.
2.  **Redact**: Within the refinery, PII is tokenized (e.g., `Héctor Eduardo` -> `TOKEN_8291`).
3.  **Validate**: Data is checked against the **Regulatory Matrix (GDPR, HIPAA, etc.)**.
4.  **Forward**: Only sanitized data reaches the external LLM.

## Key Principles

### 1. Fail-Closed by Design
If the Refinery cannot determine if a piece of data is safe, it **blocks the request**. Security always takes precedence over availability in the default configuration.

### 2. Immutable Cryptographic Proof
Every redaction event is logged to the **Immutable Vault**. This provides a defensible audit trail that can be used to prove compliance to regulators.

### 3. Local-First Processing
All PII detection and tokenization happens **locally** on your infrastructure. Sensitive data never touches the OCULTAR cloud or the AI provider's cloud in its raw form.

> [!SECURITY]
> OCULTAR's Zero-Egress guarantee is maintained even if the upstream AI provider is compromised. Since they only ever see tokens, the risk of a data breach at the LLM provider is functionally mitigated.
