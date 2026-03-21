---
name: pilot_manager
description: Professional-grade orchestrator for the 14-day Enterprise Pilot lifecycle. Manages technical onboarding, risk assessment verification, and trial-to-purchase transition.
---

# Pilot Operations Manager (v1.1)

## Purpose

The POM drives the "Proof of Value" (PoV) for Enterprise clients. It ensures that technical verification and risk quantification reports are completed within the 14-day window to facilitate a successful transition to production.

## Inputs / Outputs

### Inputs
- `client_name`: Enterprise identifier.
- `pilot_start_date`: RFC3339 activation timestamp.
- `protected_entity_sample`: Path to `entities.json` for validation.

### Outputs
- `pilot_readiness`: `0-100%`.
- `risk_quantification`: Mock or real dollar-value of PII protected.
- `verdict`: `SUCCESS` | `STALLED` | `EXPAND_TO_PROD`.

## Preconditions
- `manage_ocultar_license` has issued a `PRO_PILOT` token.

---

## Instructions

### 1. Zero-Trust Onboarding (Day 1)
- Verify `OCU_MASTER_KEY` rotation from default.
- **Action**: Call `dictionary_shield_manager` to sync `protected_entity_sample`.
- **Gate**: If `license_tier` != `PRO_PILOT`, halt onboarding.

### 2. Value Quantification (Day 7)
- Query the ROI Accountant for blocked egress events:
  `ocultar-roi-tool query --client="{{client_name}}" --timeframe="7d"`
- **Metrics**: 
    - `Total_PII_Blocked`: Count of redact events.
    - `Avoided_Liability`: Calculated using standard GDPR/CCPA fine estimates (e.g., $100/record).

### 3. Hardening & Handover (Day 14)
- Execute `pii-regression-suite-runner` to prove accuracy.
- Generate the `Pilot_Success_Artifact.md` containing the ROI metrics and security verdicts.

## Failure Handling
- **`LOW_ENGAGEMENT`**: If `Total_PII_Blocked` == 0 after 5 days, trigger a mandatory workshop on "Connector Configuration".
- **`LICENSE_EXPIRED`**: If Pilot exceeds 14 days without an extension, signal `continuous-ai-orchestrator` to enter `RESTRICTED_MODE`.

## Postconditions
- Artifact: Final signed PoV Report must be archived in `evidence-archiver`.
