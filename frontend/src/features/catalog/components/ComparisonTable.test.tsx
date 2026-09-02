import { render, screen, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import { ComparisonTable } from './ComparisonTable'
import type { ProductDetail } from '../schemas'

// MerchantAction fetches offers over the network. The row under test is about
// where the vendor exit sits and which source it is attributed to, not about
// how an offer is loaded, so the fetch is replaced by a marker carrying the
// attribution source — the one prop whose value decides whether a click is
// counted against the comparison surface or silently misfiled.
vi.mock('./MerchantAction', () => ({
  MerchantAction: ({ slug, source }: { slug: string; source: string }) => (
    <div data-source={source} data-testid={`merchant-action-${slug}`} />
  ),
}))

function product(overrides: Partial<ProductDetail> = {}): ProductDetail {
  return {
    id: 'a',
    name: 'Tool',
    slug: 'tool',
    brand: { name: 'Vendor', slug: 'vendor' },
    category: { name: 'CRM', slug: 'crm' },
    price: { amount_minor: 2000, currency: 'USD' },
    primary_image: null,
    key_specification: { label: 'Billing', value: 'Per month' },
    suitability: [],
    scores: {
      quality: 80,
      value: 80,
      durability: 80,
      beginner: 80,
      advanced: 80,
      apartment: 0,
      noise: 0,
      portability: 80,
    },
    is_demo: false,
    description: '',
    images: [],
    dimensions: { length_mm: 0, width_mm: 0, height_mm: 0 },
    weight_grams: 0,
    max_capacity_grams: null,
    material: '',
    warranty_months: 0,
    attributes: [],
    strengths: [],
    weaknesses: [],
    use_cases: [],
    alternatives: [],
    fact_revision_id: '00000000-0000-4000-8000-000000000000',
    score_revision_id: '00000000-0000-4000-8000-000000000000',
    evidence: [],
    ...overrides,
  }
}

const products = [
  product({ id: 'a', name: 'Alpha', slug: 'alpha' }),
  product({ id: 'b', name: 'Beta', slug: 'beta' }),
]

describe('ComparisonTable vendor row', () => {
  it('offers every compared product a vendor exit, not just the first', () => {
    render(
      <MemoryRouter>
        <ComparisonTable onRemove={vi.fn()} products={products} />
      </MemoryRouter>,
    )

    // Both columns, because a comparison that only lets you leave through one
    // of them has picked a winner without saying so.
    expect(screen.getByTestId('merchant-action-alpha')).toBeInTheDocument()
    expect(screen.getByTestId('merchant-action-beta')).toBeInTheDocument()
  })

  it('attributes the click to the comparison surface', () => {
    render(
      <MemoryRouter>
        <ComparisonTable onRemove={vi.fn()} products={products} />
      </MemoryRouter>,
    )

    // 'comparison' is one of the sources the API accepts; anything else is
    // rejected as invalid attribution and the reader lands nowhere.
    for (const slug of ['alpha', 'beta']) {
      expect(screen.getByTestId(`merchant-action-${slug}`)).toHaveAttribute(
        'data-source',
        'comparison',
      )
    }
  })

  it('keeps the exit inside the table footer, after the facts', () => {
    const { container } = render(
      <MemoryRouter>
        <ComparisonTable onRemove={vi.fn()} products={products} />
      </MemoryRouter>,
    )

    const footer = container.querySelector('tfoot')
    expect(footer).not.toBeNull()
    expect(
      within(footer as HTMLElement).getByText('Go to vendor'),
    ).toBeVisible()
    expect(
      within(footer as HTMLElement).getByTestId('merchant-action-alpha'),
    ).toBeInTheDocument()
  })

  it('says the buttons are affiliate links and that scoring ignores them', () => {
    render(
      <MemoryRouter>
        <ComparisonTable onRemove={vi.fn()} products={products} />
      </MemoryRouter>,
    )

    // The compact MerchantAction drops its own per-link disclosure, so this
    // sentence is the only one a reader gets. If it goes, the disclosure goes.
    expect(screen.getByText(/affiliate links/i)).toBeInTheDocument()
    expect(
      screen.getByText(/some of these tools\s+pay us and some do not/i),
    ).toBeInTheDocument()
  })
})

describe('ComparisonTable read-only mode', () => {
  it('hides the remove controls and keeps everything else', () => {
    render(
      <MemoryRouter>
        <ComparisonTable products={products} readOnly />
      </MemoryRouter>,
    )

    // An article's side-by-side shows a fixed set. A control that removes a
    // column from a comparison the reader did not build would have nothing
    // to remove from, so it is not drawn rather than drawn and ignored.
    expect(
      screen.queryByRole('button', { name: /Remove .* from comparison/ }),
    ).toBeNull()

    // The facts, the product links and the vendor exits are untouched.
    expect(screen.getByRole('link', { name: 'Alpha' })).toHaveAttribute(
      'href',
      '/products/alpha',
    )
    expect(screen.getByRole('link', { name: 'Beta' })).toHaveAttribute(
      'href',
      '/products/beta',
    )
    expect(screen.getByTestId('merchant-action-alpha')).toBeInTheDocument()
    expect(screen.getByTestId('merchant-action-beta')).toBeInTheDocument()
    expect(screen.getByText(/affiliate links/i)).toBeInTheDocument()
  })

  it('still offers the remove control when a handler is given and it is not read-only', () => {
    render(
      <MemoryRouter>
        <ComparisonTable onRemove={vi.fn()} products={products} />
      </MemoryRouter>,
    )
    expect(
      screen.getAllByRole('button', { name: /Remove .* from comparison/ }),
    ).toHaveLength(2)
  })
})
