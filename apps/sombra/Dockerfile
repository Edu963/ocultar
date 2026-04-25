FROM node:22-bookworm AS dashboard-builder
WORKDIR /dashboard
COPY apps/dashboard/package*.json ./
RUN npm install
COPY apps/dashboard/ ./
RUN npm run build

FROM golang:1.25-bookworm AS builder

# Install build dependencies for DuckDB/CGO
RUN apt-get update && apt-get install -y --no-install-recommends gcc g++ libc6-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# Copy local modules first to leverage cache
COPY internal/pii/ /workspace/internal/pii/
COPY services/refinery/ /workspace/services/refinery/
COPY services/vault/ /workspace/services/vault/

# Copy sombra app
COPY apps/sombra/ /workspace/apps/sombra/

WORKDIR /workspace/apps/sombra

# Build the Sombra binary (CGO enabled for DuckDB)
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -a -o /workspace/sombra ./main.go

# Patch dummy key (64 zeros) with a 64-character placeholder
RUN sed -i 's/0000000000000000000000000000000000000000000000000000000000000000/OCU_MASTER_KEY_PLACEHOLDER______________________________________/g' /workspace/sombra

# ── Final minimal image ────────────────────────────────────────────────────────
FROM debian:bookworm-slim

# Add ca-certificates and wget for health checks
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata wget && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /workspace/sombra /app/sombra
# Copy config from refinery
COPY --from=builder /workspace/services/refinery/configs/ /app/configs/

# Copy dashboard assets
COPY --from=dashboard-builder /dashboard/dist/ /app/dashboard/

# Default port
EXPOSE 8081

ENTRYPOINT ["/app/sombra"]
