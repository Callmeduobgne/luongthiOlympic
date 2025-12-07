/**
 * Common types used across the application
 */

export type Status = 'idle' | 'loading' | 'success' | 'error'

export interface BaseEntity {
  id: string
  createdAt: string
  updatedAt: string
}

export interface SelectOption<T = string> {
  label: string
  value: T
  disabled?: boolean
}

export interface TableColumn<T = unknown> {
  key: string
  label: string
  sortable?: boolean
  render?: (value: unknown, row: T) => React.ReactNode
}

export interface FilterOption {
  key: string
  label: string
  value: string | number | boolean
}

export interface SortOption {
  field: string
  order: 'asc' | 'desc'
}

