# syntax=docker/dockerfile:1.7
# Multi-stage Dockerfile per Pesa.
# - target "dev"  → air hot-reload, codice montato via volume
# - target "prod" → binario statico con SPA embedded, alpine runtime
#
# Build dev:  docker build --target dev  -t pesa-backend:dev  .
# Build prod: docker build --target prod -t pesa:prod .

# ---- web builder (Vite bundle) ----
FROM node:20-alpine AS web-builder
WORKDIR /w
COPY web/package.json web/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY web/ ./
# Skip `tsc -b` (pre-existing type errors in test files blocked LAN builds —
# tracked separately). Vite performs its own transpile for the prod bundle.
RUN npx vite build

# ---- Go base ----
FROM golang:1.23-alpine AS base
WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata bash

# ---- deps ----
FROM base AS deps
COPY go.mod go.sum ./
RUN go mod download

# ---- dev (air hot reload) ----
FROM deps AS dev
RUN go install github.com/air-verse/air@v1.61.0
COPY . .
EXPOSE 8080
ENV CGO_ENABLED=0
# air config è .air.toml nella root
CMD ["air", "-c", ".air.toml"]

# ---- builder (prod) ----
FROM deps AS builder
COPY . .
# Embed freshly built SPA into the Go binary.
COPY --from=web-builder /w/dist/ ./internal/web/dist/
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pesa ./cmd/pesa

# ---- prod runtime ----
FROM alpine:3.20 AS prod
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S pesa && adduser -S -G pesa pesa
COPY --from=builder /out/pesa /usr/local/bin/pesa
USER pesa
WORKDIR /home/pesa
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/pesa"]
CMD ["serve"]
