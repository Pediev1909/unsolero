import { render, screen } from '@testing-library/react'
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
