# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app

# Security scanning stage
FROM aquasec/trivy:latest AS trivy
COPY --from=builder /app/main /app/main
RUN trivy filesystem --no-progress --severity HIGH,CRITICAL /app/main

FROM snyk/snyk:golang AS snyk
COPY --from=builder /app /app
WORKDIR /app
RUN snyk test --severity-threshold=high

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates curl tzdata

# Create non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy migrations directory
COPY --from=builder /app/internal/migrations /app/migrations

# Copy environment file
COPY --from=builder /app/.env.development ./.env.development
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"] 