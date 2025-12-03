/*
 * Copyright (c) 2025 IBN Network
 */

/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 */

export interface PackageWarning {
  level: 'warning' | 'error'
  message: string
}

/**
 * Validate chaincode package
 */
export function validateChaincodePackage(file: File): { valid: boolean; errors: string[] } {
  const errors: string[] = []

  // Check file type
  if (!file.name.endsWith('.tar.gz') && !file.name.endsWith('.tgz')) {
    errors.push('Package must be a .tar.gz or .tgz file')
  }

  // Check file size (max 10MB)
  const maxSize = 10 * 1024 * 1024 // 10MB
  if (file.size > maxSize) {
    errors.push(`Package size must be less than ${maxSize / 1024 / 1024}MB`)
  }

  return {
    valid: errors.length === 0,
    errors,
  }
}

/**
 * Get package warnings
 */
export function getPackageWarnings(file: File): PackageWarning[] {
  const warnings: PackageWarning[] = []

  // Check file size (warn if > 5MB)
  const warnSize = 5 * 1024 * 1024 // 5MB
  if (file.size > warnSize) {
    warnings.push({
      level: 'warning',
      message: 'Package size is large, may take longer to deploy',
    })
  }

  return warnings
}

