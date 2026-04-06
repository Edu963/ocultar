FROM golang:1.25-bookworm AS builder

# Install build dependencies for CGO/llama.cpp
RUN apt-get update && apt-get install -y --no-install-recommends gcc g++ libc6-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# Copy slm-engine source and internal dependencies
COPY apps/slm-engine/ /workspace/apps/slm-engine/

WORKDIR /workspace/apps/slm-engine

# Build the SLM engine (CGO enabled for llama.cpp)
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -a -o /workspace/slm-engine ./main.go

# ── Final minimal image ─────────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata wget && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /workspace/slm-engine /app/slm-engine

# Default model path
ENV SLM_MODEL_PATH=/app/models/qwen-1.5b-q4_k_m.gguf
ENV PORT=8085

EXPOSE 8085

ENTRYPOINT ["/app/slm-engine"]
