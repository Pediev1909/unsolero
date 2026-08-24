import { screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { renderWithProviders } from '../test/renderWithProviders'
import { HomePage } from './HomePage'

// The featured cards used to be a hardcoded array, so this file could assert a
// product name directly. They come from the catalog now, and the point worth
// pinning is no longer which product appears — it is that whatever the catalog
// returns is rendered as a link to that product's page.
const featuredProduct = {
  id: 'a7f1c2d4-5b6e-4a8c-9d0e-1f2a3b4c5d6e',
  name: 'ClickUp Unlimited',
  slug: 'clickup-unlimited',
  brand: { id: 'b1', name: 'ClickUp', slug: 'clickup' },
  category: {
    id: 'c1',
    name: 'Project management',
    slug: 'project-management',
  },
  price: { amount_minor: 1000, currency: 'USD' },
  primary_image: null,
  key_specification: { label: 'Plan', value: 'Unlimited' },
  suitability: [],
  scores: {},
  is_demo: false,
}

vi.mock('../features/catalog/queries', async (importOriginal) => {
  const actual =
    await importOriginal<typeof import('../features/catalog/queries')>()
  return {
    ...actual,
    useProducts: () => ({
      data: {
        products: [featuredProduct],
        page: 1,
        page_size: 4,
        total: 1,
        total_pages: 1,
      },
      isPending: false,
      isError: false,
      isSuccess: true,
    }),
  }
})

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
    expect(screen.getByRole('link', { name: /^CRM/i })).toHaveAttribute(
      'href',
      '/categories/crm',
    )
    expect(
      screen.getByRole('link', {
        name: /view details for clickup unlimited/i,
      }),
    ).toHaveAttribute('href', '/products/clickup-unlimited')
  })

  it('labels illustrative catalog content and exposes comparison context', () => {
    renderWithProviders(<HomePage />)

    expect(screen.getByText(/read from each vendor/i)).toBeInTheDocument()
    expect(
      screen.getByRole('table', {
        name: /comparison of three business tools/i,
      }),
    ).toBeInTheDocument()
    expect(screen.queryByText(/testimonial/i)).not.toBeInTheDocument()
  })
})
