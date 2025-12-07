/**
 * API response types
 */

export interface ApiResponse<T = unknown> {
  data: T
  message?: string
  success?: boolean
}

export interface ApiError {
  message: string
  status: number
  code?: string
  errors?: Record<string, string[]>
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface PaginationParams {
  page?: number
  pageSize?: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

