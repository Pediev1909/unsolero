import { useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'

import type { CatalogQuery, CatalogSort } from './types'

export interface CatalogFilterValues {
  q: string
  category: string
  brand: string
  minPrice: string
  maxPrice: string
  /** Only products with a live vendor offer. Off unless the URL says so. */
  hasOffer: boolean
}

export interface PriceBucket {
  id: string
  label: string
  minPrice: string
  maxPrice: string
}

/**
 * Reference-price bands offered as one-tap chips, in USD per month. Each is
 * nothing more than a preset for the minimum and maximum fields, so a bucket
 * URL and a hand-typed URL with the same bounds are the same URL.
 *
 * The zero band says "No monthly fee" rather than "Free": six products in the
 * catalog charge nothing per month and take a percentage of each sale, and
 * PriceDisplay draws the same distinction for the same reason.
 */
export const priceBuckets: readonly PriceBucket[] = [
  { id: 'any', label: 'Any price', minPrice: '', maxPrice: '' },
  { id: 'no-fee', label: 'No monthly fee', minPrice: '0', maxPrice: '0' },
  { id: 'to-10', label: 'Up to $10', minPrice: '', maxPrice: '10' },
  { id: '10-20', label: '$10–20', minPrice: '10', maxPrice: '20' },
  { id: '20-50', label: '$20–50', minPrice: '20', maxPrice: '50' },
  { id: 'over-50', label: 'Over $50', minPrice: '50', maxPrice: '' },
]

/** The bucket a minimum/maximum pair spells, or `custom` when it matches none. */
export function matchPriceBucket(minPrice: string, maxPrice: string): string {
  const bucket = priceBuckets.find(
    (candidate) =>
      samePrice(candidate.minPrice, minPrice) &&
      samePrice(candidate.maxPrice, maxPrice),
  )
  return bucket?.id ?? 'custom'
}

// An empty bound means "unbounded" and never equals a number, not even zero.
function samePrice(left: string, right: string): boolean {
  if (left === '' || right === '') return left === right
  return Number(left) === Number(right)
}

const validSorts = new Set<CatalogSort>([
  'featured',
  'name_asc',
  'price_asc',
  'price_desc',
  'quality_desc',
  'value_desc',
])

export function useCatalogUrlState(fixed: {
  category?: string
  brand?: string
}) {
  const [searchParams, setSearchParams] = useSearchParams()

  const state = useMemo(() => {
    const rawSort = searchParams.get('sort') as CatalogSort | null
    const sort = rawSort && validSorts.has(rawSort) ? rawSort : 'featured'
    const page = positiveInteger(searchParams.get('page')) ?? 1
    const values: CatalogFilterValues = {
      q: searchParams.get('q') ?? '',
      category: fixed.category ?? searchParams.get('category') ?? '',
      brand: fixed.brand ?? searchParams.get('brand') ?? '',
      minPrice: searchParams.get('minPrice') ?? '',
      maxPrice: searchParams.get('maxPrice') ?? '',
      hasOffer: searchParams.get('has_offer') === 'true',
    }
    const query: CatalogQuery = {
      q: values.q || undefined,
      category: values.category || undefined,
      brand: values.brand || undefined,
      minPriceMinor: dollarsToMinor(values.minPrice),
      maxPriceMinor: dollarsToMinor(values.maxPrice),
      hasOffer: values.hasOffer || undefined,
      sort,
      page,
      pageSize: 12,
    }
    return { values, query, sort, page }
  }, [fixed.brand, fixed.category, searchParams])

  function applyFilters(values: CatalogFilterValues) {
    setSearchParams(buildParams(values, state.sort, 1, fixed))
  }

  function setSort(sort: CatalogSort) {
    setSearchParams(buildParams(state.values, sort, 1, fixed))
  }

  function setPage(page: number) {
    setSearchParams(buildParams(state.values, state.sort, page, fixed))
  }

  function clearFilters() {
    setSearchParams(
      buildParams(emptyCatalogFilters(fixed), 'featured', 1, fixed),
    )
  }

  return { ...state, applyFilters, setSort, setPage, clearFilters }
}

function buildParams(
  values: CatalogFilterValues,
  sort: CatalogSort,
  page: number,
  fixed: { category?: string; brand?: string },
) {
  const params = new URLSearchParams()
  if (values.q.trim()) params.set('q', values.q.trim())
  if (!fixed.category && values.category)
    params.set('category', values.category)
  if (!fixed.brand && values.brand) params.set('brand', values.brand)
  if (values.minPrice) params.set('minPrice', values.minPrice)
  if (values.maxPrice) params.set('maxPrice', values.maxPrice)
  // Same spelling as the API parameter, so a URL can be read against API.md.
  if (values.hasOffer) params.set('has_offer', 'true')
  if (sort !== 'featured') params.set('sort', sort)
  if (page > 1) params.set('page', String(page))
  return params
}

/** Every filter off, with the page's fixed category or brand kept. */
export function emptyCatalogFilters(fixed: {
  category?: string
  brand?: string
}): CatalogFilterValues {
  return {
    q: '',
    category: fixed.category ?? '',
    brand: fixed.brand ?? '',
    minPrice: '',
    maxPrice: '',
    hasOffer: false,
  }
}

function dollarsToMinor(value: string): number | undefined {
  if (!value) return undefined
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed >= 0
    ? Math.round(parsed * 100)
    : undefined
}

function positiveInteger(value: string | null): number | undefined {
  if (!value) return undefined
  const parsed = Number(value)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : undefined
}
