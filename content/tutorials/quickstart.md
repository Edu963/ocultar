# Quickstart: From Zero to Zero-Egress

This guide will help you set up OCULTAR in a local development environment.

## Prerequisites
- Docker & Docker Compose
- 4GB RAM minimum
- Internet access (for the initial binary pull)

## 1. Launch the Stack
Run the following command to start the OCULTAR Refinery and a mock API for testing.

```bash
docker-compose -f docker-compose.community.yml up -d
```

## 2. Verify Deployment
Ensure the refinery is active by checking the health endpoint:

```bash
curl http://localhost:3030/health
```

Expected response:
```json
{ "status": "active", "tier": "community", "vault": "healthy" }
```

## 3. Your First Refinement
Send a sample prompt containing PII to the refinery proxy:

```bash
curl -X POST http://localhost:3030/refine \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "My name is John Doe and my SSN is 000-11-2222."}
    ]
  }'
```

The refinery will intercept and return:
```json
{
  "content": "My name is [PERSON_1] and my SSN is [SSN_1]."
}
```

## Next Steps
- [Configure custom PII rules](/docs/guides/pii-rules)
- [Connect to Slack](/docs/guides/slack)
- [Set up the Enterprise Vault](/docs/explanation/vault)
