import { describe, expect, it } from 'vitest'

import {
  rowIsIdentical,
  scoreBand,
  visibleComparisonRows,
  comparisonRows,
} from './comparisonRows'
import type { ProductDetail } from './schemas'

function product(overrides: Partial<ProductDetail> = {}): ProductDetail {
  return {
    id: overrides.id ?? 'a',
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

describe('scoreBand', () => {
  // A number out of a hundred looks precise and says nothing: nobody knows
  // whether 74 is good. The bands are coarse so a two-point gap does not read
  // as a difference when it is not one.
  it('gives coarse bands, so near-identical scores read the same', () => {
    expect(scoreBand(82)).toBe(scoreBand(80))
    expect(scoreBand(95)).toBe('Exceptional')
    expect(scoreBand(40)).toBe('Weak')
  })
})

describe('rowIsIdentical', () => {
  const priceRow = comparisonRows.find((r) => r.key === 'price')!

  it('spots a row every product answers the same', () => {
    expect(rowIsIdentical(priceRow, [product(), product({ id: 'b' })])).toBe(
      true,
    )
  })

  it('spots a row where they differ', () => {
    const dearer = product({
      id: 'b',
      price: { amount_minor: 9900, currency: 'USD' },
    })
    expect(rowIsIdentical(priceRow, [product(), dearer])).toBe(false)
  })
})

describe('visibleComparisonRows', () => {
  // The category row is only worth its space when the products are not all
  // from one category, which is the usual case.
  it('drops the category row when everything is in one category', () => {
    const rows = visibleComparisonRows([product(), product({ id: 'b' })])
    expect(rows.some((r) => r.key === 'category')).toBe(false)
  })

  it('keeps the category row when they are mixed', () => {
    const other = product({
      id: 'b',
      category: { name: 'Payments', slug: 'payments' },
    })
    const rows = visibleComparisonRows([product(), other])
    expect(rows.some((r) => r.key === 'category')).toBe(true)
  })

  // The billing basis is the row nobody else shows and the one that most often
  // decides it: a monthly rate against an annual one is not a comparison.
  it('always includes the billing basis', () => {
    const rows = visibleComparisonRows([product()])
    expect(rows.some((r) => r.key === 'billing')).toBe(true)
  })
})
