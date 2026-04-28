# Sombra Proxy Configuration

Sombra is the high-performance transparent proxy component of OCULTAR. It allows you to protect traffic between your applications and AI providers without modifying your application code (beyond changing the base URL).

## Transparent Proxy Mode
In this mode, Sombra acts as a drop-in replacement for the OpenAI/Anthropic SDK endpoints.

### Setup
Instead of pointing your SDK to `api.openai.com`, point it to your Sombra instance:

```javascript
const openai = new OpenAI({
  apiKey: 'YOUR_API_KEY',
  baseURL: 'http://ocultar.internal:3030/api/v1/proxy/openai/v1'
});
```

## Supported Providers
Sombra supports multiple providers via their standard API formats:
- **OpenAI**: `/v1/proxy/openai/v1`
- **Anthropic**: `/v1/proxy/anthropic/v1`
- **Azure OpenAI**: `/v1/proxy/azure-openai`

---

## Performance Tuning

### Concurrency Limits
Control the load on your Refinery engine using the `max_concurrency` setting.

```yaml
settings:
  max_concurrency: 50
  queue_size: 20
```

### Timeouts
AI deep scans (Tier 2) can be slow. Set a strict timeout to ensure low latency.

```yaml
settings:
  inference_timeout: "5s"
```

## Security: Rehydration Fallback
If the Vault becomes unavailable or rehydration fails, you can control whether Sombra should return the tokenized (sanitized) data or fail with an error.

```yaml
settings:
  rehydrate_fallback_enabled: false # Returns 500 error if rehydration fails
```

> [!WARNING]
> Enabling `rehydrate_fallback_enabled: true` might break the LLM's logic if it receives tokens like `[PERSON_1]` instead of the rehydrated name, but it prevents total service downtime.
