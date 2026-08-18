import { MemoryRouter } from 'react-router-dom'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { SiteHeader } from './SiteHeader'

describe('SiteHeader', () => {
  it('exposes skip navigation and marks the current section', () => {
    render(
      <MemoryRouter initialEntries={['/products/demo-product']}>
        <SiteHeader position="static" />
        <main id="main-content">Product</main>
      </MemoryRouter>,
    )

    expect(
      screen.getByRole('link', { name: 'Skip to main content' }),
    ).toHaveAttribute('href', '#main-content')
    expect(
      screen.getByRole('link', { name: 'UNSOLERO home' }),
    ).toHaveTextContent('UNSOLERO')
    expect(
      within(
        screen.getByRole('navigation', { name: 'Primary navigation' }),
      ).getByRole('link', { name: 'Software' }),
    ).toHaveAttribute('aria-current', 'page')
  })

  it('opens and closes the accessible mobile navigation', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <SiteHeader position="static" />
      </MemoryRouter>,
    )

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
