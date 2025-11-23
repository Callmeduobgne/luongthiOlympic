/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * Utility functions for validating chaincode package files
 */

/**
 * Validate chaincode package file before upload
 * Note: Full validation requires parsing tar.gz which is complex in browser
 * This provides basic validation - full validation happens on backend
 */
export const validateChaincodePackage = (file: File): { valid: boolean; error?: string } => {
  // 1. Check file extension
  const fileName = file.name.toLowerCase()
  if (!fileName.endsWith('.tar.gz') && !fileName.endsWith('.gz')) {
    return {
      valid: false,
      error: 'File phải có định dạng .tar.gz hoặc .gz',
    }
  }

  // 2. Check file size (max 100MB)
  const maxSize = 100 * 1024 * 1024 // 100MB
  if (file.size > maxSize) {
    return {
      valid: false,
      error: `File quá lớn. Kích thước tối đa là ${maxSize / (1024 * 1024)}MB`,
    }
  }

  // 3. Check file is not empty
  if (file.size === 0) {
    return {
      valid: false,
      error: 'File không được để trống',
    }
  }

  // 4. Check file name doesn't contain invalid characters
  const invalidChars = /[<>:"|?*\x00-\x1f]/
  if (invalidChars.test(file.name)) {
    return {
      valid: false,
      error: 'Tên file chứa ký tự không hợp lệ',
    }
  }

  // Note: Full package structure validation (metadata.json, code.tar.gz)
  // is done on backend after upload because parsing tar.gz in browser
  // requires complex libraries and may not be reliable

  return { valid: true }
}

/**
 * Get package validation warnings (non-blocking)
 */
export const getPackageWarnings = (file: File): string[] => {
  const warnings: string[] = []

  // Warning if file name doesn't follow common naming convention
  const fileName = file.name.toLowerCase()
  if (!fileName.match(/^[a-z0-9_-]+\.tar\.gz$/)) {
    warnings.push('Tên file nên chỉ chứa chữ cái, số, dấu gạch dưới và dấu gạch ngang')
  }

  // Warning if file is very small (might be empty or corrupted)
  if (file.size < 1024) {
    warnings.push('File có vẻ quá nhỏ. Đảm bảo package đã được tạo đúng cách.')
  }

  // Warning if file is very large (might be uncompressed)
  if (file.size > 50 * 1024 * 1024) {
    warnings.push('File khá lớn. Đảm bảo package đã được nén đúng cách.')
  }

  return warnings
}

