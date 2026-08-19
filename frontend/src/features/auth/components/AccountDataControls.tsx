import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'

import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { useDeleteAccount } from '../securityQueries'

export function AccountDataControls() {
  const deletion = useDeleteAccount()
  const navigate = useNavigate()
  const [password, setPassword] = useState('')
  const [confirmation, setConfirmation] = useState('')
  async function submit(event: FormEvent) {
    event.preventDefault()
    await deletion.mutateAsync({ password, confirmation })
    await navigate('/', { replace: true })
  }
  return (
    <section aria-labelledby="data-controls-heading">
      <h3 className="text-xl font-semibold" id="data-controls-heading">
        Your data
      </h3>
      <div className="mt-5 grid gap-10 md:grid-cols-2">
        <div>
          <p className="font-semibold">Structured account export</p>
          <p className="mt-2 text-sm leading-6 text-ink/70">
            Includes only your appropriate account, profile, wishlist, setups,
            recommendations, consent history, exportable validated analytics,
            and security metadata. Secrets are excluded.
          </p>
          <a
            className="mt-5 inline-flex min-h-10 items-center border border-ink/20 px-4 text-sm font-semibold hover:border-ink"
            href="/api/account/export"
          >
            Download JSON export
          </a>
        </div>
        <form
          className="space-y-4 border-l-2 border-danger/50 pl-5"
          onSubmit={(event) => void submit(event)}
        >
          <div>
            <p className="font-semibold">Delete account</p>
            <p className="mt-2 text-sm leading-6 text-ink/70">
              Owned planning data is removed. Recommendation, commerce, and
              immutable audit history is retained only after personal
              identifiers are removed.
            </p>
          </div>
          <Input
            autoComplete="current-password"
            label="Current password"
            onChange={(event) => setPassword(event.target.value)}
            required
            type="password"
            value={password}
          />
          <Input
            autoComplete="off"
            hint="Type DELETE exactly."
            label="Confirmation"
            onChange={(event) => setConfirmation(event.target.value)}
            required
            value={confirmation}
          />
          {deletion.isError && (
            <p className="text-sm text-danger" role="alert">
              {deletion.error.message}
            </p>
          )}
          <Button
            disabled={confirmation !== 'DELETE'}
            loading={deletion.isPending}
            type="submit"
            variant="danger"
          >
            Delete my account
          </Button>
        </form>
      </div>
    </section>
  )
}
