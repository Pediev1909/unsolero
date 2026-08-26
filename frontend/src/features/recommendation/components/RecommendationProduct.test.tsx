import { screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { renderWithProviders } from '../../../test/renderWithProviders'
import type { RecommendationResult } from '../schemas'
import { RecommendationProduct } from './RecommendationProduct'

vi.mock('../../catalog/components/MerchantAction', () => ({
  MerchantAction: () => <div data-testid="merchant-action" />,
}))

const breakdown = {
  goal_match: 96,
  budget_match: 90,
  space_match: 100,
  experience_match: 88,
  preference_match: 84,
  quality: 91,
  value: 87,
  durability: 90,
  compatibility: 86,
  portability: 75,
  noise: 100,
}

const item: RecommendationResult['recommended_products'][number] = {
  rank: 1,
  quantity: 1,
  score: 87,
  breakdown,
  reasons: [
    {
      code: 'budget',
      message: 'Within your budget',
      dimension: 'budget',
      score: 90,
    },
  ],
  product: {
    id: 'shopify-basic',
    name: 'Shopify Basic',
    slug: 'shopify-basic',
    brand: { name: 'Shopify', slug: 'shopify' },
    category: { name: 'Commerce', slug: 'commerce' },
    price: { amount_minor: 2900, currency: 'USD' },
    primary_image: null,
    key_specification: { label: 'Billing', value: 'Per month' },
    suitability: [],
    scores: {
      quality: 91,
      value: 87,
      durability: 90,
      beginner: 88,
      advanced: 84,
      apartment: 100,
      noise: 100,
      portability: 100,
    },
    is_demo: false,
  },
}

describe('RecommendationProduct', () => {
  it('identifies the recommendation with the self-hosted vendor mark', () => {
    const { container } = renderWithProviders(
      <RecommendationProduct
        compared={false}
        item={item}
        onCompare={vi.fn()}
        onSave={vi.fn()}
        savePending={false}
        saved={false}
      />,
    )

    expect(
      screen.getByRole('heading', { name: 'Shopify Basic' }),
    ).toBeInTheDocument()
    expect(container.querySelector('img')?.getAttribute('src')).toBe(
      '/images/brands/shopify.png',
    )
    expect(screen.getByText('87/100 match')).toBeInTheDocument()
    expect(screen.getByTestId('merchant-action')).toBeInTheDocument()
  })
})
