import { useMemo, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'

import { AuthLayout } from '../components/layout/AuthLayout'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { completePasswordReset } from '../features/auth/securityApi'
import { ApiError } from '../lib/api/client'

export function ResetPasswordPage() {
  const token = useMemo(() => window.location.hash.slice(1), [])
  const [password, setPassword] = useState('')
  const [pending, setPending] = useState(false)
  const [complete, setComplete] = useState(false)
  const [error, setError] = useState(
    token ? '' : 'This reset link is missing its security token.',
  )

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (!token) return
    setPending(true)
    setError('')
    try {
      await completePasswordReset(token, password)
      window.history.replaceState(null, '', window.location.pathname)
      setComplete(true)
    } catch (caught) {
      setError(
        caught instanceof ApiError
          ? caught.message
          : 'The password could not be reset.',
      )
    } finally {
      setPending(false)
    }
  }

  return (
    <AuthLayout
      description="A successful reset revokes every existing session. The reset token can be used once."
      eyebrow="Account recovery"
      title="Choose a new password."
    >
      {complete ? (
        <div role="status">
          <h2 className="text-2xl font-medium">Password replaced</h2>
          <p className="mt-4 text-sm text-ink/60">
            All previous sessions have been revoked.
          </p>
          <Link
            className="mt-6 inline-block font-semibold underline underline-offset-4"
            to="/login"
          >
            Sign in securely
          </Link>
        </div>
      ) : (
        <form className="space-y-6" onSubmit={(event) => void submit(event)}>
          <Input
            autoComplete="new-password"
            hint="Use at least 12 characters."
            label="New password"
            minLength={12}
            onChange={(event) => setPassword(event.target.value)}
            required
            type="password"
            value={password}
          />
          {error && (
            <p className="text-sm text-danger" role="alert">
              {error}
            </p>
          )}
          <Button
            disabled={!token}
            fullWidth
            loading={pending}
            loadingLabel="Replacing password…"
            type="submit"
          >
            Replace password
          </Button>
        </form>
      )}
    </AuthLayout>
  )
}
