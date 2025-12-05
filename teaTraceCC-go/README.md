# teaTraceCC-go

Hyperledger Fabric chaincode for tea traceability on the IBN network - Written in Go.

## Overview

This is the Go implementation of the teaTraceCC chaincode, providing complete traceability for tea products from farm to consumer.

## Features

### Batch Management
- `CreateBatch` - Create a new tea batch with farm location, harvest date, processing info
- `VerifyBatch` - Verify batch integrity using hash comparison
- `GetBatchInfo` - Get batch details by ID
- `GetAllBatches` - List all batches with pagination
- `GetBatchesByStatus` - Filter batches by status (CREATED, VERIFIED, EXPIRED)
- `GetBatchesByOwner` - Filter batches by owner
- `GetBatchHistory` - Get complete history of batch changes
- `UpdateBatchStatus` - Update batch status

### Package Management
- `CreatePackage` - Create a new package from a batch
- `VerifyPackage` - Verify package using blockhash
- `GetPackageInfo` - Get package details by ID
- `GetAllPackages` - List all packages with pagination
- `GetPackagesByBatch` - Get packages for a specific batch (uses composite keys)
- `GetPackagesByStatus` - Filter packages by status (CREATED, VERIFIED, SOLD, EXPIRED)
- `GetPackageHistory` - Get complete history of package changes
- `UpdatePackageStatus` - Update package status

## Project Structure

```
teaTraceCC-go/
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── go.sum                  # Dependencies checksum
├── chaincode/
│   └── contract.go         # Main contract implementation
├── models/
│   ├── teabatch.go         # TeaBatch model
│   ├── teapackage.go       # TeaPackage model
│   └── mspconfig.go        # MSP configuration
├── utils/
│   ├── hash.go             # Hash utilities
│   └── validation.go       # Input validation
└── META-INF/
    └── statedb/
        └── couchdb/
            └── indexes/    # CouchDB indexes
```

## Building

```bash
# Download dependencies
go mod tidy

# Build the chaincode
go build -o teaTraceCC

# Run tests
go test ./...
```

## Packaging for Deployment

```bash
# Create chaincode package
peer lifecycle chaincode package teaTraceCC-go.tar.gz \
  --path . \
  --lang golang \
  --label teaTraceCC_1.1.0
```

## Advantages over Node.js Implementation

1. **Faster Build** - Go compiles to binary, no npm install needed
2. **Better Performance** - Native binary execution
3. **Simpler Packaging** - No complex tar.gz structure issues
4. **Native SDK** - Go is Fabric's primary language
5. **Type Safety** - Compile-time type checking
6. **Lower Memory** - Smaller footprint than Node.js

## API Compatibility

This Go implementation is fully compatible with the Node.js teaTraceCC:
- Same data structures (JSON field names match)
- Same function signatures
- Same status values
- Same validation rules

## MSP Configuration

Default MSP roles:
- `Farmer` (Org1MSP) - Create batches and packages
- `Verifier` (Org1MSP) - Verify batches
- `Admin` (Org1MSP) - All operations

## License

Apache License 2.0
