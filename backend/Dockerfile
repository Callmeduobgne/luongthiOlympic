# Build stage
FROM golang:alpine AS builder

# Build metadata
LABEL maintainer="IBN Network Team"
LABEL description="IBN Backend API - Blockchain traceability system"

# Install build dependencies
RUN apk add --no-cache git make ca-certificates

# Install swag for Swagger docs generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

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

# Generate Swagger documentation
RUN mkdir -p docs/swagger && \
    export PATH=$PATH:$(go env GOPATH)/bin && \
    swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal || true

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/server ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

# Set timezone
ENV TZ=Asia/Ho_Chi_Minh

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/server /app/server

# Copy migrations (if needed)
COPY --from=builder /app/migrations /app/migrations

# Change ownership
RUN chown -R app:app /app

# Switch to non-root user
USER app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/server"]

