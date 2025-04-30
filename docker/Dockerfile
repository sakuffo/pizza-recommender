# syntax=docker/dockerfile:1

ARG GO_VERSION=1.24.1
FROM golang:${GO_VERSION}-alpine AS builder

# Install git for Go modules that require it
RUN --mount=type=cache,target=/var/cache/apk \
    apk add --no-cache git

WORKDIR /build

# Copy go.mod and go.sum first for dependency resolution and caching
COPY --link app/go.mod .
COPY --link app/go.sum .

# Download dependencies (with cache mounts for speed)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy only the Go source code required for the backend
COPY --link app/pizza-mix.go .

# Build the Go binary (static, stripped)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pizza-backend pizza-mix.go

# --- Final minimal image ---
FROM alpine:latest AS final

# Create non-root user
RUN addgroup -S pizza && adduser -S pizza -G pizza

WORKDIR /app

# Install CA certificates for HTTPS (might still be needed if backend makes external calls)
RUN --mount=type=cache,target=/var/cache/apk \
    apk add --no-cache ca-certificates

# Copy only the built backend binary
COPY --from=builder /build/pizza-backend /app/pizza-backend

# Use non-root user
USER pizza

# Backend API listens on 8080
EXPOSE 8080

ENTRYPOINT ["/app/pizza-backend"] 