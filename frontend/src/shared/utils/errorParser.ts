/*
 * Copyright (c) 2025 IBN Network
 */

/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 */

/**
 * Extract error message from error object
 */
export function extractErrorMessage(error: unknown): string {
  if (typeof error === 'string') {
    return error
  }

  if (error instanceof Error) {
    return error.message
  }

  if (typeof error === 'object' && error !== null) {
    const err = error as Record<string, unknown>
    
    // Try common error message fields
    if (typeof err.message === 'string') {
      return err.message
    }
    
    if (typeof err.error === 'string') {
      return err.error
    }
    
    if (typeof err.msg === 'string') {
      return err.msg
    }
  }

  return 'An unknown error occurred'
}

