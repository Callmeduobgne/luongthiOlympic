# Multi-stage Dockerfile for Admin Service

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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/admin-service cmd/server/main.go

# Stage 2: Runtime
FROM alpine:3.19

# Install ca-certificates for HTTPS and tools for downloading peer CLI
# Also install libc6-compat for peer CLI compatibility
RUN apk --no-cache add ca-certificates tzdata wget tar bash libc6-compat

# Download and install Hyperledger Fabric peer CLI
# Using Fabric 2.5.9 to match the network version
RUN cd /tmp && \
    wget -q https://github.com/hyperledger/fabric/releases/download/v2.5.9/hyperledger-fabric-linux-amd64-2.5.9.tar.gz && \
    tar -xzf hyperledger-fabric-linux-amd64-2.5.9.tar.gz && \
    mkdir -p /usr/local/bin && \
    cp /tmp/bin/peer /usr/local/bin/peer && \
    chmod +x /usr/local/bin/peer && \
    cd / && \
    rm -rf /tmp/hyperledger-fabric-linux-amd64-2.5.9.tar.gz /tmp/bin /tmp/config /tmp/fabric-samples

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/admin-service .

# Set ownership
RUN chown -R appuser:appuser /app

# Note: Fabric certificates will be mounted as volume in docker-compose.yml
# Note: Peer CLI is installed at /usr/local/bin/peer and available in PATH

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

# Run application
CMD ["./admin-service"]

