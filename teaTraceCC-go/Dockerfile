# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /chaincode

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the chaincode
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o chaincode .

# Production stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /chaincode

# Copy the binary from builder
COPY --from=builder /chaincode/chaincode .

# Chaincode listens on this port
EXPOSE 9999

# Set environment variables
ENV CHAINCODE_SERVER_ADDRESS=0.0.0.0:9999
ENV CHAINCODE_ID=""
ENV CORE_CHAINCODE_ID_NAME=""

# Run the chaincode
CMD ["./chaincode"]
