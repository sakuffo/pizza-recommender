# syntax=docker/dockerfile:1

# --- Build Stage ---
FROM golang:1.24.1-alpine AS builder

WORKDIR /build

# Install build dependencies if needed (e.g., for CGO like go-sqlite3)
# We need gcc and potentially other build tools for CGO
RUN apk add --no-cache build-base

# Restore dependencies using caches
# Copy go.mod and go.sum first to leverage Docker cache
COPY app/go.mod app/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy the entire app source code
# Make sure to copy ALL .go files needed for the build
COPY app/ ./

# Build the Go app
# Use 'go build .' to build all go files in the current directory (/build)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/pizza-backend .
    # Changed CGO_ENABLED=1 because go-sqlite3 requires it
    # Changed output path to /app/pizza-backend
    # Changed 'pizza-mix.go' to '.' to build all files

# --- Final Stage ---
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies if needed (e.g., CA certificates, SQLite libs)
# go-sqlite3 is CGO based and links against libc, which is standard in alpine.
# If using a different base or needing specific SQLite features, might need:
# RUN apk add --no-cache sqlite-libs
RUN apk add --no-cache ca-certificates tzdata sqlite-libs # Add sqlite-libs

# Create a non-root user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the compiled binary from the builder stage
COPY --from=builder /app/pizza-backend /app/pizza-backend

# Create the data directory and change its ownership to the non-root user
# Do this *before* switching the user context
RUN mkdir /app/data && chown appuser:appgroup /app/data

# Switch to the non-root user
USER appuser

# Expose the port the app listens on (optional but good practice)
EXPOSE 8080

# Command to run the application
# The database will be created in /app/data/pizza.db which now exists and is writable
CMD ["/app/pizza-backend"] 