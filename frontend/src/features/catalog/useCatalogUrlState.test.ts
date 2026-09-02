import { act, renderHook } from '@testing-library/react'
import { createElement, type ReactNode } from 'react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import {
  emptyCatalogFilters,
  matchPriceBucket,
  priceBuckets,
  useCatalogUrlState,
} from './useCatalogUrlState'

function renderUrlState(route: string, fixed: { category?: string } = {}) {
  return renderHook(() => useCatalogUrlState(fixed), {
    wrapper: ({ children }: { children: ReactNode }) =>
      createElement(MemoryRouter, { initialEntries: [route] }, children),
  })
}

describe('matchPriceBucket', () => {
  it('names the bucket a pair of bounds spells, however the numbers are written', () => {
    expect(matchPriceBucket('', '')).toBe('any')
    expect(matchPriceBucket('0', '0')).toBe('no-fee')
    expect(matchPriceBucket('', '10')).toBe('to-10')
    expect(matchPriceBucket('10.00', '20')).toBe('10-20')
    expect(matchPriceBucket('50', '')).toBe('over-50')
  })

  // An empty bound is "unbounded", which is not the same as zero: "Up to $10"
  // includes the products with no monthly fee, "No monthly fee" is only them.
  it('treats hand-typed bounds that match no band as custom', () => {
    expect(matchPriceBucket('0', '')).toBe('custom')
    expect(matchPriceBucket('15', '40')).toBe('custom')
    expect(matchPriceBucket('', '0')).toBe('custom')
  })

  it('never offers a band that would read as free for a product taking a cut of sales', () => {
    expect(priceBuckets.map((bucket) => bucket.label)).not.toContain('Free')
  })
})

describe('useCatalogUrlState live-offer filter', () => {
  it('reads has_offer=true into the form values and the API query', () => {
    const { result } = renderUrlState(
      '/products?has_offer=true&minPrice=10&maxPrice=20',
    )
    expect(result.current.values.hasOffer).toBe(true)
    expect(result.current.query.hasOffer).toBe(true)
    expect(result.current.query.minPriceMinor).toBe(1000)
    expect(result.current.query.maxPriceMinor).toBe(2000)
  })

  // The API accepts only "true", so nothing else may be written; and "off" is
  // the parameter's absence rather than has_offer=false.
  it('writes the parameter only when the filter is on', () => {
    const { result } = renderUrlState('/products')
    expect(result.current.values.hasOffer).toBe(false)
    expect(result.current.query.hasOffer).toBeUndefined()

    act(() => {
      result.current.applyFilters({
        ...emptyCatalogFilters({}),
        hasOffer: true,
      })
    })
    expect(result.current.values.hasOffer).toBe(true)
    expect(result.current.query.hasOffer).toBe(true)

    act(() => {
      result.current.applyFilters(emptyCatalogFilters({}))
    })
    expect(result.current.values.hasOffer).toBe(false)
    expect(result.current.query.hasOffer).toBeUndefined()
  })

  it('ignores any spelling other than true', () => {
    const { result } = renderUrlState('/products?has_offer=1')
    expect(result.current.values.hasOffer).toBe(false)
    expect(result.current.query.hasOffer).toBeUndefined()
  })

  it('clears the filter with the others while keeping the fixed category', () => {
    const { result } = renderUrlState('/categories/crm?has_offer=true&q=zoho', {
      category: 'crm',
    })
    act(() => result.current.clearFilters())
    expect(result.current.values).toEqual(
      emptyCatalogFilters({ category: 'crm' }),
    )
    expect(result.current.query.category).toBe('crm')
  })
})
