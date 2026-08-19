import { z } from 'zod'

const timestamp = z.string().datetime()
const nullableString = z.string().nullable()
const pageFields = {
  page: z.number().int().positive(),
  page_size: z.number().int().positive(),
  total: z.number().int().nonnegative(),
  total_pages: z.number().int().nonnegative(),
}

export const dashboardSchema = z.object({
  counts: z.object({
    products: z.number().int().nonnegative(),
    published: z.number().int().nonnegative(),
    offers: z.number().int().nonnegative(),
    active_offers: z.number().int().nonnegative(),
    users: z.number().int().nonnegative(),
    recommendations: z.number().int().nonnegative(),
  }),
  analytics: z.object({
    recommendation_starts: z.number().int().nonnegative(),
    completed_recommendations: z.number().int().nonnegative(),
    product_views: z.number().int().nonnegative(),
    affiliate_clicks: z.number().int().nonnegative(),
    saved_products: z.number().int().nonnegative(),
    saved_setups: z.number().int().nonnegative(),
  }),
})

const rankedEntitySchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  count: z.number().int().nonnegative(),
})

export const analyticsReportSchema = z.object({
  summary: z.object({
    users: z.number().int().nonnegative(),
    recommendation_sessions: z.number().int().nonnegative(),
    onboarding_started: z.number().int().nonnegative(),
    onboarding_completed: z.number().int().nonnegative(),
    recommendation_completion_rate: z.number().nonnegative().nullable(),
    product_views: z.number().int().nonnegative(),
    affiliate_clicks: z.number().int().nonnegative(),
    affiliate_clicks_raw: z.number().int().nonnegative(),
    affiliate_ctr: z.number().nonnegative().nullable(),
  }),
  most_recommended_products: z.array(rankedEntitySchema),
  most_viewed_products: z.array(rankedEntitySchema),
  most_clicked_products: z.array(rankedEntitySchema),
  top_merchants: z.array(rankedEntitySchema),
  top_categories: z.array(rankedEntitySchema),
  traffic_sources: z.array(
    z.object({
      source: z.string(),
      count: z.number().int().nonnegative(),
    }),
  ),
  window: z.object({
    from: timestamp,
    to: timestamp,
    reportable_from: timestamp,
    complete_through: timestamp,
    coverage: z.enum(['partial', 'complete']),
    data_state: z.enum(['no_data', 'insufficient_data', 'available']),
    layer: z.literal('validated_filtered'),
    minimum_sample_size: z.number().int().positive(),
  }),
  ingestion: z.object({
    received: z.number().int().nonnegative(),
    accepted: z.number().int().nonnegative(),
    rejected: z.number().int().nonnegative(),
    privacy_filtered: z.number().int().nonnegative(),
    bot_filtered: z.number().int().nonnegative(),
    deduplicated: z.number().int().nonnegative(),
  }),
})

export const categorySchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  slug: z.string(),
  is_active: z.boolean(),
  products: z.number().int().nonnegative(),
  updated_at: timestamp,
})

export const brandSchema = categorySchema

export const merchantSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  slug: z.string(),
  website_url: z.string().url(),
  country_code: z.string(),
  trust_score: z.number(),
  status: z.string(),
  offers: z.number().int().nonnegative(),
  updated_at: timestamp,
})

const scoresSchema = z.object({
  quality: z.number(),
  value: z.number(),
  durability: z.number(),
  beginner: z.number(),
  advanced: z.number(),
  apartment: z.number(),
  noise: z.number(),
  portability: z.number(),
})

const imageSchema = z.object({
  id: z.string().uuid(),
  url: z.string(),
  alt_text: z.string(),
  sort_order: z.number().int(),
  is_primary: z.boolean(),
})

const attributeSchema = z.object({
  id: z.string().uuid(),
  key: z.string(),
  type: z.enum(['number', 'text', 'boolean']),
  numeric_value: z.number().nullable(),
  text_value: z.string().nullable(),
  boolean_value: z.boolean().nullable(),
  unit: z.string().nullable(),
  is_filterable: z.boolean(),
})

export const productSchema = z.object({
  id: z.string().uuid(),
  category_id: z.string().uuid(),
  category_name: z.string(),
  brand_id: z.string().uuid(),
  brand_name: z.string(),
  name: z.string(),
  slug: z.string(),
  description: z.string(),
  price_minor: z.number().int().nonnegative(),
  currency: z.string(),
  // Zero, not positive: software has no physical form and the API reports it
  // as such. Demanding a measurement made the entire product list fail to
  // parse, so the admin products page showed "Something went wrong" while the
  // API behind it returned 200 with the data intact.
  length_mm: z.number().nonnegative(),
  width_mm: z.number().nonnegative(),
  height_mm: z.number().nonnegative(),
  weight_grams: z.number().nonnegative(),
  max_capacity_grams: z.number().nonnegative().nullable(),
  material: z.string(),
  warranty_months: z.number().int().nonnegative(),
  scores: scoresSchema,
  status: z.enum(['draft', 'published', 'discontinued']),
  images: z.array(imageSchema),
  attributes: z.array(attributeSchema),
  updated_at: timestamp,
})

export const productPageSchema = z.object({
  items: z.array(productSchema),
  ...pageFields,
})

export const offerSchema = z.object({
  id: z.string().uuid(),
  merchant_id: z.string().uuid(),
  merchant_name: z.string(),
  product_id: z.string().uuid(),
  product_name: z.string(),
  merchant_sku: z.string(),
  product_url: z.string().url(),
  price_minor: z.number().int().nonnegative(),
  shipping_minor: z.number().int().nonnegative(),
  currency: z.string(),
  availability: z.string(),
  condition: z.string(),
  is_active: z.boolean(),
  last_checked_at: timestamp,
  expires_at: timestamp.nullable(),
  freshness_status: z.enum(['fresh', 'expired', 'platform_policy']),
  affiliate_links: z.number().int().nonnegative(),
  updated_at: timestamp,
})

export const offerPageSchema = z.object({
  items: z.array(offerSchema),
  ...pageFields,
})

export const affiliateLinkSchema = z.object({
  id: z.string().uuid(),
  offer_id: z.string().uuid(),
  product_name: z.string(),
  merchant_name: z.string(),
  provider: z.string(),
  destination_url: z.string().url(),
  external_reference: nullableString,
  disclosure_label: z.string(),
  is_active: z.boolean(),
  priority: z.number().int(),
  program_id: nullableString,
  commission_type: z.string(),
  commission_rate_bps: z.number().int().nullable(),
  commission_amount_minor: z.number().int().nullable(),
  commission_currency: nullableString,
  updated_at: timestamp,
})

export const affiliatePageSchema = z.object({
  items: z.array(affiliateLinkSchema),
  ...pageFields,
})

export const commerceProviderSchema = z.object({
  id: z.string().uuid(),
  merchant_id: z.string().uuid(),
  merchant_name: z.string(),
  provider_key: z.string(),
  adapter_key: z.string(),
  external_merchant_id: z.string(),
  credential_reference: nullableString,
  lifecycle_status: z.enum([
    'disabled',
    'configured',
    'active',
    'degraded',
    'suspended',
  ]),
  configuration_verified_at: timestamp.nullable(),
  schedule_interval_minutes: z.number().int().positive(),
  freshness_ttl_minutes: z.number().int().positive(),
  cursor: nullableString,
  next_import_at: timestamp.nullable(),
  last_import_started_at: timestamp.nullable(),
  last_import_succeeded_at: timestamp.nullable(),
  last_import_failed_at: timestamp.nullable(),
  consecutive_failures: z.number().int().nonnegative(),
  last_error_code: nullableString,
  conversion_cursor: nullableString,
  next_conversion_import_at: timestamp.nullable(),
  last_conversion_import_succeeded_at: timestamp.nullable(),
  last_conversion_import_failed_at: timestamp.nullable(),
  conversion_consecutive_failures: z.number().int().nonnegative(),
  last_conversion_error_code: nullableString,
  conversion_ingestion_enabled: z.boolean(),
  conversion_configuration_verified_at: timestamp.nullable(),
  created_at: timestamp,
  updated_at: timestamp,
})

export const commerceProvidersSchema = z.object({
  items: z.array(commerceProviderSchema),
})

export const commerceImportSchema = z.object({
  id: z.string().uuid(),
  provider_configuration: commerceProviderSchema,
  trigger: z.enum(['scheduled', 'manual', 'retry']),
  status: z.enum([
    'queued',
    'running',
    'retry_wait',
    'succeeded',
    'partial',
    'failed',
    'cancelled',
  ]),
  idempotency_key: z.string(),
  requested_by: nullableString,
  cursor_before: nullableString,
  cursor_after: nullableString,
  attempt_count: z.number().int().nonnegative(),
  max_attempts: z.number().int().positive(),
  records_received: z.number().int().nonnegative(),
  records_applied: z.number().int().nonnegative(),
  records_rejected: z.number().int().nonnegative(),
  offers_deactivated: z.number().int().nonnegative(),
  error_code: nullableString,
  error_message: nullableString,
  next_retry_at: timestamp.nullable(),
  started_at: timestamp.nullable(),
  completed_at: timestamp.nullable(),
  created_at: timestamp,
  updated_at: timestamp,
})

export const commerceImportsSchema = z.object({
  items: z.array(commerceImportSchema),
  ...pageFields,
})

export const commerceImportFailureSchema = z.object({
  id: z.string().uuid(),
  import_run_id: z.string().uuid(),
  external_record_id: nullableString,
  error_code: z.string(),
  error_message: z.string(),
  record_fingerprint: nullableString,
  created_at: timestamp,
})

export const commerceImportFailuresSchema = z.object({
  items: z.array(commerceImportFailureSchema),
  ...pageFields,
})

const conversionOrderStatus = z.enum([
  'pending',
  'confirmed',
  'cancelled',
  'reversed',
  'rejected',
])
const commissionStatus = z.enum([
  'pending',
  'approved',
  'reversed',
  'rejected',
  'paid',
])
const metricStatus = z.enum(['available', 'no_data', 'insufficient_data'])

export const verifiedConversionSchema = z.object({
  id: z.string().uuid(),
  provider_configuration_id: z.string().uuid(),
  provider: z.string(),
  merchant_id: z.string().uuid(),
  merchant_name: z.string(),
  external_conversion_id: z.string(),
  order_reference: nullableString,
  order_status: conversionOrderStatus,
  order_value_minor: z.number().int().nonnegative().nullable(),
  order_currency: nullableString,
  commission_amount_minor: z.number().int().nonnegative().nullable(),
  commission_currency: nullableString,
  commission_status: commissionStatus.nullable(),
  attribution_status: z.enum(['attributed', 'unattributed']),
  click_id: nullableString,
  recommendation_id: nullableString,
  source: nullableString,
  campaign: nullableString,
  verification_state: z.literal('verified'),
  event_timestamp: timestamp,
  received_at: timestamp,
  updated_at: timestamp,
  reconciliation_status: z
    .enum(['matched', 'missing', 'conflicting', 'stale', 'unresolved'])
    .nullable(),
})

export const verifiedConversionsSchema = z.object({
  items: z.array(verifiedConversionSchema),
  ...pageFields,
})

export const conversionImportSchema = z.object({
  id: z.string().uuid(),
  provider_configuration: commerceProviderSchema,
  trigger: z.enum(['scheduled', 'manual', 'retry']),
  status: z.enum([
    'queued',
    'running',
    'retry_wait',
    'succeeded',
    'partial',
    'failed',
    'cancelled',
  ]),
  attempt_count: z.number().int().nonnegative(),
  max_attempts: z.number().int().positive(),
  records_received: z.number().int().nonnegative(),
  records_applied: z.number().int().nonnegative(),
  records_rejected: z.number().int().nonnegative(),
  cursor_before: nullableString,
  cursor_after: nullableString,
  coverage_start: timestamp.nullable(),
  coverage_end: timestamp.nullable(),
  error_code: nullableString,
  error_message: nullableString,
  created_at: timestamp,
  started_at: timestamp.nullable(),
  completed_at: timestamp.nullable(),
})

export const conversionImportsSchema = z.object({
  items: z.array(conversionImportSchema),
  ...pageFields,
})

export const reconciliationSchema = z.object({
  id: z.string().uuid(),
  provider_configuration: commerceProviderSchema,
  status: z.enum(['running', 'succeeded', 'partial', 'failed']),
  coverage_start: timestamp,
  coverage_end: timestamp,
  matched: z.number().int().nonnegative(),
  missing: z.number().int().nonnegative(),
  conflicting: z.number().int().nonnegative(),
  stale: z.number().int().nonnegative(),
  unresolved: z.number().int().nonnegative(),
  error_code: nullableString,
  started_at: timestamp,
  completed_at: timestamp.nullable(),
})

export const reconciliationsSchema = z.object({
  items: z.array(reconciliationSchema),
  ...pageFields,
})

const ratioMetricSchema = z.object({
  status: metricStatus,
  value: z.number().nullable(),
  numerator: z.number().int().nonnegative(),
  denominator: z.number().int().nonnegative(),
  definition: z.string(),
})
const currencyMetricGroupSchema = z.object({
  status: metricStatus,
  values: z.array(
    z.object({
      currency: z.string(),
      amount_minor: z.number().int(),
      denominator: z.number().int().nonnegative(),
      value_minor: z.number().nullable(),
    }),
  ),
  definition: z.string(),
})

export const monetizationReportSchema = z.object({
  window_start: timestamp,
  window_end: timestamp,
  fresh_through: timestamp.nullable(),
  affiliate_conversion_rate: ratioMetricSchema,
  earnings_per_click: currencyMetricGroupSchema,
  revenue_per_visitor: currencyMetricGroupSchema,
  revenue_per_recommendation: currencyMetricGroupSchema,
  commission: currencyMetricGroupSchema,
  reversal_rate: ratioMetricSchema,
  repeat_user_rate: ratioMetricSchema,
  currency_policy: z.string(),
})

export const recommendationSchema = z.object({
  id: z.string().uuid(),
  session_id: z.string().uuid(),
  user_email: nullableString,
  goal: z.string(),
  experience: z.string(),
  session_status: z.string(),
  objective_score: z.number(),
  total_price_minor: z.number(),
  currency: z.string(),
  policy_version: z.string(),
  engine_version: z.string(),
  created_at: timestamp,
})

export const recommendationPageSchema = z.object({
  items: z.array(recommendationSchema),
  ...pageFields,
})

const reasonSchema = z.object({
  code: z.string(),
  message: z.string(),
  dimension: z.string(),
  score: z.number(),
})

export const recommendationDetailSchema = z.object({
  recommendation: recommendationSchema,
  scores: z.record(z.string(), z.number()),
  items: z.array(
    z.object({
      product_id: z.string().uuid(),
      product_name: z.string(),
      item_type: z.string(),
      rank: z.number(),
      quantity: z.number(),
      objective_score: z.number(),
      reason_code: z.string(),
      reason_summary: z.string(),
      rejection_code: nullableString,
      reasons: z.array(reasonSchema),
    }),
  ),
})

export const userPageSchema = z.object({
  items: z.array(
    z.object({
      id: z.string().uuid(),
      email: z.string().email(),
      status: z.string(),
      roles: z.array(z.string()),
      last_login_at: timestamp.nullable(),
      created_at: timestamp,
    }),
  ),
  ...pageFields,
})

export const eventPageSchema = z.object({
  items: z.array(
    z.object({
      id: z.string().uuid(),
      name: z.string(),
      user_id: nullableString,
      anonymous_id: nullableString,
      session_id: nullableString,
      surface: z.string(),
      properties: z.record(z.string(), z.unknown()),
      page_path: nullableString,
      traffic_source: nullableString,
      traffic_medium: nullableString,
      campaign: nullableString,
      referrer_host: nullableString,
      consent_state: z.string(),
      occurred_at: timestamp,
    }),
  ),
  ...pageFields,
})

export const referencesSchema = z.object({
  categories: z.array(categorySchema),
  brands: z.array(brandSchema),
  merchants: z.array(merchantSchema),
  products: z.array(
    z.object({ id: z.string().uuid(), name: z.string(), slug: z.string() }),
  ),
})

const evidenceRevisionSchema = z.object({
  fact_revision_id: z.string().uuid(),
  score_revision_id: z.string().uuid(),
  fact_version: z.number().int().positive(),
  score_version: z.number().int().positive(),
  status: z.enum([
    'draft',
    'in_review',
    'approved',
    'published',
    'rejected',
    'superseded',
  ]),
  created_at: timestamp,
  submitted_at: timestamp.nullable(),
  reviewed_at: timestamp.nullable(),
  published_at: timestamp.nullable(),
  valid_until: timestamp.nullable(),
  review_note: z.string(),
})

const evidenceSourceSchema = z.object({
  id: z.string().uuid(),
  source_type: z.enum([
    'manufacturer_documentation',
    'independent_testing',
    'verified_merchant_data',
    'editorial_assessment',
    'demo_fixture',
  ]),
  title: z.string(),
  publisher: z.string(),
  source_url: z.string().url().nullable(),
  is_fictional: z.boolean(),
  review_status: z.enum(['pending', 'verified', 'rejected']),
  reviewed_at: timestamp.nullable(),
  review_note: z.string(),
  created_at: timestamp,
})

export const productGovernanceSchema = z.object({
  product_id: z.string().uuid(),
  product_name: z.string(),
  status: z.enum(['draft', 'published', 'discontinued']),
  published_fact_revision_id: z.string().uuid().nullable(),
  published_score_revision_id: z.string().uuid().nullable(),
  revisions: z.array(evidenceRevisionSchema),
  provenance: z.array(
    z.object({
      fact_key: z.string(),
      score_key: z.string(),
      classification: z.string(),
      rationale: z.string(),
      observation: z.object({
        id: z.string().uuid(),
        observed_at: timestamp,
        expires_at: timestamp.nullable(),
        confidence: z.number().int().min(0).max(100),
        notes: z.string(),
      }),
      source: evidenceSourceSchema,
    }),
  ),
  audit: z.array(
    z.object({
      action: z.string(),
      actor_email: z.string().email().nullable(),
      changes: z.record(z.string(), z.string()),
      occurred_at: timestamp,
    }),
  ),
})

export const productGovernancePageSchema = z.object({
  items: z.array(productGovernanceSchema),
  ...pageFields,
})

export type DashboardData = z.infer<typeof dashboardSchema>
export type AnalyticsReportData = z.infer<typeof analyticsReportSchema>
export type AdminProduct = z.infer<typeof productSchema>
export type AdminOffer = z.infer<typeof offerSchema>
export type AdminAffiliateLink = z.infer<typeof affiliateLinkSchema>
export type CommerceProvider = z.infer<typeof commerceProviderSchema>
export type CommerceImport = z.infer<typeof commerceImportSchema>
export type VerifiedConversion = z.infer<typeof verifiedConversionSchema>
export type ConversionImport = z.infer<typeof conversionImportSchema>
export type Reconciliation = z.infer<typeof reconciliationSchema>
export type MonetizationReport = z.infer<typeof monetizationReportSchema>
export type AdminRecommendation = z.infer<typeof recommendationSchema>
export type AdminReferences = z.infer<typeof referencesSchema>
export type ProductGovernance = z.infer<typeof productGovernanceSchema>
