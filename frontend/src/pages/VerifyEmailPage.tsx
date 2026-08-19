import { useEffect, useMemo, useRef, useState } from 'react'
import { Link } from 'react-router-dom'

import { AuthLayout } from '../components/layout/AuthLayout'
import { LoadingState } from '../components/ui/LoadingState'
import { completeEmailVerification } from '../features/auth/securityApi'
import { ApiError } from '../lib/api/client'

export function VerifyEmailPage() {
  const token = useMemo(() => window.location.hash.slice(1), [])
  const verification = useRef<Promise<void> | null>(null)
  const [state, setState] = useState<'pending' | 'complete' | 'error'>(
    token ? 'pending' : 'error',
  )
  const [message, setMessage] = useState(
    token ? '' : 'This verification link is missing its security token.',
  )

  useEffect(() => {
    if (!token) return
    verification.current ??= completeEmailVerification(token)
    let active = true
    void verification.current
      .then(() => {
        if (!active) return
        window.history.replaceState(null, '', window.location.pathname)
        setState('complete')
      })
      .catch((error: unknown) => {
        if (!active) return
        setMessage(
          error instanceof ApiError
            ? error.message
            : 'The address could not be verified.',
        )
        setState('error')
      })
    return () => {
      active = false
    }
  }, [token])

  return (
    <AuthLayout
      description="Verification tokens are hashed at rest, expire, and can be used once."
      documentDescription="Completing email verification for your UNSOLERO account."
      documentTitle="Verify your email | UNSOLERO"
      eyebrow="Account security"
      title="Verify your email."
    >
      {state === 'pending' && (
        <LoadingState
          description="Checking the one-time verification token."
          title="Verifying address"
        />
      )}
      {state === 'complete' && (
        <div role="status">
          <h2 className="text-2xl font-medium">Email verified</h2>
          <p className="mt-4 text-sm text-ink/70">
            Your account now has a verified address.
          </p>
          <Link
            className="mt-6 inline-block font-semibold underline underline-offset-4"
            to="/login"
          >
            Continue to sign in
          </Link>
        </div>
      )}
      {state === 'error' && (
        <div role="alert">
          <h2 className="text-2xl font-medium">Verification unavailable</h2>
          <p className="mt-4 text-sm text-ink/70">{message}</p>
          <Link
            className="mt-6 inline-block font-semibold underline underline-offset-4"
            to="/login"
          >
            Return to sign in
          </Link>
        </div>
      )}
    </AuthLayout>
  )
}
