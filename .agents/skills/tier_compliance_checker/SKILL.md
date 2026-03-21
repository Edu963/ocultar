---
name: tier_compliance_checker
description: >
  Cross-references the active OCULTAR license with the Sombra gateway policy.
  Ensures Pro features are only enabled when backed by a valid Enterprise license.
---

# Skill: Tier Compliance Checker

## Context
Use this skill to identify "licensing drift" where highly-sensitive connectors (Slack, SharePoint) or features are enabled in `sombra.yaml` without a corresponding Enterprise license.

## Inputs / Outputs

### Inputs
| Variable | Description |
|---|---|
| `OCU_LICENSE_KEY` | The active license token |
| `SOMBRA_CONFIG` | Path to `sombra.yaml` |

### Outputs
| Variable | Description |
|---|---|
| `COMPLIANT` | Boolean |
| `VIOLATIONS` | List of features enabled but not licensed |

## Steps

### 1. Decode Active License
```bash
TOKEN="<OCU_LICENSE_KEY>"
TIER=$(echo "$TOKEN" | cut -d. -f2 | base64 -d | jq -r .Tier)
CAPS=$(echo "$TOKEN" | cut -d. -f2 | base64 -d | jq -r .Capabilities)
```

### 2. Scan Sombra Policy for Pro Connectors
```bash
# Check for Slack connector
grep -A 2 "type: slack" apps/sombra/configs/sombra.yaml
```

### 3. Verify Compliance Logic

| Feature | Requirement | Logic |
|---|---|---|
| Slack Connector | Enterprise + CapBit 1 | If `slack` found AND (`TIER` != "enterprise" OR `CAPS` & 1 == 0) -> **VIOLATION** |
| SharePoint Connector | Enterprise + CapBit 2 | If `sharepoint` found AND (`TIER` != "enterprise" OR `CAPS` & 2 == 0) -> **VIOLATION** |

## Remediation
- **Scenario A**: Features enabled but no license -> **Downgrade Sombra Config** (remove connector).
- **Scenario B**: Valid client but missing bitmask -> **Reissue license** with correct `--capabilities` flag.
