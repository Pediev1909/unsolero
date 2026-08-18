import { useEffect, type PropsWithChildren } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'

import { ToastProvider } from '../components/ui/ToastProvider'
import { authKeys } from '../features/auth/queries'
import { onAuthenticationExpired } from '../lib/api/client'
import { queryClient } from './queryClient'

export function AppProviders({ children }: PropsWithChildren) {
  useEffect(
    () =>
      onAuthenticationExpired(() => {
        queryClient.setQueryData(authKeys.currentUser, null)
        queryClient.removeQueries({ queryKey: ['account'] })
        queryClient.removeQueries({ queryKey: ['admin'] })
      }),
    [],
  )
  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>{children}</ToastProvider>
    </QueryClientProvider>
  )
}
