import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import type { ProductSummary } from '../schemas'
import { CatalogProductCard } from './CatalogProductCard'

const product: ProductSummary = {
  id: 'product-1',
  name: 'Demo Compact Trainer',
  slug: 'demo-compact-trainer',
  brand: { name: 'Demo Form', slug: 'demo-form' },
  category: { name: 'Strength', slug: 'strength' },
  price: { amount_minor: 19900, currency: 'USD' },
  primary_image: {
    url: '/images/demo-power-rack.webp',
    alt_text: 'Illustrative studio image for a fictional demo product.',
    is_primary: true,
    width_px: 1000,
    height_px: 750,
  },
  key_specification: { label: 'Capacity', value: '100 kg' },
  suitability: [{ key: 'apartment', label: 'Apartment', score: 91 }],
  scores: {
    quality: 82,
    value: 88,
    durability: 80,
    beginner: 87,
    advanced: 72,
    apartment: 91,
    noise: 86,
    portability: 94,
  },
  is_demo: true,
}

describe('CatalogProductCard', () => {
  it('renders facts and actions without inventing review ratings', () => {
    render(
      <MemoryRouter>
        <CatalogProductCard
          compared={false}
          onCompare={vi.fn()}
          onSave={vi.fn()}
          product={product}
          saved={false}
        />
      </MemoryRouter>,
    )

    expect(
      screen.getByRole('heading', { name: product.name }),
    ).toBeInTheDocument()
    expect(screen.getByText('100 kg')).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: `Save ${product.name}` }),
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Compare' })).toBeInTheDocument()
    expect(screen.queryByText(/rating|reviews?/i)).not.toBeInTheDocument()
  })
})

describe('CatalogProductCard billing basis', () => {
  const software = {
    ...product,
    key_specification: { label: 'Billing', value: 'Per month' },
  }

  // A vendor that sells only yearly contracts quotes a per-month figure that
  // looks exactly like a monthly rate. The card is where most readers meet
  // the price, so it is where the difference has to be said.
  it('says billed yearly for a vendor that sells only annual contracts', () => {
    render(
      <MemoryRouter>
        <CatalogProductCard
          compared={false}
          onCompare={vi.fn()}
          onSave={vi.fn()}
          product={{
            ...software,
            billing: {
              period: 'annual',
              unit: 'per_user',
              unit_note: null,
              annual_price_minor: null,
            },
          }}
          saved={false}
        />
      </MemoryRouter>,
    )
    expect(screen.getByText('Per user, billed yearly')).toBeInTheDocument()
    expect(screen.queryByText('Per month')).toBeNull()
  })

  // A response cached before the field existed carries only the server's
  // phrase, and the card prints that rather than nothing.
  it('prints the server phrase when the response has no billing object', () => {
    render(
      <MemoryRouter>
        <CatalogProductCard
          compared={false}
          onCompare={vi.fn()}
          onSave={vi.fn()}
          product={software}
          saved={false}
        />
      </MemoryRouter>,
    )
    expect(screen.getByText('Per month')).toBeInTheDocument()
  })
})

describe('CatalogProductCard save toggle', () => {
  it('asks the caller to save the product and reports the unsaved state', async () => {
    const onSave = vi.fn()
    render(
      <MemoryRouter>
        <CatalogProductCard
          compared={false}
          onCompare={vi.fn()}
          onSave={onSave}
          product={product}
          saved={false}
        />
      </MemoryRouter>,
    )
    const toggle = screen.getByRole('button', { name: `Save ${product.name}` })
    expect(toggle).toHaveAttribute('aria-pressed', 'false')
    expect(toggle).toHaveTextContent('Save')
    await userEvent.click(toggle)
    expect(onSave).toHaveBeenCalledWith(product)
  })

  it('reads as pressed and says Saved once the product is on the list', () => {
    render(
      <MemoryRouter>
        <CatalogProductCard
          compared={false}
          onCompare={vi.fn()}
          onSave={vi.fn()}
          product={product}
          saved
        />
      </MemoryRouter>,
    )
    const toggle = screen.getByRole('button', {
      name: `Saved ${product.name}`,
    })
    expect(toggle).toHaveAttribute('aria-pressed', 'true')
    expect(toggle).toHaveTextContent('Saved')
  })
})

describe('CatalogProductCard vendor exit', () => {
  const purchasable = {
    ...product,
    purchase_path: '/api/affiliate/click/offer-1',
    merchant_name: 'ActiveCampaign',
    disclosure_label: 'Affiliate link',
  }

  it('offers the vendor when the API says there is a live offer', () => {
    render(
      <MemoryRouter>
        <CatalogProductCard
          compared={false}
          onCompare={vi.fn()}
          onSave={vi.fn()}
          product={purchasable}
          saved={false}
        />
      </MemoryRouter>,
    )
    const link = screen.getByRole('link', { name: /View at ActiveCampaign/ })
    expect(link.getAttribute('href')).toContain('/api/affiliate/click/offer-1')
    expect(link).toHaveAttribute('rel', 'nofollow noopener sponsored')
    expect(screen.getByText(/Commission never/)).toBeInTheDocument()
  })

  // A card with no live offer shows nothing rather than a disabled control:
  // a greyed-out button reads as something being withheld.
  it('shows nothing when there is no live offer', () => {
    render(
      <MemoryRouter>
        <CatalogProductCard
          compared={false}
          onCompare={vi.fn()}
          onSave={vi.fn()}
          product={product}
          saved={false}
        />
      </MemoryRouter>,
    )
    expect(screen.queryByRole('link', { name: /View at/ })).toBeNull()
    expect(screen.queryByText(/Commission never/)).toBeNull()
    // The card still does its actual job.
    expect(
      screen.getByRole('link', { name: /View details/ }),
    ).toBeInTheDocument()
  })

  it('attributes the click to the surface it was drawn on', () => {
    render(
      <MemoryRouter>
        <CatalogProductCard
          compared={false}
          onCompare={vi.fn()}
          onSave={vi.fn()}
          product={purchasable}
          saved={false}
          source="comparison"
        />
      </MemoryRouter>,
    )
    expect(
      screen
        .getByRole('link', { name: /View at ActiveCampaign/ })
        .getAttribute('href'),
    ).toContain('source=comparison')
  })
})
