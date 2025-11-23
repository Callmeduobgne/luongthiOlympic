# Multi-stage Dockerfile for Go application

# Stage 1: Build
FROM golang:1.23-alpine AS builder

# Configure DNS resolvers for reliable network access during build
# Use public DNS servers (Google DNS, Cloudflare DNS)
RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf && \
    echo "nameserver 8.8.4.4" >> /etc/resolv.conf && \
    echo "nameserver 1.1.1.1" >> /etc/resolv.conf

# Install build dependencies
RUN apk add --no-cache git make ca-certificates

# Set working directory
WORKDIR /app

# Configure Go proxy - use direct mode to bypass proxy issues
# If proxy.golang.org is not accessible, use direct mode
ENV GOPROXY=direct
ENV GOSUMDB=off

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api-gateway cmd/server/main.go

# Stage 2: Runtime
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

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

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run application
CMD ["./api-gateway"]

