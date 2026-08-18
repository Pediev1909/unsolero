import { useState, type FormEvent } from 'react'

import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import type { AuthUser } from '../schemas'
import { useChangePassword, useRequestVerification } from '../securityQueries'

export function PasswordAndEmailSecurity({ user }: { user: AuthUser }) {
  const verification = useRequestVerification()
  const password = useChangePassword()
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')

  async function submitPassword(event: FormEvent) {
    event.preventDefault()
    await password.mutateAsync({ current: currentPassword, next: newPassword })
    setCurrentPassword('')
    setNewPassword('')
  }

  return (
    <div className="grid gap-10 md:grid-cols-2">
      <section aria-labelledby="email-security-heading">
        <h3 className="text-xl font-semibold" id="email-security-heading">
          Email verification
        </h3>
        <p className="mt-3 text-sm leading-6 text-ink/60">
          {user.email_verified
            ? 'Your account email is verified.'
            : 'Your account email has not been verified yet.'}
        </p>
        {!user.email_verified && (
          <Button
            className="mt-5"
            loading={verification.isPending}
            onClick={() => verification.mutate(user.email)}
            size="sm"
            variant="secondary"
          >
            Record another verification request
          </Button>
        )}
        {verification.isSuccess && (
          <p className="mt-3 text-sm" role="status">
            If eligible, a delivery intent was recorded.
          </p>
        )}
        {verification.isError && (
          <p className="mt-3 text-sm text-danger" role="alert">
            The request could not be recorded.
          </p>
        )}
      </section>

      <section aria-labelledby="password-security-heading">
        <h3 className="text-xl font-semibold" id="password-security-heading">
          Change password
        </h3>
        <p className="mt-3 text-sm leading-6 text-ink/60">
          Changing your password keeps this session and revokes every other
          session.
        </p>
        <form
          className="mt-5 space-y-4"
          onSubmit={(event) => void submitPassword(event)}
        >
          <Input
            autoComplete="current-password"
            label="Current password"
            onChange={(event) => setCurrentPassword(event.target.value)}
            required
            type="password"
            value={currentPassword}
          />
          <Input
            autoComplete="new-password"
            hint="Use at least 12 characters."
            label="New password"
            minLength={12}
            onChange={(event) => setNewPassword(event.target.value)}
            required
            type="password"
            value={newPassword}
          />
          {password.isError && (
            <p className="text-sm text-danger" role="alert">
              {password.error.message}
            </p>
          )}
          {password.isSuccess && (
            <p className="text-sm" role="status">
              Password changed. Other sessions were revoked.
            </p>
          )}
          <Button
            loading={password.isPending}
            loadingLabel="Changing…"
            size="sm"
            type="submit"
          >
            Change password
          </Button>
        </form>
      </section>
    </div>
  )
}
