FROM golang:1.24-bookworm AS builder

# Install build dependencies for DuckDB/CGO (Debian has glibc — required by libduckdb.a)
RUN apt-get update && apt-get install -y --no-install-recommends gcc g++ libc6-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy the engine module (dependency via replace directive)
COPY services/engine/ /build/services/engine/

# Copy the proxy module
COPY apps/proxy/ /build/apps/proxy/

WORKDIR /build/apps/proxy

# Download Go module dependencies
RUN go mod download

# Build the proxy binary (CGO enabled for DuckDB)
RUN CGO_ENABLED=1 GOOS=linux go build -a -o /go/bin/ocultar-proxy ./main.go

# ── Final minimal image ─────────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata wget && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /go/bin/ocultar-proxy /app/ocultar-proxy

# Include engine configs so protected_entities.json is available at runtime
COPY services/engine/configs/ /app/configs/

EXPOSE 8081

ENTRYPOINT ["/app/ocultar-proxy"]
