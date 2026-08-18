import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { getCurrentUser, login, logout, register } from './api'
import type { Credentials } from './schemas'

export const authKeys = {
  currentUser: ['auth', 'current-user'] as const,
}

export function useCurrentUser() {
  return useQuery({
    queryKey: authKeys.currentUser,
    queryFn: getCurrentUser,
    retry: false,
    staleTime: 60_000,
  })
}

export function useLogin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (credentials: Credentials) => login(credentials),
    onSuccess: (result) => {
      if ('id' in result) queryClient.setQueryData(authKeys.currentUser, result)
    },
  })
}

export function useRegister() {
  return useMutation({
    mutationFn: (credentials: Credentials) => register(credentials),
  })
}

export function useLogout() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: logout,
    onSuccess: () => {
      queryClient.setQueryData(authKeys.currentUser, null)
      queryClient.removeQueries({ queryKey: ['account'] })
    },
  })
}
