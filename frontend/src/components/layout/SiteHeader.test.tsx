import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { renderWithProviders } from '../../test/renderWithProviders'
import { SiteHeader } from './SiteHeader'

describe('SiteHeader', () => {
  it('exposes skip navigation and marks the current section', () => {
    renderWithProviders(
      <>
        <SiteHeader position="static" />
        <main id="main-content">Product</main>
      </>,
      { route: '/products/demo-product' },
    )

    expect(
      screen.getByRole('link', { name: 'Skip to main content' }),
    ).toHaveAttribute('href', '#main-content')
    expect(
      screen.getByRole('link', { name: 'UNSOLERO home' }),
    ).toHaveTextContent('UNSOLERO')
    // Browse has no page of its own, so it borrows the state of the catalog
    // pages it contains. A product page is inside Browse.
    expect(
      within(
        screen.getByRole('navigation', { name: 'Primary navigation' }),
      ).getByRole('button', { name: /Browse/ }),
    ).toHaveAttribute('aria-current', 'true')
  })

  it('opens and closes the accessible mobile navigation', async () => {
    const user = userEvent.setup()
    renderWithProviders(<SiteHeader position="static" />)

    const trigger = screen.getByRole('button', { name: 'Open navigation' })
    expect(trigger).toHaveAttribute('aria-expanded', 'false')

    await user.click(trigger)

    const drawer = screen.getByRole('dialog', { name: 'Navigate' })
    expect(drawer).toHaveAttribute('open')
    expect(trigger).toHaveAttribute('aria-expanded', 'true')
    expect(
      screen.getByRole('navigation', { name: 'Mobile navigation' }),
    ).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Close drawer' }))
    expect(drawer).not.toHaveAttribute('open')
    expect(trigger).toHaveAttribute('aria-expanded', 'false')
  })
})
