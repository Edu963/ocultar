# Sombra | OCULTAR Agentic Gateway

Sombra is the agentic gateway for the OCULTAR ecosystem. It is a drop-in
replacement for the OpenAI API: point any OpenAI-compatible SDK at Sombra and every
prompt is scrubbed before it leaves the building — with no changes to client code.

```
OPENAI_BASE_URL=http://sombra.internal:8086/v1
```

## Key Features

- **OpenAI-compatible `/v1/chat/completions`** — supports buffered and real token streaming (`"stream": true`)
- **Multi-model routing** — OpenAI, Anthropic Claude, Google Gemini, Mistral, and any local Ollama/llama.cpp endpoint
- **PII scrub on ingress** — every message content run through the Ocultar Refinery before reaching any upstream model
- **Vault-token-boundary streaming** — `StreamRehydrator` holds partial `[TYPE_xxxxxxxx]` tokens across SSE chunk boundaries; no partial token ever forwarded to the client
- **Zero-egress domain allow-list** — adapters not in the approved domain list are blocked at the router (fail-closed)
- **Immutable audit log** — Ed25519-signed `sombra_audit.log`, one entry per completion
- **Tier 2 AI NER** — optional SLM sidecar activated by setting `SLM_SIDECAR_URL`

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/chat/completions` | OpenAI-compatible chat completions (buffered + streaming) |
| `POST` | `/query` | Legacy single-prompt endpoint |
| `POST` | `/v1/slack/events` | Slack Events API connector |
| `GET`  | `/healthz` | Health check — returns `{"status":"ok"}` |

## Package Layout

```
apps/sombra/
├── main.go              # Bootstrap: vault, router, gateway, HTTP server
└── pkg/
    ├── handler/
    │   ├── gateway.go           # Handler wiring + actor extraction (JWT / dev passthrough)
    │   ├── openai_proxy.go      # HandleV1ChatCompletions + handleStreamingResponse
    │   ├── query_handler.go     # HandleQuery (legacy endpoint)
    │   ├── slack_handler.go     # HandleSlackEvent
    │   └── stream_rehydrator.go # Vault-token-boundary-aware SSE rehydration
    ├── router/
    │   ├── router.go    # Router + domain allow-list enforcement
    │   ├── stream.go    # Streamer interface + Router.SendStream (with buffered fallback)
    │   ├── openai.go    # OpenAI adapter (Send + SendStream)
    │   ├── claude.go    # Anthropic Claude adapter (Send + SendStream)
    │   ├── gemini.go    # Google Gemini adapter (Send + SendStream)
    │   └── local.go     # Local/Ollama adapter (Send + SendStream)
    ├── connector/       # File and Slack connectors + DataPolicy
    └── scrubber/        # Pre-scrub helpers
```

## Getting Started

### Prerequisites

- Go 1.22+
- API keys for the models you want to use

### Run

```bash
cp ../../.env.example .env   # fill in OCU_MASTER_KEY + provider API keys
go run ./apps/sombra
```

Default port: **8086**. Override with `SOMBRA_PORT`.

### Test a streaming completion

```bash
curl -X POST http://localhost:8086/v1/chat/completions \
  -H "Authorization: Bearer any-value-in-dev-mode" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello, my name is Jean Dupont."}],
    "stream": true
  }'
```

### Run tests

```bash
go test ./apps/sombra/...
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `OCU_MASTER_KEY` | Yes (prod) | Master key for vault AES-256-GCM encryption |
| `OCU_SALT` | No | HKDF salt (default: `ocultar-v112-kdf-salt-fixed-16`) |
| `OCU_JWT_SECRET` | No | JWT secret for actor auth; if unset, any Bearer value is accepted (dev mode) |
| `OCU_VAULT_PATH` | No | Path to SQLite vault file (default: `sombra_vault.db`) |
| `SOMBRA_PORT` | No | HTTP listen port (default: `8086`) |
| `OPENAI_API_KEY` | For gpt-4o | OpenAI API key |
| `ANTHROPIC_API_KEY` | For Claude | Anthropic API key |
| `GEMINI_API_KEY` | For Gemini | Google Gemini API key |
| `MISTRAL_API_KEY` | For Mistral | Mistral API key |
| `SLM_SIDECAR_URL` | For Tier 2 | URL of local SLM NER sidecar (default: `http://localhost:8085`) |
| `SOMBRA_MOCK_AI_URL` | No | Mock OpenAI-compatible server for offline testing |

## Registered Models

| Client model name | Provider | Upstream identifier |
|-------------------|----------|---------------------|
| `gemini-flash-latest` | Google Gemini | `gemini-2.0-flash` |
| `gpt-4o` | OpenAI | `gpt-4o` |
| `gpt-4o-mini` | OpenAI | `gpt-4o-mini` |
| `mistral-large-latest` | Mistral | `mistral-large-latest` |
| `claude-sonnet-4-6` | Anthropic | `claude-sonnet-4-6` |
| `local-slm` | Ollama / llama.cpp | configured via `SLM_SIDECAR_URL` |

## Streaming Architecture

All four cloud adapters implement the `Streamer` interface. The `StreamRehydrator` handles
vault tokens that span SSE chunk boundaries so clients always receive complete, rehydrated text.

```
Client → POST /v1/chat/completions (stream: true)
             │
             ▼  PII scrub (Refinery.RefineString per message)
             │
             ▼  Router.SendStream → adapter.SendStream (native SSE per provider)
             │
             ▼  StreamRehydrator.Push(delta)
             │    holds incomplete [TYPE_xxxxxxxx] tail until token is complete
             │
             ▼  emit as OpenAI SSE chunk → client
             │
             ▼  StreamRehydrator.Flush() — drain any held tail
             ▼  stop chunk + data: [DONE]
```

Adapters that do not implement `Streamer` fall back transparently: `Router.SendStream` calls
`Send()` and emits the full response as a single delta — no handler changes required.
