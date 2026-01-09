# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build both binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /jinwoo-login ./cmd/login
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /jinwoo-channel ./cmd/channel

# Runtime stage - Login Server
FROM alpine:3.19 AS login

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 jinwoo && \
    adduser -u 1000 -G jinwoo -s /bin/sh -D jinwoo

WORKDIR /app

COPY --from=builder /jinwoo-login /app/jinwoo-login

# Data directories
RUN mkdir -p /app/data && chown -R jinwoo:jinwoo /app

USER jinwoo

EXPOSE 8484

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD nc -z localhost 8484 || exit 1

ENTRYPOINT ["/app/jinwoo-login"]

# Runtime stage - Channel Server
FROM alpine:3.19 AS channel

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 jinwoo && \
    adduser -u 1000 -G jinwoo -s /bin/sh -D jinwoo

WORKDIR /app

COPY --from=builder /jinwoo-channel /app/jinwoo-channel

# Data and scripts directories
RUN mkdir -p /app/data /app/scripts && chown -R jinwoo:jinwoo /app

USER jinwoo

EXPOSE 8585

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD nc -z localhost 8585 || exit 1

ENTRYPOINT ["/app/jinwoo-channel"]

