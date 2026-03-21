---
name: industry-snapshot-generator
description: Instantly configures the Ocultar Refinery, Dictionary Shield, and Enterprise Dashboard for a specific industry (Finance, Healthcare, GovTech, etc.).
---

# Industry-Specific Snapshot Generator

## Purpose

This skill bridges the gap between a generic security tool and a deep, vertical-specific compliance solution. It allows Sales and Customer Success teams to demonstrate immediate relevance by pre-configuring the system for the prospect's specific regulatory environment.

## When To Use This Skill

Use this skill:
- Preparing for a sales demo to a prospect in a specific vertical (e.g., a bank).
- During initial client onboarding to provide a "warm start" configuration.
- Designing industry-specific "Gold Images" for enterprise distribution.
- Testing the Refinery against industry-standard data formats (e.g., HL7 for healthcare).

## Instructions

1.  **Define Target Vertical**: Identify the primary regulatory framework (GDPR, HIPAA, PCI-DSS) and industry (Finance, Healthcare, etc.).
2.  **Select Rule Set**: Load pre-defined configuration templates:
    - **Finance**: IBAN, SWIFT, Credit Card (PCI), Proprietary Ticker symbols.
    - **Healthcare**: Patient IDs, ICD-10 codes, HIPAA-protected identifiers.
    - **GovTech**: Social Security Numbers, Classified Project codenames, Tax IDs.
3.  **Configure Dictionary Shield**: Infuse the Tier 0 shield with industry-specific sensitive terminology (e.g., "M&A", "Internal Trial Data", "Classified").
4.  **Mock Dashboard Data**: Populate the Risk Matrix and ROI Analytics with realistic "Day 1" data points to visualize the value proposition.
5.  **Apply Sales Persona**: Ensure the Dashboard metrics (e.g., "Potential Fines Avoided") are calibrated to the vertical's average regulatory penalties.
6.  **Safety Guardrail**: Ensure these snapshots are applied to *staging/demo* environments; never overwrite production settings without explicit confirmation.

## Examples

### Finance Snapshot
**Action**: Load European IBAN regex, SWIFT/BIC patterns, and populate the Risk Matrix with mocked "Financial Leak" events from a simulated chatbot.

### Healthcare Onboarding
**Action**: Activate the HIPAA detection profile, load ICD-10 SLM prompts, and sets the Dashboard ROI multiplier based on healthcare data breach costs ($10M+).
