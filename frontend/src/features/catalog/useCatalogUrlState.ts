import { useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'

import type { CatalogQuery, CatalogSort } from './types'

export interface CatalogFilterValues {
  q: string
  category: string
  brand: string
  minPrice: string
  maxPrice: string
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
    }
    const query: CatalogQuery = {
      q: values.q || undefined,
      category: values.category || undefined,
      brand: values.brand || undefined,
      minPriceMinor: dollarsToMinor(values.minPrice),
      maxPriceMinor: dollarsToMinor(values.maxPrice),
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
    setSearchParams(buildParams(emptyValues(fixed), 'featured', 1, fixed))
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
  if (sort !== 'featured') params.set('sort', sort)
  if (page > 1) params.set('page', String(page))
  return params
}

function emptyValues(fixed: {
  category?: string
  brand?: string
}): CatalogFilterValues {
  return {
    q: '',
    category: fixed.category ?? '',
    brand: fixed.brand ?? '',
    minPrice: '',
    maxPrice: '',
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
