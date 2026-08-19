import { useState, type FormEvent } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'

import { AuthLayout } from '../components/layout/AuthLayout'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { authKeys } from '../features/auth/queries'
import { synchronizeAnalyticsConsentAfterAuthentication } from '../features/analytics/consent'
import { completeMfaLogin } from '../features/auth/securityApi'
import { ApiError } from '../lib/api/client'

export function MfaLoginPage() {
  const [code, setCode] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  async function submit(event: FormEvent) {
    event.preventDefault()
    setPending(true)
    setError('')
    try {
      const user = await completeMfaLogin(code)
      queryClient.setQueryData(authKeys.currentUser, user)
      void synchronizeAnalyticsConsentAfterAuthentication().catch(() => {
        // Authentication succeeds independently; analytics remains fail-closed.
      })
      await navigate('/account', { replace: true })
    } catch (caught) {
      setError(
        caught instanceof ApiError
          ? caught.message
          : 'Authentication could not be completed.',
      )
    } finally {
      setPending(false)
    }
  }
  return (
    <AuthLayout
      description="Use the six-digit code from your authenticator, or one unused recovery code."
      documentDescription="Enter your authenticator code to finish signing in to UNSOLERO."
      documentTitle="Two-step verification | UNSOLERO"
      eyebrow="Privileged authentication"
      title="Complete secure sign in."
    >
      <form className="space-y-6" onSubmit={(event) => void submit(event)}>
        <Input
          autoComplete="one-time-code"
          label="Authentication or recovery code"
          onChange={(event) => setCode(event.target.value)}
          required
          value={code}
        />
        {error && (
          <p className="text-sm text-danger" role="alert">
            {error}
          </p>
        )}
        <Button
          fullWidth
          loading={pending}
          loadingLabel="Verifying…"
          type="submit"
        >
          Verify and sign in
        </Button>
      </form>
    </AuthLayout>
  )
}
