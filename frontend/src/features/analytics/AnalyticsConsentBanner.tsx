import { useEffect, useState } from 'react'

import { Button } from '../../components/ui/Button'
import { Container } from '../../components/ui/Container'
import {
  getAnalyticsConsent,
  openAnalyticsConsentEvent,
  setAnalyticsConsent,
} from './consent'

export function AnalyticsConsentBanner() {
  const [open, setOpen] = useState(() => getAnalyticsConsent() === 'unknown')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    const showPreferences = () => setOpen(true)
    window.addEventListener(openAnalyticsConsentEvent, showPreferences)
    return () =>
      window.removeEventListener(openAnalyticsConsentEvent, showPreferences)
  }, [])

  if (!open) return null

  const choose = async (consent: 'granted' | 'denied') => {
    setPending(true)
    setError('')
    try {
      await setAnalyticsConsent(consent, 'banner')
      setOpen(false)
    } catch {
      setError(
        'We could not save this preference. Optional analytics remain disabled.',
      )
    } finally {
      setPending(false)
    }
  }

  return (
    <aside
      aria-label="Analytics preferences"
      className="fixed inset-x-0 bottom-0 z-50 border-t border-ink/15 bg-canvas py-4 shadow-elevated"
      role="region"
    >
      <Container className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <p className="max-w-2xl text-sm leading-6 text-ink/70">
          Optional first-party analytics help us improve product decisions. We
          do not send them until you allow them. Merchant clicks are recorded
          when you choose “View at Merchant.”
        </p>
        <div className="flex shrink-0 gap-2">
          <Button
            disabled={pending}
            onClick={() => void choose('denied')}
            size="sm"
            variant="secondary"
          >
            Decline
          </Button>
          <Button
            disabled={pending}
            loading={pending}
            loadingLabel="Saving preference…"
            onClick={() => void choose('granted')}
            size="sm"
            variant="secondary"
          >
            Allow analytics
          </Button>
        </div>
        {error && (
          <p className="text-sm text-danger" role="alert">
            {error}
          </p>
        )}
      </Container>
    </aside>
  )
}
