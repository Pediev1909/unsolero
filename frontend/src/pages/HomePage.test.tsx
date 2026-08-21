import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { renderWithProviders } from '../test/renderWithProviders'
import { HomePage } from './HomePage'

describe('HomePage', () => {
  it('communicates the value proposition and routes both primary actions', () => {
    renderWithProviders(<HomePage />)

    expect(
      screen.getByRole('heading', {
        level: 1,
        name: /build the right software stack/i,
      }),
    ).toBeInTheDocument()
    expect(
      screen.getByText(/^commission never changes objective ranking/i),
    ).toBeInTheDocument()
    expect(
      screen.getAllByRole('link', { name: /build my setup/i })[0],
    ).toHaveAttribute('href', '/build')
    expect(
      screen.getAllByRole('link', { name: /explore categories/i })[0],
    ).toHaveAttribute('href', '/#categories')
    // "How it works" used to be a hash anchor that scrolled this page. A nav
    // item that scrolls where every other one navigates is a small trap, and
    // an anchor cannot be linked to from elsewhere or ranked on its own.
    expect(
      screen.getAllByRole('link', { name: 'How it works' })[0],
    ).toHaveAttribute('href', '/how-it-works')
    expect(document.getElementById('method')).toBeInTheDocument()
    expect(document.getElementById('categories')).toBeInTheDocument()
    expect(document.getElementById('trust')).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: /^CRM/i }),
    ).toHaveAttribute('href', '/categories/crm')
    expect(
      screen.getByRole('link', {
        name: /view details for clickup unlimited/i,
      }),
    ).toHaveAttribute('href', '/products/clickup-unlimited')
  })

  it('labels illustrative catalog content and exposes comparison context', () => {
    renderWithProviders(<HomePage />)

    expect(
      screen.getByText(/read from each vendor/i),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('table', {
        name: /comparison of three business tools/i,
      }),
    ).toBeInTheDocument()
    expect(screen.queryByText(/testimonial/i)).not.toBeInTheDocument()
  })
})
