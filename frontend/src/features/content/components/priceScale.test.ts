import { renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { usePriceRows } from './priceRows'
import type { ProductSummary } from '../../catalog/schemas'

function product(name: string, amountMinor: number, currency = 'USD') {
  return {
    id: `id-${name}`,
    name,
    slug: name.toLowerCase(),
    brand: { id: 'b', name, slug: name.toLowerCase() },
    category: { id: 'c', name: 'Analytics', slug: 'analytics' },
    price: { amount_minor: amountMinor, currency },
    primary_image: null,
    key_specification: { label: '', value: '' },
    suitability: [],
    scores: {},
    is_demo: false,
  } as unknown as ProductSummary
}

const rowsFor = (products: ProductSummary[]) =>
  renderHook(() => usePriceRows(products)).result.current

describe('usePriceRows', () => {
  it('orders cheapest first and marks it', () => {
    const scale = rowsFor([
      product('Umami', 2000),
      product('Fathom', 1500),
      product('Simple', 2000),
    ])
    expect(scale?.rows.map((row) => row.name)).toEqual([
      'Fathom',
      'Umami',
      'Simple',
    ])
    expect(scale?.rows[0]?.cheapest).toBe(true)
    expect(scale?.rows[1]?.cheapest).toBe(false)
  })

  it('scales every bar against the dearest', () => {
    const scale = rowsFor([product('A', 1000), product('B', 2000)])
    expect(scale?.rows[0]?.width).toBe(50)
    expect(scale?.rows[1]?.width).toBe(100)
  })

  // A scale of identical prices shows nothing, and a single product is not a
  // comparison. Both keep the illustration rather than render a flat chart.
  it('declines when every price is the same', () => {
    expect(rowsFor([product('A', 2000), product('B', 2000)])).toBeNull()
  })

  it('declines a single product', () => {
    expect(rowsFor([product('A', 1500)])).toBeNull()
  })

  it('declines when the prices are in different currencies', () => {
    expect(rowsFor([product('A', 1500), product('B', 2000, 'EUR')])).toBeNull()
  })

  // A free tier is a real price and the row has to stay visible, or it reads
  // as missing data rather than as zero.
  it('keeps a free product visible and labels it', () => {
    const scale = rowsFor([product('Free', 0), product('Paid', 2000)])
    expect(scale?.rows[0]?.free).toBe(true)
    expect(scale?.rows[0]?.width).toBeGreaterThan(0)
    expect(scale?.rows[0]?.cheapest).toBe(true)
  })

  it('ignores products the catalog has no price for', () => {
    const priced = [product('A', 1000), product('B', 2000)]
    const unpriced = { ...product('C', 0), price: undefined }
    const scale = rowsFor([...priced, unpriced as unknown as ProductSummary])
    expect(scale?.rows).toHaveLength(2)
  })
})
