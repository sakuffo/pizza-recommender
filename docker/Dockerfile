# syntax=docker/dockerfile:1

# --- Build Stage ---
FROM golang:1.24.1-alpine AS builder

WORKDIR /build

# Restore dependencies using caches
# Copy go.mod and go.sum first to leverage Docker cache
COPY app/go.mod app/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy the entire app source code
COPY app/ ./

# Build the Go app
# PostgreSQL driver (lib/pq) is pure Go, no CGO needed
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/pizza-backend .

# --- Final Stage ---
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the compiled binary from the builder stage
COPY --from=builder /app/pizza-backend /app/pizza-backend

# Switch to the non-root user
USER appuser

# Expose the port the app listens on
EXPOSE 8080

# Command to run the application
CMD ["/app/pizza-backend"] 