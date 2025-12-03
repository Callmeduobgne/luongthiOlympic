/*
 * Copyright (c) 2025 IBN Network
 */

/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 */

/**
 * Batch status enum
 */
export const BatchStatus = {
  PENDING: 'pending',
  PROCESSING: 'processing',
  COMPLETED: 'completed',
  FAILED: 'failed',
  CANCELLED: 'cancelled',
} as const

export type BatchStatus = typeof BatchStatus[keyof typeof BatchStatus]

/**
 * Batch status constants
 */
export const BATCH_STATUS = {
  PENDING: BatchStatus.PENDING,
  PROCESSING: BatchStatus.PROCESSING,
  COMPLETED: BatchStatus.COMPLETED,
  FAILED: BatchStatus.FAILED,
  CANCELLED: BatchStatus.CANCELLED,
} as const

