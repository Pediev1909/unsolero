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

  useEffect(() => {
    const showPreferences = () => setOpen(true)
    window.addEventListener(openAnalyticsConsentEvent, showPreferences)
    return () =>
      window.removeEventListener(openAnalyticsConsentEvent, showPreferences)
  }, [])

  if (!open) return null

  const choose = (consent: 'granted' | 'denied') => {
    setAnalyticsConsent(consent)
    setOpen(false)
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
          <Button onClick={() => choose('denied')} size="sm" variant="quiet">
            Decline
          </Button>
          <Button onClick={() => choose('granted')} size="sm">
            Allow analytics
          </Button>
        </div>
      </Container>
    </aside>
  )
}
