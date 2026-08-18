import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import { HomePage } from './HomePage'

describe('HomePage', () => {
  it('communicates the value proposition and routes both primary actions', () => {
    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>,
    )

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
    expect(
      screen.getAllByRole('link', { name: 'How it works' })[0],
    ).toHaveAttribute('href', '/#method')
    expect(document.getElementById('method')).toBeInTheDocument()
    expect(document.getElementById('categories')).toBeInTheDocument()
    expect(document.getElementById('trust')).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: /^CRM/i }),
    ).toHaveAttribute('href', '/categories/crm')
    expect(
      screen.getByRole('link', {
        name: /view details for demo northwind crm/i,
      }),
    ).toHaveAttribute('href', '/products/saas-northwind-crm')
  })

  it('labels illustrative catalog content and exposes comparison context', () => {
    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>,
    )

    expect(
      screen.getByText(/fictional demo products shown/i),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('table', {
        name: /comparison of three fictional demo business tools/i,
      }),
    ).toBeInTheDocument()
    expect(screen.queryByText(/testimonial/i)).not.toBeInTheDocument()
  })
})
