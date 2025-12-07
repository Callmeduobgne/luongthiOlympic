import { useQuery, useMutation } from '@tanstack/react-query'
import type { UseQueryOptions, UseMutationOptions } from '@tanstack/react-query'

/**
 * Wrapper hook for React Query useQuery
 * Provides consistent defaults and error handling
 */
export function useApiQuery<TData = unknown, TError = unknown>(
  queryKey: string[],
  queryFn: () => Promise<TData>,
  options?: Omit<UseQueryOptions<TData, TError>, 'queryKey' | 'queryFn'>
) {
  return useQuery<TData, TError>({
    queryKey,
    queryFn,
    staleTime: 60000, // 1 minute default
    refetchOnWindowFocus: false,
    retry: 1,
    ...options,
  })
}

/**
 * Wrapper hook for React Query useMutation
 * Provides consistent defaults and error handling
 */
export function useApiMutation<TData = unknown, TVariables = unknown, TError = unknown>(
  mutationFn: (variables: TVariables) => Promise<TData>,
  options?: Omit<UseMutationOptions<TData, TError, TVariables>, 'mutationFn'>
) {
  return useMutation<TData, TError, TVariables>({
    mutationFn,
    ...options,
  })
}

