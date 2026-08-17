import { describe, expect, it } from 'vitest'

import { attributeLabel, attributeValue, prominentSuitability } from './model'
import type { ProductSummary } from './schemas'

const product: ProductSummary = {
  id: 'product-1',
  name: 'Demo Compact Trainer',
  slug: 'demo-compact-trainer',
  brand: { name: 'Demo Form', slug: 'demo-form' },
  category: { name: 'Strength', slug: 'strength' },
  price: { amount_minor: 19900, currency: 'USD' },
  primary_image: null,
  key_specification: { label: 'Capacity', value: '100 kg' },
  suitability: [
    { key: 'apartment', label: 'Apartment', score: 91 },
    { key: 'beginner', label: 'Beginner', score: 87 },
    { key: 'advanced', label: 'Advanced', score: 72 },
    { key: 'portable', label: 'Portable', score: 94 },
  ],
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

describe('catalog presentation model', () => {
  it('shows only the two strongest high-confidence suitability badges', () => {
    expect(prominentSuitability(product).map((item) => item.key)).toEqual([
      'portable',
      'apartment',
    ])
  })

  it('formats structured attributes without exposing storage keys', () => {
    expect(attributeLabel('folded_length')).toBe('Folded length')
    expect(
      attributeValue({
        key: 'folded_length',
        type: 'number',
        numeric_value: 640,
        unit: 'mm',
      }),
    ).toBe('640 mm')
  })
})
