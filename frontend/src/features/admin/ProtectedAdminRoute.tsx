import { Navigate, Outlet, useLocation } from 'react-router-dom'

import { ErrorState } from '../../components/ui/ErrorState'
import { LoadingState } from '../../components/ui/LoadingState'
import { useCurrentUser } from '../auth/queries'

export function ProtectedAdminRoute() {
  const account = useCurrentUser()
  const location = useLocation()

  if (account.isPending) {
    return (
      <main className="grid min-h-screen place-items-center bg-charcoal px-6 text-canvas">
        <LoadingState
          description="Verifying your administrative access."
          title="Checking permissions"
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
          title="We could not verify your permissions."
        />
      </main>
    )
  }
  if (!account.data) {
    return <Navigate replace state={{ from: location.pathname }} to="/login" />
  }
  if (!account.data.roles.includes('admin')) {
    return <Navigate replace to="/account" />
  }
  return <Outlet />
}
