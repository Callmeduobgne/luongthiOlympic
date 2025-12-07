import toast from 'react-hot-toast'

export interface ApiError {
  message: string
  status: number
  code?: string
  name: string
}

export const createApiError = (
  message: string,
  status: number,
  code?: string
): ApiError => {
  const error = new Error(message) as ApiError
  error.status = status
  error.code = code
  error.name = 'ApiError'
  return error
}

interface ApiErrorResponse {
  response?: {
    status: number
    data?: {
      message?: string
      code?: string
    }
  }
  request?: unknown
}

export const handleApiError = (error: unknown) => {
  const apiError = error as ApiErrorResponse
  
  if (apiError.response) {
    // Server responded with error
    const { status, data } = apiError.response

    switch (status) {
      case 400:
        toast.error(data?.message || 'Invalid request')
        break
      case 401:
        toast.error('Unauthorized. Please login again.')
        break
      case 403:
        toast.error('You do not have permission to perform this action')
        break
      case 404:
        toast.error('Resource not found')
        break
      case 500:
        toast.error('Server error. Please try again later.')
        break
      default:
        toast.error(data?.message || 'An error occurred')
    }

    throw createApiError(data?.message || 'An error occurred', status, data?.code)
  } else if (apiError.request) {
    // Request made but no response
    toast.error('Network error. Please check your connection.')
    throw createApiError('Network error', 0)
  } else {
    // Something else happened
    toast.error('An unexpected error occurred')
    throw error
  }
}

