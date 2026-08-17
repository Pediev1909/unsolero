import { Navigate, Outlet, useLocation } from 'react-router-dom'

import { ErrorState } from '../../components/ui/ErrorState'
import { LoadingState } from '../../components/ui/LoadingState'
import { useCurrentUser } from './queries'

export function ProtectedRoute() {
  const account = useCurrentUser()
  const location = useLocation()

  if (account.isPending) {
    return (
      <main className="grid min-h-screen place-items-center px-4">
        <LoadingState
          description="Securely checking your session."
          title="Checking your account"
        />
      </main>
    )
  }

  if (account.isError) {
    return (
      <main className="grid min-h-screen place-items-center px-6">
        <ErrorState
          description="Check your connection and try again."
          onRetry={() => void account.refetch()}
          title="We could not verify your session."
        />
      </main>
    )
  }

  if (!account.data) {
    return <Navigate replace state={{ from: location.pathname }} to="/login" />
  }

  return <Outlet />
}
