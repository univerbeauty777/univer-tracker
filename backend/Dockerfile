# syntax=docker/dockerfile:1.7

# =============================================================================
# Build stage
# =============================================================================
FROM golang:1.23-alpine AS builder

WORKDIR /src

# Resolve modules first so dependency layers stay cached when only
# application code changes.
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# Compile both binaries with shared module + build caches so incremental
# rebuilds (CI, Coolify) finish in seconds instead of minutes.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

# =============================================================================
# API runtime
# =============================================================================
FROM alpine:3.20 AS api

RUN apk add --no-cache ca-certificates curl tzdata && \
    addgroup -S app && adduser -S app -G app

ENV TZ=America/Sao_Paulo

WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /src/migrations /app/migrations

USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -fsS http://localhost:8080/healthz || exit 1

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

# =============================================================================
# Migrate runtime — bakes migrations into the image so we don't need a
# host bind mount (Coolify runs compose inside a helper container, so
# relative-path mounts resolve to a path that does not exist on the
# Docker daemon's host).
# =============================================================================
FROM migrate/migrate:v4.18.1 AS migrate

COPY migrations /migrations
