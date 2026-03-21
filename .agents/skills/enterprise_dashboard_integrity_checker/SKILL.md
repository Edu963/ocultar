---
name: enterprise-dashboard-integrity-checker
description: Validates UI consistency for Risk Matrix, ROI, and Audit Logs. Confirms navigation links and permission tiers are correct.
---

# Enterprise Dashboard Integrity Checker

## Purpose

This skill prevents "Visual Compliance" regressions. It ensures the dashboard accurately reflects the state of the backend refinery and the user's license tier.

## When To Use This Skill

Use this skill when:
- Modifying the Enterprise Dashboard (React/Frontend).
- Changing the license tier logic (Pro vs. Enterprise).
- Updating the Audit Log or ROI metrics schema.
- Adding new panels or charts to the Risk Matrix.

## Instructions

1.  **Check License Alignment**: Ensure that "Elite" or "Enterprise" features (like the Identity Vault) are appropriately locked/unlocked based on the license state.
2.  **Validate Audit Mapping**: Confirm that the frontend correctly parses the regulatory tags (GDPR, HIPAA) provided by the backend.
3.  **Verify Navigation Integrity**: Check that all internal links, breadcrumbs, and sidebar items are valid and lead to the correct components.
4.  **Match ROI Metrics**: Ensure that the data visualizations (charts, gauges) use the correct units and time-series data from the `ROI Engine`.
5.  **Prevent UI Regressions**: Look for common issues like "double menus", misaligned tables, or broken search filters.

## Examples

### Adding "Financial Data" to Risk Matrix
**Action**: Verify that the frontend can display the "PCI-DSS" category and that the corresponding audit log entries are correctly filtered when clicking the "Financial" tile.
