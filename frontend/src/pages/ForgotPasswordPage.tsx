import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'

import { AuthLayout } from '../components/layout/AuthLayout'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { requestPasswordReset } from '../features/auth/securityApi'

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [pending, setPending] = useState(false)
  const [complete, setComplete] = useState(false)
  const [error, setError] = useState('')

  async function submit(event: FormEvent) {
    event.preventDefault()
    setPending(true)
    setError('')
    try {
      await requestPasswordReset(email)
      setComplete(true)
    } catch {
      setError('The reset request could not be recorded. Please try again.')
    } finally {
      setPending(false)
    }
  }

  return (
    <AuthLayout
      description="For privacy, the response is identical whether or not the address belongs to an account."
      documentDescription="Request a password reset link for your UNSOLERO account."
      documentTitle="Reset your password | UNSOLERO"
      eyebrow="Account recovery"
      title="Reset your password."
    >
      <h2 className="text-2xl font-medium tracking-[-0.035em]">
        Request a secure link
      </h2>
      {complete ? (
        <div className="mt-8" role="status">
          <p className="text-sm leading-6 text-ink/65">
            If the address is eligible, a password-reset delivery intent has
            been recorded. The link expires after a bounded period.
          </p>
          <Link
            className="mt-6 inline-block font-semibold underline underline-offset-4"
            to="/login"
          >
            Return to sign in
          </Link>
        </div>
      ) : (
        <form
          className="mt-8 space-y-6"
          onSubmit={(event) => void submit(event)}
        >
          <Input
            autoComplete="email"
            label="Email"
            onChange={(event) => setEmail(event.target.value)}
            required
            type="email"
            value={email}
          />
          {error && (
            <p className="text-sm text-danger" role="alert">
              {error}
            </p>
          )}
          <Button
            fullWidth
            loading={pending}
            loadingLabel="Recording request…"
            type="submit"
          >
            Request reset
          </Button>
        </form>
      )}
    </AuthLayout>
  )
}
