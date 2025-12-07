import { z } from 'zod'

/**
 * Common validation schemas
 */

export const emailSchema = z.string().email('Invalid email address')

export const passwordSchema = z
  .string()
  .min(8, 'Password must be at least 8 characters')
  .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
  .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
  .regex(/[0-9]/, 'Password must contain at least one number')

export const batchIdSchema = z
  .string()
  .min(1, 'Batch ID is required')
  .max(50, 'Batch ID must be less than 50 characters')
  .regex(/^[A-Z0-9_-]+$/, 'Batch ID can only contain uppercase letters, numbers, hyphens, and underscores')

export const hashSchema = z
  .string()
  .regex(/^[a-f0-9]{64}$/i, 'Hash must be a valid SHA-256 hash (64 hex characters)')

export const urlSchema = z.string().url('Invalid URL')

/**
 * Validation helper functions
 */
export const validateEmail = (email: string): boolean => {
  return emailSchema.safeParse(email).success
}

export const validatePassword = (password: string): boolean => {
  return passwordSchema.safeParse(password).success
}

export const validateBatchId = (batchId: string): boolean => {
  return batchIdSchema.safeParse(batchId).success
}

