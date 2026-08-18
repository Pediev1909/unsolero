import { apiRequest } from '../../lib/api/client'
import {
  brandSchema,
  categorySchema,
  offerSchema,
  productDetailSchema,
  productPageSchema,
  productSelectionSchema,
  wishlistSchema,
} from './schemas'
import type { CatalogQuery } from './types'
import { z } from 'zod'

function queryString(query: CatalogQuery): string {
  const values = new URLSearchParams()
  if (query.q) values.set('q', query.q)
  if (query.category) values.set('category', query.category)
  if (query.brand) values.set('brand', query.brand)
  if (query.minPriceMinor !== undefined)
    values.set('min_price_minor', String(query.minPriceMinor))
  if (query.maxPriceMinor !== undefined)
    values.set('max_price_minor', String(query.maxPriceMinor))
  if (query.sort) values.set('sort', query.sort)
  if (query.page) values.set('page', String(query.page))
  if (query.pageSize) values.set('page_size', String(query.pageSize))
  if (query.ids?.length) values.set('ids', query.ids.join(','))
  const encoded = values.toString()
  return encoded ? `?${encoded}` : ''
}

export function getComparison() {
  return apiRequest('/account/comparison', { method: 'GET' }, (value) =>
    productSelectionSchema.parse(value),
  )
}

export function replaceComparison(productIDs: string[]) {
  return apiRequest(
    '/account/comparison',
    { method: 'PUT', body: { product_ids: productIDs } },
    (value) => productSelectionSchema.parse(value),
  )
}

export function getProducts(query: CatalogQuery) {
  return apiRequest(
    `/catalog/products${queryString(query)}`,
    { method: 'GET' },
    (value) => productPageSchema.parse(value),
  )
}

export function getProduct(slug: string) {
  return apiRequest(`/catalog/products/${slug}`, { method: 'GET' }, (value) =>
    productDetailSchema.parse(value),
  )
}

export function getCategories() {
  return apiRequest('/catalog/categories', { method: 'GET' }, (value) =>
    z.array(categorySchema).parse(value),
  )
}

export function getCategory(slug: string) {
  return apiRequest(`/catalog/categories/${slug}`, { method: 'GET' }, (value) =>
    categorySchema.parse(value),
  )
}

export function getBrands() {
  return apiRequest('/catalog/brands', { method: 'GET' }, (value) =>
    z.array(brandSchema).parse(value),
  )
}

export function getBrand(slug: string) {
  return apiRequest(`/catalog/brands/${slug}`, { method: 'GET' }, (value) =>
    brandSchema.parse(value),
  )
}

export function getOffers(slug: string) {
  return apiRequest(
    `/catalog/products/${slug}/offers`,
    { method: 'GET' },
    (value) => z.array(offerSchema).parse(value),
  )
}

export async function getWishlist() {
  const productIDs: string[] = []
  let page = 1
  while (true) {
    const result = await apiRequest(
      `/account/wishlist?page=${page}&page_size=100`,
      { method: 'GET' },
      (value) => wishlistSchema.parse(value),
    )
    productIDs.push(...result.product_ids)
    if (page >= result.total_pages || page >= 10_000) break
    page += 1
  }
  return { product_ids: productIDs }
}

export function saveWishlistProduct(productID: string) {
  return apiRequest(
    `/account/wishlist/${productID}`,
    { method: 'PUT' },
    () => undefined,
  )
}

export function deleteWishlistProduct(productID: string) {
  return apiRequest(
    `/account/wishlist/${productID}`,
    { method: 'DELETE' },
    () => undefined,
  )
}
