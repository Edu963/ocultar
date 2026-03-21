---
name: repository-knowledge-map
description: Use this skill whenever working in the Ocultar repository to understand its structure, documentation layout, and conceptual components before making changes.
---

# Repository Knowledge Map

## Purpose

This skill provides the AI agent with a structured understanding of the Ocultar repository.

It explains:

- the documentation layout
- the operational documentation
- internal reports
- pilot program materials
- core technical documentation

The goal is to ensure the AI agent modifies the correct files and respects the intended project structure.

## Repository Documentation Structure

### Core Technical Documentation

Located in:

documentation/

This directory contains the primary technical documentation of the system.

Key files include:

README.md  
ARCHITECTURE.md  
API_REFERENCE.md  
DEVELOPER_GUIDE.md  
SETUP_GUIDE.md  
ENTERPRISE_SETUP_GUIDE.md  
RELEASE_GUIDE.md  
TESTING_GUIDE.md  
TECH_DOCS.md  
SOMBRA_GUIDE.md  
CONNECTORS_GUIDE.md

These documents explain how the system works and how developers interact with it.

### Internal Technical Documentation

Located in:

documentation/internal/

Example:

LICENSE_MANAGEMENT.md

This directory contains internal operational documentation that is not meant for external distribution.

### Operational and Deployment Documentation

Located in:

docs/

Files include:

setup_guide.md
smoke_test.sh (located in tools/scripts/scripts/)

These documents describe operational procedures and deployment workflows.

### Pilot Program Documentation

Located in:

docs/pilot/

Files include:

onboarding.md  
playbook.md  
report_template.md  

These documents support pilot deployments and evaluation processes.

### Internal Reports and Compliance Analysis

Located in:

docs/internal/

Examples include:

COMPLIANCE_COMPARISON_REPORT.md  
GDPR_Exposure_Report.md  
SLM_Optimization_Report.md  

These documents contain internal research, regulatory analysis, or optimization reports.

They are not intended for client distribution.

## Conceptual Components

The repository includes several conceptual areas:

Ocultar  
The main system or platform being developed.

Sombra  
A component or subsystem documented in SOMBRA_GUIDE.md.

Pilot Program  
Operational framework used to onboard and evaluate pilot users.

Enterprise Deployment  
Enterprise deployment procedures documented in ENTERPRISE_SETUP_GUIDE.md.

Sombra Pro Connectors
Modular architecture for platform-specific ingestion (Slack, SharePoint, etc.) located in services/engine/pkg/connector/.

Internal Compliance and Research  
Internal analysis documents stored in docs/internal/.

## Instructions

Before making changes to the repository:

1. Identify which conceptual area the task belongs to.

Possible categories:

- system architecture
- APIs
- developer workflows
- setup and installation
- enterprise deployment
- pilot program operations
- compliance or research documentation

2. Locate the correct documentation directory.

3. Ensure modifications are applied only to relevant documentation.

4. Avoid modifying internal documentation unless the task explicitly concerns internal processes.

5. Ensure that updates remain consistent with the repository structure.

## Examples

Example 1

Input:
A new API endpoint is implemented.

Action:
Update documentation/API_REFERENCE.md.

Example 2

Input:
A new pilot onboarding workflow is introduced.

Action:
Update docs/pilot/onboarding.md and docs/pilot/playbook.md.

Example 3

Input:
A licensing workflow changes.

Action:
Update documentation/internal/LICENSE_MANAGEMENT.md.
