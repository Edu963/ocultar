# EU Sovereign Detection Pack (v1)

## Sovereignty-First Refinery

The **EU Sovereign Detection Pack (v1)** transforms Ocultar into a production-grade detection refinery tailored for the European regulatory landscape. It combines high-performance regex patterns with algorithmic checksum validation to provide near-zero false positive rates for critical national identifiers.

### Key Capabilities

*   **Deterministic Coverage**: Full support for Tier 0 identifiers across major EU economies and the UK.
*   **Validation-First Logic**: Every detection for ES DNI, FR NIR, IT CF, NL BSN, PL PESEL, and DE Steuer-ID is verified against national checksum algorithms (Mod 11, 23, weights, etc.).
*   **Audit-Ready Metadata**: Detailed detection reports include the exact validation method used, confidence scores, and precise document offsets.
*   **Fail-Closed Security**: Integrated with the Sombra Gateway to block un-scanned or invalidly redacted data from reaching external LLMs.

### Regional Identifiers Included

| Country | Entity | Validation Method | Regulation |
|---|---|---|---|
| **Spain** | ES DNI / NIE | Mod 23 Check Character | LOPD, GDPR Art 9 |
| **France** | FR NIR | 15-digit Format | GDPR, CNIL |
| **Italy** | IT Codice Fiscale | Check Character Logic | GDPR |
| **Germany** | DE Steuer-ID | Mod 11 | BDSG, GDPR |
| **Netherlands**| NL BSN | 11-test | UAVG, GDPR |
| **Poland** | PL PESEL | Weighted Checksum | GDPR |
| **UK** | UK NINO / NHS | Format Boundaries | UK GDPR |

### Market Advantage

Unlike generic cloud-based NER (Named Entity Recognition), the Ocultar EU Sovereign Pack is:
1.  **On-Premise Only**: No data leaves your VPC for detection.
2.  **Algorithmic**: Not dependent on probabilistic "best guesses".
3.  **Audit-Ready**: Provides the "Proof of Redaction" required by EU regulators (GDPR Art 30).

---

> [!TIP]
> This pack is available for **Enterprise** customers. Contact your account manager to enable "Sovereign Mode" in your Refinery configuration.
