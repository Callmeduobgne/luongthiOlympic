/**
 * Application constants
 */

export const APP_NAME = 'IBN Network'
export const APP_VERSION = '1.0.0'

/**
 * Status constants
 */
export const BATCH_STATUS = {
  CREATED: 'CREATED',
  PROCESSING: 'PROCESSING',
  VERIFIED: 'VERIFIED',
  SHIPPED: 'SHIPPED',
  DELIVERED: 'DELIVERED',
  FAILED: 'FAILED',
} as const

export type BatchStatus = typeof BATCH_STATUS[keyof typeof BATCH_STATUS]

/**
 * Channel constants
 */
export const DEFAULT_CHANNEL = 'ibnchannel'

/**
 * Chaincode constants
 */
export const CHAINCODE_NAMES = {
  TEA_TRACE: 'teaTraceCC',
} as const

/**
 * Pagination constants
 */
export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 10,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
} as const

/**
 * Date format constants
 */
export const DATE_FORMATS = {
  DISPLAY: 'dd/MM/yyyy',
  DISPLAY_WITH_TIME: 'dd/MM/yyyy HH:mm',
  ISO: "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'",
} as const

/**
 * API timeout constants
 */
export const API_TIMEOUT = {
  DEFAULT: 30000, // 30 seconds
  UPLOAD: 60000, // 60 seconds
} as const

