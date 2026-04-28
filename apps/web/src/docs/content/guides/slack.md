# Slack Zero-Egress Connector

The Slack connector allows you to use AI apps within Slack workspace while maintaining Zero-Egress guarantees. OCULTAR intercepts messages sent to AI bots and redacts PII before they leave your Slack enclave.

## Architecture
The connector essentially creates a "Sanitized Loop":
1. User posts a message in Slack.
2. Slack Webhook hits the **OCULTAR Sombra Proxy**.
3. Sombra refines the message (removing PII).
4. Sized message is sent to the AI backend.
5. AI response is returned to Sombra.
6. Sombra rehydrates PII if necessary and posts back to Slack.

## Setup Guide

### 1. Create a Slack App
Go to the [Slack API Dashboard](https://api.slack.com/apps) and create a new app.

### 2. Configure Event Subscriptions
Set your Request URL to your OCULTAR public endpoint:
`https://ocultar.yourcorp.com/api/v1/connectors/slack/events`

### 3. Permissions
Ensure your app has the following `bot` scopes:
- `chat:write`
- `app_mentions:read`
- `channels:history`

### 4. OCULTAR Configuration
Add your Slack tokens to `config.yaml` or set them as environment variables.

```yaml
connectors:
  slack:
    enabled: true
    signing_secret: "SLACK_SIGNING_SECRET"
    bot_token: "xoxb-YOUR-TOKEN"
```

## Security Considerations
- **Local Vaulting**: All Slack User IDs are tokenized and stored in your local Vault. The AI model never sees the actual Slack handles.
- **Fail-Closed**: If the Slack API is unreachable or OCULTAR refinement fails, the message will not be posted, ensuring no un-redacted data leaks.

> [!SECURITY]
> Always use HTTPS for your OCULTAR endpoint. Sombra requires the `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384` cipher suite for enterprise Slack integrations.
