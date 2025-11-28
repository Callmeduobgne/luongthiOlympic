# Multi-stage Dockerfile for Go application

# Stage 1: Build
FROM golang:1.23-alpine AS builder

# Build metadata
LABEL maintainer="IBN Network Team"
LABEL description="IBN API Gateway - Blockchain operations proxy"

# Install build dependencies
RUN apk add --no-cache git make ca-certificates

# Set working directory
WORKDIR /app

# Configure Go proxy with fallback
# Use proxy.golang.org first, fallback to direct if proxy fails
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GOTIMEOUT=300s

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies with retry mechanism
# Retry up to 3 times with exponential backoff
RUN for i in 1 2 3; do \
        go mod download && break || \
        (echo "Attempt $i failed, retrying..." && sleep $((i * 5))); \
    done || (echo "All download attempts failed" && exit 1)

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api-gateway cmd/server/main.go

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies (including wget for healthcheck)
RUN apk --no-cache add ca-certificates tzdata wget

# Note: Peer CLI removed - lifecycle operations are now handled by Admin Service
# API Gateway only handles business transactions (invoke/query) via Fabric Gateway SDK

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/api-gateway .

# Set ownership
RUN chown -R appuser:appuser /app

# Note: Fabric certificates will be mounted as volume in docker-compose.yml
# Note: Peer CLI is installed at /usr/local/bin/peer and available in PATH

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check uses GET instead of HEAD to match handler implementation
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health >/dev/null || exit 1

# Run application
CMD ["./api-gateway"]

