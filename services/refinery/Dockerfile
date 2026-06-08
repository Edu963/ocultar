FROM golang:1.25-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends gcc g++ libc6-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

COPY internal/pii/ /workspace/internal/pii/
COPY services/refinery/ /workspace/services/refinery/
COPY services/vault/ /workspace/services/vault/

WORKDIR /workspace/services/refinery

RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -o /workspace/ocultar-refinery ./cmd/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /workspace/ocultar-refinery /app/ocultar-refinery
COPY --from=builder /workspace/services/refinery/configs/ /app/configs/

EXPOSE 8080

ENTRYPOINT ["/app/ocultar-refinery", "--serve", "8080"]
