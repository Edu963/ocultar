# OCULTAR on AWS Marketplace

**Document classification:** Internal operations reference  
**Last updated:** 2026-04-28

---

## 1. Listing Description (500 characters)

> OCULTAR is a zero-egress PII refinery for AI pipelines. It runs entirely in your VPC — no data leaves your infrastructure. A 5-tier detection pipeline (regex, dictionary, entropy, phone/address heuristics, and local AI NER) tokenizes PII in-place before any upstream API call. AES-256-GCM vault, Ed25519-signed audit log, SSRF protection, and fail-closed enforcement. GDPR Article 25 compliant by architecture.

*(497 characters)*

---

## 2. Pricing Strategy

### Public listing price
**$299 / month** (SaaS contract, metered off)

This price exists to satisfy AWS Marketplace's requirement that all listings have a public price. It is not the price enterprise customers pay.

### Real enterprise price
**€24,900 / year** via AWS Private Offer

Private Offers bypass the public price and allow custom contract terms, multi-year pricing, and procurement via existing AWS EDP commitments. This is the primary revenue path.

### Unit economics
| | |
|---|---|
| Listed public price | $299/month = $3,588/year |
| Enterprise Private Offer | €24,900/year ≈ $27,000/year |
| AWS take (≈20%) | ~$5,400 |
| **Net per enterprise customer** | **~$21,600/year** |

### How to create a Private Offer
1. Log in to AWS Marketplace Management Portal → **Offers** → **Create Private Offer**
2. Select the OCULTAR listing
3. Set: Custom price (€24,900 or USD equivalent), 12-month term, auto-renewal
4. Set buyer account ID (customer's AWS account number)
5. Send offer link — customer accepts through their AWS console and it charges against their EDP

---

## 3. CloudFormation — Secrets Manager for `OCU_MASTER_KEY`

Deploy this snippet alongside the OCULTAR ECS stack to provision secrets securely. Never pass `OCU_MASTER_KEY` as a plaintext environment variable.

```yaml
AWSTemplateFormatVersion: "2010-09-09"
Description: OCULTAR secrets provisioning

Parameters:
  MasterKeyValue:
    Type: String
    NoEcho: true
    Description: "32-byte hex AES master key (openssl rand -hex 32)"
  SaltValue:
    Type: String
    NoEcho: true
    Description: "16-byte hex deployment salt (openssl rand -hex 16)"

Resources:
  OcultarMasterKey:
    Type: AWS::SecretsManager::Secret
    Properties:
      Name: /ocultar/OCU_MASTER_KEY
      Description: "OCULTAR AES-256 master key — loss is unrecoverable"
      SecretString: !Ref MasterKeyValue

  OcultarSalt:
    Type: AWS::SecretsManager::Secret
    Properties:
      Name: /ocultar/OCU_SALT
      Description: "OCULTAR per-deployment HKDF salt"
      SecretString: !Ref SaltValue

  OcultarAuditKey:
    Type: AWS::SecretsManager::Secret
    Properties:
      Name: /ocultar/OCU_AUDIT_PRIVATE_KEY
      Description: "Ed25519 signing key for immutable audit log (enterprise)"
      SecretString: "REPLACE_WITH_openssl_rand_-hex_32"

Outputs:
  MasterKeyArn:
    Value: !Ref OcultarMasterKey
    Export:
      Name: OcultarMasterKeyArn
  SaltArn:
    Value: !Ref OcultarSalt
    Export:
      Name: OcultarSaltArn
  AuditKeyArn:
    Value: !Ref OcultarAuditKey
    Export:
      Name: OcultarAuditKeyArn
```

---

## 4. ECS Task Definition (Least-Privilege IAM)

```json
{
  "family": "ocultar-refinery",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "executionRoleArn": "arn:aws:iam::ACCOUNT_ID:role/OcultarECSExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT_ID:role/OcultarECSTaskRole",
  "containerDefinitions": [
    {
      "name": "ocultar-refinery",
      "image": "ACCOUNT_ID.dkr.ecr.REGION.amazonaws.com/ocultar-refinery:latest",
      "portMappings": [
        { "containerPort": 8080, "protocol": "tcp" },
        { "containerPort": 8081, "protocol": "tcp" }
      ],
      "secrets": [
        {
          "name": "OCU_MASTER_KEY",
          "valueFrom": "arn:aws:secretsmanager:REGION:ACCOUNT_ID:secret:/ocultar/OCU_MASTER_KEY"
        },
        {
          "name": "OCU_SALT",
          "valueFrom": "arn:aws:secretsmanager:REGION:ACCOUNT_ID:secret:/ocultar/OCU_SALT"
        },
        {
          "name": "OCU_AUDIT_PRIVATE_KEY",
          "valueFrom": "arn:aws:secretsmanager:REGION:ACCOUNT_ID:secret:/ocultar/OCU_AUDIT_PRIVATE_KEY"
        }
      ],
      "environment": [
        { "name": "OCU_PROXY_TARGET", "value": "https://api.openai.com" },
        { "name": "OCU_PROXY_PORT", "value": "8081" }
      ],
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -sf http://localhost:8080/api/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 15
      },
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/ocultar-refinery",
          "awslogs-region": "REGION",
          "awslogs-stream-prefix": "ocultar"
        }
      }
    }
  ]
}
```

### IAM — OcultarECSTaskRole (least-privilege)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ReadOcultarSecrets",
      "Effect": "Allow",
      "Action": ["secretsmanager:GetSecretValue"],
      "Resource": [
        "arn:aws:secretsmanager:REGION:ACCOUNT_ID:secret:/ocultar/*"
      ]
    },
    {
      "Sid": "WriteAuditLogs",
      "Effect": "Allow",
      "Action": ["logs:CreateLogStream", "logs:PutLogEvents"],
      "Resource": "arn:aws:logs:REGION:ACCOUNT_ID:log-group:/ecs/ocultar-refinery:*"
    }
  ]
}
```

No S3, no EC2, no DynamoDB, no VPC permissions. The task role is scoped to exactly two actions: read its own secrets and write its own logs.

---

## 5. Health Check Configuration

OCULTAR exposes `GET /api/health` on port 8080. A healthy response looks like:

```json
{
  "status": "healthy",
  "vault": {"status": "online"},
  "slm":   {"status": "offline"},
  "version": "1.1.0"
}
```

`vault.status` is `"online"` when the DuckDB vault is open. `slm.status` is `"offline"` in Community (Tier 2 AI NER is Enterprise-only). The ALB health check only validates HTTP 200 — the body is informational.

| Check | Target | Healthy | Unhealthy action |
|---|---|---|---|
| ALB target group | `GET /api/health` HTTP 200 | `{"status":"ok"}` | Deregister + replace task |
| ECS health check | `curl -sf http://localhost:8080/api/health` | exit 0 | Restart container after 3 failures |
| CloudWatch alarm | HTTP 5xx rate > 1% over 5 min | — | SNS alert to ops |

---

## 6. Customer FAQ

**Q: Does OCULTAR send my data to Anthropic, AWS, or any third party?**  
No. OCULTAR runs entirely inside your VPC. The refinery, vault, and AI NER model all execute on your compute. The only outbound connection is the one you configure via `OCU_PROXY_TARGET` — your chosen AI provider — and that connection only ever receives tokenized text with PII replaced by opaque placeholders.

**Q: What happens if the OCULTAR container crashes mid-request?**  
OCULTAR is fail-closed. If the refinery process is unavailable, the proxy returns HTTP 503 before opening any connection to the upstream AI provider. No partial or unredacted request is forwarded. The vault is backed by a named Docker volume (or PostgreSQL in HA deployments) and survives container restarts without data loss.

**Q: How do I verify my audit log hasn't been tampered with?**  
Each audit log entry contains a SHA-256 hash of the previous entry and an Ed25519 signature over the full event payload. Run the built-in verifier:
```bash
go run ./services/refinery/cmd/main.go --verify-audit audit.log
```
Any gap, deletion, or modification in the chain produces a verification failure with the index of the first broken entry.

---

## 7. Submission Checklist

Before submitting to AWS Marketplace:

- [ ] Container image published to ECR Public (`public.ecr.aws/ocultar/refinery`) — build: `docker/Dockerfile.marketplace`, push: `scripts/ecr-push.sh`
- [ ] `GET /api/health` returns HTTP 200 within 15 seconds of container start
- [ ] Listing description ≤ 500 characters (see Section 1)
- [ ] EULA / pricing confirmed: $299/month public, Private Offer for enterprise
- [ ] CloudFormation template tested in a clean AWS account
- [ ] Support email configured: `support@ocultar.dev`
- [ ] Product logo uploaded (min 120×120 PNG)
- [ ] AWS Marketplace Management Portal category: **Security → Data Protection**
