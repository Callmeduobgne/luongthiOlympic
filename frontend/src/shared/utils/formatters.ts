import { format, formatDistanceToNow, parseISO, isValid } from 'date-fns'

/**
 * Format date to Vietnamese locale
 */
export const formatDate = (date: Date | string | null | undefined, formatStr: string = 'dd/MM/yyyy'): string => {
  if (!date) return 'N/A'
  const dateObj = typeof date === 'string' ? parseISO(date) : date
  if (!isValid(dateObj)) return 'N/A'
  return format(dateObj, formatStr)
}

/**
 * Format date to relative time (e.g., "2 hours ago")
 */
export const formatRelativeTime = (date: Date | string | null | undefined): string => {
  if (!date) return 'N/A'
  const dateObj = typeof date === 'string' ? parseISO(date) : date
  if (!isValid(dateObj)) return 'N/A'
  return formatDistanceToNow(dateObj, { addSuffix: true })
}

/**
 * Format number with thousand separators
 */
export const formatNumber = (num: number): string => {
  return new Intl.NumberFormat('vi-VN').format(num)
}

/**
 * Format currency
 */
export const formatCurrency = (amount: number, currency: string = 'VND'): string => {
  return new Intl.NumberFormat('vi-VN', {
    style: 'currency',
    currency,
  }).format(amount)
}

/**
 * Format file size
 */
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

/**
 * Truncate string with ellipsis
 */
export const truncate = (str: string, length: number): string => {
  if (str.length <= length) return str
  return str.substring(0, length) + '...'
}

/**
 * Format hash (show first and last characters)
 */
export const formatHash = (hash: string, start: number = 6, end: number = 4): string => {
  if (hash.length <= start + end) return hash
  return `${hash.substring(0, start)}...${hash.substring(hash.length - end)}`
}

