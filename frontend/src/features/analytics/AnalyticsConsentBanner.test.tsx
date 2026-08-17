import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { AnalyticsConsentBanner } from './AnalyticsConsentBanner'
import { getAnalyticsConsent } from './consent'

afterEach(() => window.localStorage.clear())

describe('AnalyticsConsentBanner', () => {
  it('defaults to no tracking and stores an explicit decision', async () => {
    const user = userEvent.setup()
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
