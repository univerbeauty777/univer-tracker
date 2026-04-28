# syntax=docker/dockerfile:1.7

# =============================================================================
# Build stage
# =============================================================================
FROM golang:1.23-alpine AS builder

WORKDIR /src

# Cache dependencies layer.
COPY go.mod go.sum* ./
RUN go mod download || true

# Build binaries.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

# =============================================================================
# API runtime
# =============================================================================
FROM alpine:3.20 AS api

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

ENV TZ=America/Sao_Paulo

WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /src/migrations /app/migrations

USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/api"]

# =============================================================================
# Worker runtime
# =============================================================================
FROM alpine:3.20 AS worker

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

ENV TZ=America/Sao_Paulo

WORKDIR /app
COPY --from=builder /out/worker /app/worker

USER app

ENTRYPOINT ["/app/worker"]
