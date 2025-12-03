/*
 * Copyright (c) 2025 IBN Network
 */

/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 */

import { z } from 'zod'

/**
 * Batch ID validation schema
 */
export const batchIdSchema = z.string()
  .min(1, 'Batch ID is required')
  .max(100, 'Batch ID must be less than 100 characters')
  .regex(/^[a-zA-Z0-9_-]+$/, 'Batch ID can only contain letters, numbers, underscores, and hyphens')

