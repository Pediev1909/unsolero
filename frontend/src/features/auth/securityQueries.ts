import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { authKeys } from './queries'
import {
  beginMfaEnrollment,
  changePassword,
  deleteAccount,
  listSessions,
  regenerateRecoveryCodes,
  requestEmailVerification,
  revokeOtherSessions,
  revokeSession,
  stepUpMfa,
  verifyMfaEnrollment,
} from './securityApi'

export const securityKeys = {
  sessions: ['account', 'security', 'sessions'] as const,
}

export function useSecuritySessions() {
  return useQuery({ queryKey: securityKeys.sessions, queryFn: listSessions })
}

export function useRevokeSession() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: revokeSession,
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: securityKeys.sessions }),
  })
}

export function useRevokeOtherSessions() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: revokeOtherSessions,
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: securityKeys.sessions }),
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: ({ current, next }: { current: string; next: string }) =>
      changePassword(current, next),
  })
}

export function useRequestVerification() {
  return useMutation({ mutationFn: requestEmailVerification })
}

export function useBeginMfa() {
  return useMutation({ mutationFn: beginMfaEnrollment })
}

export function useVerifyMfa() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: verifyMfaEnrollment,
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: authKeys.currentUser }),
  })
}

export function useRegenerateRecoveryCodes() {
  return useMutation({ mutationFn: regenerateRecoveryCodes })
}

export function useMfaStepUp() {
  return useMutation({ mutationFn: stepUpMfa })
}

export function useDeleteAccount() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      password,
      confirmation,
    }: {
      password: string
      confirmation: string
    }) => deleteAccount(password, confirmation),
    onSuccess: () => {
      queryClient.setQueryData(authKeys.currentUser, null)
      queryClient.clear()
    },
  })
}
