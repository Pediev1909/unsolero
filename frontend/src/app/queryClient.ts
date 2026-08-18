import { QueryClient } from '@tanstack/react-query'

import { ApiError } from '../lib/api/client'

export function shouldRetryQuery(
  failureCount: number,
  error: unknown,
): boolean {
  if (failureCount >= 1) return false
  if (error instanceof ApiError) {
    return error.status === 0 || error.status >= 500
  }
  return true
}

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: shouldRetryQuery,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
})
