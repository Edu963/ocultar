FROM golang:1.24-bookworm AS builder

# Install build dependencies for DuckDB/CGO
RUN apt-get update && apt-get install -y --no-install-recommends gcc g++ libc6-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# Copy local modules first to leverage cache
COPY internal/pii/ /workspace/internal/pii/
COPY services/refinery/ /workspace/services/refinery/
COPY services/vault/ /workspace/services/vault/

# Copy proxy app
COPY apps/proxy/ /workspace/apps/proxy/

WORKDIR /workspace/apps/proxy

# Build the proxy binary (CGO enabled for DuckDB)
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -a -o /workspace/ocultar-proxy ./main.go

# ── Final minimal image ─────────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata wget && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /workspace/ocultar-proxy /app/ocultar-proxy
# Copy config from refinery
COPY --from=builder /workspace/services/refinery/configs/ /app/configs/

EXPOSE 8081

ENTRYPOINT ["/app/ocultar-proxy"]
