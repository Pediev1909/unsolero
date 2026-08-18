import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { AnalyticsConsentBanner } from './AnalyticsConsentBanner'
import { getAnalyticsConsent } from './consent'

afterEach(() => {
  vi.unstubAllGlobals()
  window.localStorage.clear()
})

describe('AnalyticsConsentBanner', () => {
  it('defaults to no tracking and stores an explicit decision', async () => {
    const user = userEvent.setup()
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            state: 'denied',
            policy_version: 'analytics-v1',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        ),
      ),
    )
    render(<AnalyticsConsentBanner />)

    expect(
      screen.getByRole('region', { name: 'Analytics preferences' }),
    ).toBeInTheDocument()
    expect(getAnalyticsConsent()).toBe('unknown')

    await user.click(screen.getByRole('button', { name: 'Decline' }))

    expect(getAnalyticsConsent()).toBe('denied')
    expect(
      screen.queryByRole('region', { name: 'Analytics preferences' }),
    ).not.toBeInTheDocument()
  })
})
