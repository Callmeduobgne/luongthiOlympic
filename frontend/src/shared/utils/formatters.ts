/*
 * Copyright (c) 2025 IBN Network
 */

/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 */

/**
 * Format date to string
 */
export function formatDate(date: Date | string | number, format: string = 'dd/MM/yyyy'): string {
  const d = typeof date === 'string' || typeof date === 'number' ? new Date(date) : date
  
  if (isNaN(d.getTime())) {
    return ''
  }

  const day = String(d.getDate()).padStart(2, '0')
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const year = d.getFullYear()

  return format
    .replace(/dd/g, day)
    .replace(/MM/g, month)
    .replace(/yyyy/g, String(year))
    .replace(/HH/g, String(d.getHours()).padStart(2, '0'))
    .replace(/mm/g, String(d.getMinutes()).padStart(2, '0'))
    .replace(/ss/g, String(d.getSeconds()).padStart(2, '0'))
}

/**
 * Format hash string (truncate if too long)
 */
export function formatHash(hash: string, length: number = 8): string {
  if (!hash) return ''
  if (hash.length <= length * 2) return hash
  return `${hash.substring(0, length)}...${hash.substring(hash.length - length)}`
}

