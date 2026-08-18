import { useState, type FormEvent } from 'react'

import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import type { AuthUser } from '../schemas'
import {
  useBeginMfa,
  useMfaStepUp,
  useRegenerateRecoveryCodes,
  useVerifyMfa,
} from '../securityQueries'
import type { MFAEnrollment } from '../securitySchemas'

export function MfaSecurity({ user }: { user: AuthUser }) {
  const begin = useBeginMfa()
  const verify = useVerifyMfa()
  const regenerate = useRegenerateRecoveryCodes()
  const stepUp = useMfaStepUp()
  const [password, setPassword] = useState('')
  const [code, setCode] = useState('')
  const [enrollment, setEnrollment] = useState<MFAEnrollment | null>(null)
  const [recoveryCodes, setRecoveryCodes] = useState<string[]>([])

  async function beginEnrollment(event: FormEvent) {
    event.preventDefault()
    const result = await begin.mutateAsync(password)
    setPassword('')
    setEnrollment(result)
  }
  async function verifyEnrollment(event: FormEvent) {
    event.preventDefault()
    const result = await verify.mutateAsync(code)
    setCode('')
    setEnrollment(null)
    setRecoveryCodes(result)
  }
  async function refreshStepUp(event: FormEvent) {
    event.preventDefault()
    await stepUp.mutateAsync(code)
    setCode('')
  }

  return (
    <section aria-labelledby="mfa-heading">
      <h3 className="text-xl font-semibold" id="mfa-heading">
        Multi-factor authentication
      </h3>
      <p className="mt-3 max-w-2xl text-sm leading-6 text-ink/60">
        TOTP secrets are encrypted server-side. Recovery codes are shown once
        and stored only as hashes.
      </p>
      {recoveryCodes.length > 0 && (
        <div
          className="mt-5 border border-bronze/40 bg-paper p-5"
          role="status"
        >
          <p className="font-semibold">Save these recovery codes now</p>
          <p className="mt-2 text-sm text-ink/60">
            Each code works once. They will not be shown again.
          </p>
          <div className="mt-4 grid gap-2 sm:grid-cols-2">
            {recoveryCodes.map((item) => (
              <code className="select-all text-sm" key={item}>
                {item}
              </code>
            ))}
          </div>
          <Button
            className="mt-5"
            onClick={() => setRecoveryCodes([])}
            size="sm"
            variant="secondary"
          >
            I saved them
          </Button>
        </div>
      )}
      {enrollment && (
        <div className="mt-5 border border-ink/15 p-5">
          <p className="font-semibold">Add UNSOLERO to your authenticator</p>
          <p className="mt-2 text-sm text-ink/60">
            Enter this secret manually. It disappears after verification.
          </p>
          <code className="mt-4 block break-all bg-paper p-3 text-sm">
            {enrollment.secret}
          </code>
          <form
            className="mt-5 flex max-w-lg flex-col gap-4 sm:flex-row sm:items-end"
            onSubmit={(event) => void verifyEnrollment(event)}
          >
            <Input
              autoComplete="one-time-code"
              containerClassName="flex-1"
              inputMode="numeric"
              label="Six-digit code"
              onChange={(event) => setCode(event.target.value)}
              required
              value={code}
            />
            <Button loading={verify.isPending} type="submit">
              Verify enrollment
            </Button>
          </form>
          {verify.isError && (
            <p className="mt-3 text-sm text-danger" role="alert">
              {verify.error.message}
            </p>
          )}
        </div>
      )}
      {!user.mfa_enabled && !enrollment && recoveryCodes.length === 0 && (
        <form
          className="mt-5 max-w-md space-y-4"
          onSubmit={(event) => void beginEnrollment(event)}
        >
          <Input
            autoComplete="current-password"
            label="Current password"
            onChange={(event) => setPassword(event.target.value)}
            required
            type="password"
            value={password}
          />
          {begin.isError && (
            <p className="text-sm text-danger" role="alert">
              {begin.error.message}
            </p>
          )}
          <Button loading={begin.isPending} type="submit">
            Begin MFA enrollment
          </Button>
        </form>
      )}
      {user.mfa_enabled && recoveryCodes.length === 0 && (
        <div className="mt-5 grid gap-8 md:grid-cols-2">
          <form
            className="space-y-4"
            onSubmit={(event) => void refreshStepUp(event)}
          >
            <p className="font-semibold">Refresh privileged verification</p>
            <Input
              autoComplete="one-time-code"
              label="Authenticator or recovery code"
              onChange={(event) => setCode(event.target.value)}
              required
              value={code}
            />
            {stepUp.isError && (
              <p className="text-sm text-danger" role="alert">
                {stepUp.error.message}
              </p>
            )}
            {stepUp.isSuccess && (
              <p className="text-sm" role="status">
                Privileged verification refreshed.
              </p>
            )}
            <Button loading={stepUp.isPending} size="sm" type="submit">
              Verify
            </Button>
          </form>
          <div>
            <p className="font-semibold">Replace recovery codes</p>
            <p className="mt-2 text-sm text-ink/60">
              Enter a current code above, then generate a completely new set.
            </p>
            <Button
              className="mt-4"
              disabled={!code}
              loading={regenerate.isPending}
              onClick={() =>
                regenerate.mutate(code, {
                  onSuccess: (items) => {
                    setRecoveryCodes(items)
                    setCode('')
                  },
                })
              }
              size="sm"
              variant="secondary"
            >
              Regenerate codes
            </Button>
          </div>
        </div>
      )}
    </section>
  )
}
