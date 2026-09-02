import { z } from 'zod'

const namedResourceSchema = z.object({ name: z.string(), slug: z.string() })
const moneySchema = z.object({
  amount_minor: z.number().int().nonnegative(),
  currency: z.string().length(3),
})
const imageSchema = z.object({
  url: z.string(),
  alt_text: z.string(),
  is_primary: z.boolean(),
  width_px: z.number().int().positive().optional(),
  height_px: z.number().int().positive().optional(),
})
const insightSchema = z.object({
  key: z.string(),
  label: z.string(),
  score: z.number().int().min(0).max(100),
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

export const billingPeriods = ['monthly', 'annual', 'free', 'usage'] as const
export const billingUnits = [
  'flat',
  'per_user',
  'per_contacts',
  'per_transaction',
  'usage',
] as const

/**
 * How the listed `price` is charged.
 *
 * `price` stays the compared figure: the monthly-billing rate wherever the
 * vendor sells month to month. `period: 'annual'` means the vendor sells only
 * yearly contracts and `price` is that contract's per-month equivalent —
 * twenty-five products in the catalog are on that basis, and until this object
 * existed nothing in the data said so. `free` puts the price at zero; `usage`
 * makes it the entry usage price, with the note saying what a unit is.
 * `annual_price_minor` is the per-month figure on annual billing when the
 * vendor offers both.
 */
export const billingSchema = z.object({
  period: z.enum(billingPeriods),
  unit: z.enum(billingUnits),
  unit_note: z.string().nullable(),
  annual_price_minor: z.number().int().nonnegative().nullable(),
})

export const productSummarySchema = z.object({
  id: z.string(),
  name: z.string(),
  slug: z.string(),
  brand: namedResourceSchema,
  category: namedResourceSchema,
  price: moneySchema,
  primary_image: imageSchema.nullable(),
  key_specification: z.object({ label: z.string(), value: z.string() }),
  // The structured basis behind `key_specification`, which the server now
  // derives from it. Anything that only prints the basis may keep reading the
  // string; anything that needs its shape — the seat calculator, the
  // comparison's billing row, a "billed yearly" mark — reads this. Optional so
  // a response cached before the field existed still parses.
  billing: billingSchema.optional(),
  suitability: z.array(insightSchema),
  scores: scoresSchema,
  is_demo: z.boolean(),
  // Set by every endpoint that returns a grid the reader can act on. Null
  // means there is no servable affiliate offer, not that nobody looked: the
  // API runs each page through one batched lookup that applies the same
  // freshness and liveness checks the redirect will apply again.
  //
  // Optional rather than nullable-required, because a cached response from
  // before these fields existed must not fail validation and blank the page.
  purchase_path: z.string().nullish(),
  disclosure_label: z.string().nullish(),
  merchant_name: z.string().nullish(),
})

export const productPageSchema = z.object({
  products: z.array(productSummarySchema),
  page: z.number().int().positive(),
  page_size: z.number().int().positive(),
  total: z.number().int().nonnegative(),
  total_pages: z.number().int().nonnegative(),
})

const attributeSchema = z.object({
  key: z.string(),
  type: z.enum(['number', 'text', 'boolean']),
  numeric_value: z.number().optional(),
  text_value: z.string().optional(),
  boolean_value: z.boolean().optional(),
  unit: z.string().optional(),
})

export const productDetailSchema = productSummarySchema.extend({
  description: z.string(),
  images: z.array(imageSchema),
  // Software has no length, weight or material, so the API reports zero for
  // these. Requiring a positive number was left over from the equipment
  // vertical and made every product in the catalog fail to parse: the detail
  // page showed "Product unavailable" and the comparison refused to open,
  // while the request itself returned 200 with correct metadata, so nothing
  // upstream registered a fault. Nothing public renders these fields; they
  // are carried for physical products, which the backend gates separately.
  dimensions: z.object({
    length_mm: z.number().int().nonnegative(),
    width_mm: z.number().int().nonnegative(),
    height_mm: z.number().int().nonnegative(),
  }),
  weight_grams: z.number().int().nonnegative(),
  max_capacity_grams: z.number().int().nonnegative().nullable(),
  material: z.string(),
  warranty_months: z.number().int().nonnegative(),
  attributes: z.array(attributeSchema),
  strengths: z.array(insightSchema),
  weaknesses: z.array(insightSchema),
  use_cases: z.array(insightSchema),
  alternatives: z.array(productSummarySchema),
  fact_revision_id: z.string().uuid(),
  score_revision_id: z.string().uuid(),
  evidence: z.array(
    z.object({
      fact_key: z.string(),
      classification: z.enum([
        'verified_fact',
        'manufacturer_claim',
        'merchant_observation',
        'editorial_assessment',
      ]),
      source_type: z.enum([
        'manufacturer_documentation',
        'independent_testing',
        'verified_merchant_data',
        'editorial_assessment',
        'demo_fixture',
      ]),
      source_title: z.string(),
      source_url: z.string().url().nullable(),
      observed_at: z.string().datetime(),
      expires_at: z.string().datetime().nullable(),
      confidence: z.number().int().min(0).max(100),
      is_fictional: z.boolean(),
    }),
  ),
})

export const categorySchema = z.object({
  id: z.string(),
  name: z.string(),
  slug: z.string(),
  description: z.string(),
  // Optional because the single-category endpoint predates it and older
  // cached responses will not carry it. A missing count renders as no count
  // rather than as zero, which would read as "we cover nothing here".
  published_products: z.number().int().nonnegative().optional(),
})

export const brandSchema = categorySchema.extend({
  country_code: z.string().optional(),
})

export const offerSchema = z.object({
  id: z.string(),
  merchant: z.object({
    name: z.string(),
    slug: z.string(),
    country_code: z.string(),
    trust_score: z.number().int().min(0).max(100),
  }),
  price: moneySchema,
  shipping_minor: z.number().int().nonnegative(),
  landed_price_minor: z.number().int().nonnegative(),
  availability: z.enum(['in_stock', 'backorder']),
  condition: z.enum(['new', 'refurbished', 'used']),
  last_checked_at: z.string(),
  observed_at: z.string().nullable(),
  expires_at: z.string().nullable(),
  // Was z.literal('fresh'), which matched a Go constant that said 'fresh' for
  // every offer regardless of age. Both were honest only while the server
  // refused to serve anything older than 72 hours — and that refusal is what
  // silently took every affiliate link offline three days after its last seed.
  freshness_status: z.enum(['fresh', 'stale']),
  purchase_path: z.string().nullable(),
  disclosure_label: z.string().nullable(),
})

/**
 * One row of the catalog-wide offers listing: the product as the grid draws
 * it, with the vendor button fields set, and the slice of its offer a row
 * prints. There is no destination URL here on purpose; the only way out is
 * the product's tracked `purchase_path`.
 */
export const liveOfferSchema = z.object({
  product: productSummarySchema,
  offer: z.object({
    price: moneySchema,
    merchant_name: z.string(),
    last_checked_at: z.string(),
    freshness_status: z.enum(['fresh', 'stale']),
  }),
})

export const liveOffersSchema = z.object({
  items: z.array(liveOfferSchema),
  generated_at: z.string(),
})

export const productSelectionSchema = z.object({
  product_ids: z.array(z.string()),
})
export const wishlistSchema = productSelectionSchema.extend({
  page: z.number().int().positive(),
  page_size: z.number().int().positive().max(100),
  total: z.number().int().nonnegative(),
  total_pages: z.number().int().nonnegative(),
})

export type Billing = z.infer<typeof billingSchema>
export type ProductSummary = z.infer<typeof productSummarySchema>
export type ProductPage = z.infer<typeof productPageSchema>
export type ProductDetail = z.infer<typeof productDetailSchema>
export type Category = z.infer<typeof categorySchema>
export type Brand = z.infer<typeof brandSchema>
export type Offer = z.infer<typeof offerSchema>
export type LiveOffer = z.infer<typeof liveOfferSchema>
export type LiveOffers = z.infer<typeof liveOffersSchema>
export type ProductInsight = z.infer<typeof insightSchema>
