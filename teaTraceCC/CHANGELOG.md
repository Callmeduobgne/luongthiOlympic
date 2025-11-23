# Changelog

All notable changes to the teaTraceCC chaincode will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2024-11-22

### Added
- **Query Functions**
  - `getAllBatches(limit, offset)` - Query tất cả batches với pagination
  - `getBatchesByStatus(status, limit, offset)` - Query batches theo trạng thái
  - `getBatchesByOwner(owner, limit, offset)` - Query batches theo owner
  - `getBatchHistory(batchId)` - Lấy lịch sử thay đổi của batch

- **Input Validation**
  - Validate batch ID format
  - Validate date format (YYYY-MM-DD)
  - Validate string inputs với max length
  - Validate pagination parameters

- **Code Quality Tools**
  - ESLint configuration (.eslintrc.json)
  - Prettier configuration (.prettierrc)
  - EditorConfig (.editorconfig)
  - Pre-commit hooks script

### Improved
- **Error Handling**
  - Better error messages với validation details
  - Consistent error format

- **Documentation**
  - CHANGELOG.md
  - Updated README với các tính năng mới

### Changed
- Input validation được thêm vào tất cả public methods
- Error messages được cải thiện với thông tin chi tiết hơn

---

## [1.0.0] - 2024-11-08

### Added
- **Initial Release** - Tea Traceability Chaincode
- **Core Functions**
  - `createBatch` - Tạo lô trà mới
  - `verifyBatch` - Xác minh hash của lô trà
  - `getBatchInfo` - Query thông tin lô trà
  - `updateBatchStatus` - Cập nhật trạng thái

- **Security Features**
  - MSP-based authorization (Farmer, Verifier, Admin)
  - SHA-256 hash verification
  - Input validation

- **Configuration**
  - MSP config file (msp-config.json)
  - Portable across different networks

- **Documentation**
  - README.md với hướng dẫn đầy đủ
  - API reference
  - Examples và troubleshooting

- **Testing**
  - Unit tests
  - Integration tests
  - Test coverage config

---

[1.1.0]: https://github.com/Callmeduobgne/luongthiOlympic/releases/tag/teaTraceCC-v1.1.0
[1.0.0]: https://github.com/Callmeduobgne/luongthiOlympic/releases/tag/teaTraceCC-v1.0.0

