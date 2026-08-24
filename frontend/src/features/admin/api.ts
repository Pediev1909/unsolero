import { apiRequest, ApiError } from '../../lib/api/client'
import {
  affiliateLinkSchema,
  analyticsReportSchema,
  affiliatePageSchema,
  brandSchema,
  categorySchema,
  commerceImportFailuresSchema,
  commerceImportSchema,
  commerceImportsSchema,
  conversionImportSchema,
  conversionImportsSchema,
  commerceProviderSchema,
  commerceProvidersSchema,
  dashboardSchema,
  eventPageSchema,
  merchantSchema,
  offerPageSchema,
  offerSchema,
  productPageSchema,
  productSchema,
  productGovernancePageSchema,
  productGovernanceSchema,
  evidenceObservationListSchema,
  evidenceSourceListSchema,
  recommendationPolicyListSchema,
  monetizationReportSchema,
  reconciliationsSchema,
  recommendationDetailSchema,
  recommendationPageSchema,
  referencesSchema,
  userPageSchema,
  verifiedConversionsSchema,
} from './schemas'

export interface EvidenceSourceInput {
  source_type: string
  title: string
  publisher: string
  source_url: string | null
  is_fictional: boolean
}

export interface EvidenceObservationInput {
  source_id: string
  product_id: string
  observed_at: string
  expires_at: string | null
  confidence: number
  notes: string
}

export interface EvidenceRevisionInput {
  product: ProductInput
  fact_links: {
    fact_key: string
    observation_id: string
    classification: string
  }[]
  score_rationales: {
    score_key: string
    rationale: string
    observation_id: string
  }[]
}

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

export interface CommerceProviderInput {
  merchant_id: string
  provider_key: string
  adapter_key: string
  external_merchant_id: string
  credential_reference: string | null
  schedule_interval_minutes: number
  freshness_ttl_minutes: number
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
  recommendationPolicies: () =>
    apiRequest('/admin/recommendation-policies', { method: 'GET' }, (value) =>
      recommendationPolicyListSchema.parse(value),
    ),
  transitionRecommendationPolicy: (
    version: string,
    action: 'submit' | 'approve' | 'reject' | 'activate' | 'deactivate',
    note: string,
  ) =>
    apiRequest(
      `/admin/recommendation-policies/${version}/${action}`,
      { method: 'POST', body: { note } },
      () => undefined,
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
  evidenceSources: () =>
    apiRequest('/admin/evidence/sources', { method: 'GET' }, (value) =>
      evidenceSourceListSchema.parse(value),
    ),
  evidenceObservations: (productID: string) =>
    apiRequest(
      `/admin/evidence/products/${productID}/observations`,
      { method: 'GET' },
      (value) => evidenceObservationListSchema.parse(value),
    ),
  createEvidenceSource: (input: EvidenceSourceInput) =>
    apiRequest(
      '/admin/evidence/sources',
      { method: 'POST', body: input },
      () => undefined,
    ),
  reviewEvidenceSource: (id: string, status: string, note: string) =>
    apiRequest(
      `/admin/evidence/sources/${id}/review`,
      { method: 'PUT', body: { status, note } },
      () => undefined,
    ),
  createEvidenceObservation: (input: EvidenceObservationInput) =>
    apiRequest(
      '/admin/evidence/observations',
      { method: 'POST', body: input },
      () => undefined,
    ),
  createEvidenceRevision: (productID: string, input: EvidenceRevisionInput) =>
    apiRequest(
      `/admin/evidence/products/${productID}/revisions`,
      { method: 'POST', body: input },
      () => undefined,
    ),
  transitionEvidenceRevision: (
    revisionID: string,
    action: 'submit' | 'approve' | 'reject' | 'publish',
    note: string,
  ) =>
    apiRequest(
      `/admin/evidence/revisions/${revisionID}/${action}`,
      { method: 'POST', body: { note } },
      () => undefined,
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
  commerceProviders: () =>
    apiRequest('/admin/commerce/providers', { method: 'GET' }, (value) =>
      commerceProvidersSchema.parse(value),
    ),
  createCommerceProvider: (input: CommerceProviderInput) =>
    apiRequest(
      '/admin/commerce/providers',
      { method: 'POST', body: input },
      (value) => commerceProviderSchema.parse(value),
    ),
  setCommerceProviderLifecycle: (id: string, status: string) =>
    apiRequest(
      `/admin/commerce/providers/${id}/lifecycle`,
      { method: 'PUT', body: { status } },
      (value) => commerceProviderSchema.parse(value),
    ),
  commerceImports: (page = 1) =>
    apiRequest(
      `/admin/commerce/imports${query(page)}`,
      { method: 'GET' },
      (value) => commerceImportsSchema.parse(value),
    ),
  triggerCommerceImport: (providerConfigurationID: string) =>
    apiRequest(
      '/admin/commerce/imports',
      {
        method: 'POST',
        headers: { 'Idempotency-Key': crypto.randomUUID() },
        body: { provider_configuration_id: providerConfigurationID },
      },
      (value) => commerceImportSchema.parse(value),
    ),
  retryCommerceImport: (id: string) =>
    apiRequest(
      `/admin/commerce/imports/${id}/retry`,
      { method: 'POST', headers: { 'Idempotency-Key': crypto.randomUUID() } },
      (value) => commerceImportSchema.parse(value),
    ),
  commerceImportFailures: (id: string, page = 1) =>
    apiRequest(
      `/admin/commerce/imports/${id}/failures${query(page)}`,
      { method: 'GET' },
      (value) => commerceImportFailuresSchema.parse(value),
    ),
  setConversionProvider: (id: string, enabled: boolean) =>
    apiRequest(
      `/admin/commerce/providers/${id}/conversions`,
      { method: 'PUT', body: { enabled } },
      (value) => commerceProviderSchema.parse(value),
    ),
  verifiedConversions: (page = 1) =>
    apiRequest(
      `/admin/commerce/conversions${query(page)}`,
      { method: 'GET' },
      (value) => verifiedConversionsSchema.parse(value),
    ),
  conversionImports: (page = 1) =>
    apiRequest(
      `/admin/commerce/conversion-imports${query(page)}`,
      { method: 'GET' },
      (value) => conversionImportsSchema.parse(value),
    ),
  triggerConversionImport: (providerConfigurationID: string) =>
    apiRequest(
      '/admin/commerce/conversion-imports',
      {
        method: 'POST',
        headers: { 'Idempotency-Key': crypto.randomUUID() },
        body: { provider_configuration_id: providerConfigurationID },
      },
      (value) => conversionImportSchema.parse(value),
    ),
  retryConversionImport: (id: string) =>
    apiRequest(
      `/admin/commerce/conversion-imports/${id}/retry`,
      { method: 'POST', headers: { 'Idempotency-Key': crypto.randomUUID() } },
      (value) => conversionImportSchema.parse(value),
    ),
  conversionReconciliations: (page = 1) =>
    apiRequest(
      `/admin/commerce/reconciliations${query(page)}`,
      { method: 'GET' },
      (value) => reconciliationsSchema.parse(value),
    ),
  reconcileConversionImport: (id: string) =>
    apiRequest(
      `/admin/commerce/conversion-imports/${id}/reconcile`,
      { method: 'POST', headers: { 'Idempotency-Key': crypto.randomUUID() } },
      (value) => reconciliationsSchema.shape.items.element.parse(value),
    ),
  monetizationMetrics: () =>
    apiRequest('/admin/commerce/metrics', { method: 'GET' }, (value) =>
      monetizationReportSchema.parse(value),
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
