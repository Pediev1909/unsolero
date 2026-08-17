import { apiRequest, ApiError } from '../../lib/api/client'
import {
  affiliateLinkSchema,
  analyticsReportSchema,
  affiliatePageSchema,
  brandSchema,
  categorySchema,
  dashboardSchema,
  eventPageSchema,
  merchantSchema,
  offerPageSchema,
  offerSchema,
  productPageSchema,
  productSchema,
  productGovernancePageSchema,
  productGovernanceSchema,
  recommendationDetailSchema,
  recommendationPageSchema,
  referencesSchema,
  userPageSchema,
} from './schemas'

export interface ProductInput {
  category_id: string
  brand_id: string
  name: string
  slug: string
  description: string
  price_minor: number
  currency: string
  length_mm: number
  width_mm: number
  height_mm: number
  weight_grams: number
  max_capacity_grams: number | null
  material: string
  warranty_months: number
  quality_score: number
  value_score: number
  durability_score: number
  beginner_score: number
  advanced_score: number
  apartment_score: number
  noise_score: number
  portability_score: number
}

export interface AffiliateInput {
  provider: string
  destination_url: string
  external_reference: string | null
  disclosure_label: string
  is_active: boolean
  priority: number
  program_id: string | null
  commission_type: 'unknown' | 'percentage' | 'fixed'
  commission_rate_bps: number | null
  commission_amount_minor: number | null
  commission_currency: string | null
}

export interface OfferInput {
  merchant_id: string
  product_id: string
  merchant_sku: string
  product_url: string
  price_minor: number
  shipping_minor: number
  currency: string
  availability: string
  condition: string
  is_active: boolean
  affiliate: AffiliateInput | null
}

const query = (page = 1, pageSize = 30) => `?page=${page}&page_size=${pageSize}`

export const adminApi = {
  dashboard: () =>
    apiRequest('/admin/dashboard', { method: 'GET' }, (value) =>
      dashboardSchema.parse(value),
    ),
  analytics: () =>
    apiRequest('/admin/analytics', { method: 'GET' }, (value) =>
      analyticsReportSchema.parse(value),
    ),
  references: () =>
    apiRequest('/admin/references', { method: 'GET' }, (value) =>
      referencesSchema.parse(value),
    ),
  products: (search = '', page = 1) =>
    apiRequest(
      `/admin/products${query(page)}&q=${encodeURIComponent(search)}`,
      { method: 'GET' },
      (value) => productPageSchema.parse(value),
    ),
  product: (id: string) =>
    apiRequest(`/admin/products/${id}`, { method: 'GET' }, (value) =>
      productSchema.parse(value),
    ),
  governedProducts: (page = 1) =>
    apiRequest(
      `/admin/evidence/products${query(page)}`,
      { method: 'GET' },
      (value) => productGovernancePageSchema.parse(value),
    ),
  productGovernance: (id: string) =>
    apiRequest(`/admin/evidence/products/${id}`, { method: 'GET' }, (value) =>
      productGovernanceSchema.parse(value),
    ),
  createProduct: (input: ProductInput) =>
    apiRequest('/admin/products', { method: 'POST', body: input }, (value) =>
      productSchema.parse(value),
    ),
  updateProduct: (id: string, input: ProductInput) =>
    apiRequest(
      `/admin/products/${id}`,
      { method: 'PATCH', body: input },
      (value) => productSchema.parse(value),
    ),
  setProductStatus: (id: string, status: string) =>
    apiRequest(
      `/admin/products/${id}/status`,
      { method: 'PUT', body: { status } },
      () => undefined,
    ),
  addImage: (
    productID: string,
    input: {
      url: string
      alt_text: string
      sort_order: number
      is_primary: boolean
    },
  ) =>
    apiRequest(
      `/admin/products/${productID}/images`,
      { method: 'POST', body: input },
      (value) => zImage(value),
    ),
  uploadImage: async (
    productID: string,
    input: {
      file: File
      alt_text: string
      sort_order: number
      is_primary: boolean
    },
  ) => {
    const form = new FormData()
    form.set('file', input.file)
    form.set('alt_text', input.alt_text)
    form.set('sort_order', String(input.sort_order))
    form.set('is_primary', String(input.is_primary))
    const response = await fetch(`/api/admin/products/${productID}/images`, {
      method: 'POST',
      body: form,
      credentials: 'include',
    })
    if (!response.ok) {
      throw new ApiError(
        response.status,
        'image_upload_failed',
        'The image could not be uploaded.',
      )
    }
    return zImage(await response.json())
  },
  deleteImage: (productID: string, imageID: string) =>
    apiRequest(
      `/admin/products/${productID}/images/${imageID}`,
      { method: 'DELETE' },
      () => undefined,
    ),
  upsertAttribute: (
    productID: string,
    key: string,
    input: Record<string, unknown>,
  ) =>
    apiRequest(
      `/admin/products/${productID}/attributes/${encodeURIComponent(key)}`,
      { method: 'PUT', body: input },
      (value) => value,
    ),
  deleteAttribute: (productID: string, key: string) =>
    apiRequest(
      `/admin/products/${productID}/attributes/${encodeURIComponent(key)}`,
      { method: 'DELETE' },
      () => undefined,
    ),
  categories: () =>
    apiRequest('/admin/categories', { method: 'GET' }, (value) =>
      categorySchema.array().parse(value),
    ),
  brands: () =>
    apiRequest('/admin/brands', { method: 'GET' }, (value) =>
      brandSchema.array().parse(value),
    ),
  merchants: () =>
    apiRequest('/admin/merchants', { method: 'GET' }, (value) =>
      merchantSchema.array().parse(value),
    ),
  offers: (page = 1) =>
    apiRequest(`/admin/offers${query(page)}`, { method: 'GET' }, (value) =>
      offerPageSchema.parse(value),
    ),
  createOffer: (input: OfferInput) =>
    apiRequest('/admin/offers', { method: 'POST', body: input }, (value) =>
      offerSchema.parse(value),
    ),
  updateOffer: (id: string, input: OfferInput) =>
    apiRequest(
      `/admin/offers/${id}`,
      { method: 'PATCH', body: input },
      (value) => offerSchema.parse(value),
    ),
  affiliateLinks: (page = 1) =>
    apiRequest(
      `/admin/affiliate-links${query(page)}`,
      { method: 'GET' },
      (value) => affiliatePageSchema.parse(value),
    ),
  updateAffiliate: (id: string, input: AffiliateInput) =>
    apiRequest(
      `/admin/affiliate-links/${id}`,
      { method: 'PATCH', body: input },
      (value) => affiliateLinkSchema.parse(value),
    ),
  recommendations: (page = 1) =>
    apiRequest(
      `/admin/recommendations${query(page)}`,
      { method: 'GET' },
      (value) => recommendationPageSchema.parse(value),
    ),
  recommendation: (id: string) =>
    apiRequest(`/admin/recommendations/${id}`, { method: 'GET' }, (value) =>
      recommendationDetailSchema.parse(value),
    ),
  users: (page = 1) =>
    apiRequest(`/admin/users${query(page)}`, { method: 'GET' }, (value) =>
      userPageSchema.parse(value),
    ),
  events: (name = '', page = 1) =>
    apiRequest(
      `/admin/events${query(page)}&name=${encodeURIComponent(name)}`,
      { method: 'GET' },
      (value) => eventPageSchema.parse(value),
    ),
}

function zImage(value: unknown) {
  return productSchema.shape.images.element.parse(value)
}
