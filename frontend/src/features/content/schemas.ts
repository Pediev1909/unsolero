import { z } from 'zod'

import { categorySchema, productSummarySchema } from '../catalog/schemas'

export const contentTypeSchema = z.enum([
  'article',
  'guide',
  'buying_guide',
  'comparison',
  // A whole set of tools for one kind of business and budget, with what was
  // left out and why. Served at /stacks/{slug}; the hub is /stacks.
  'stack',
])

const heroImageSchema = z.object({
  url: z.string(),
  alt_text: z.string(),
  is_primary: z.boolean(),
})

export const contentSummarySchema = z.object({
  id: z.string().uuid(),
  type: contentTypeSchema,
  title: z.string(),
  slug: z.string(),
  path: z.string().startsWith('/'),
  description: z.string(),
  hero_image: heroImageSchema,
  author_name: z.string(),
  published_at: z.string().datetime(),
  updated_at: z.string().datetime(),
  // The products the piece compares. A card is named after them, which is what
  // stops thirteen comparisons sharing one illustration from looking identical.
  covered: z
    .array(
      z.object({
        name: z.string(),
        price_minor: z.number().int().nonnegative(),
        currency: z.string(),
      }),
    )
    .default([]),
})

const contentBlockSchema = z.object({
  type: z.enum([
    'paragraph',
    'heading',
    'unordered_list',
    'ordered_list',
    'quote',
    'callout',
    'cta',
    'pros_cons',
    'faq',
    'offer',
  ]),
  heading: z.string().optional(),
  text: z.string().optional(),
  items: z.array(z.string()).optional(),
  attribution: z.string().optional(),
  // 'cta' only. `promotion` is a slug, never a URL — the destination lives in
  // commerce.affiliate_promotions and the block only chooses which approved
  // one to show. See the BlockCTA comment in the Go domain for why.
  promotion: z.string().optional(),
  // 'cta' and 'offer'. The text on the control; an offer block without one
  // falls back to "View at {merchant}".
  label: z.string().optional(),
  // 'pros_cons' only. The server requires both sides, 1–8 each — a list of
  // strengths with no trade-offs is an advertisement. Optional here because
  // the shape is the server's to enforce, and a cached response must not
  // fail validation and blank the page.
  pros: z.array(z.string()).optional(),
  cons: z.array(z.string()).optional(),
  // 'faq' only.
  questions: z
    .array(z.object({ question: z.string(), answer: z.string() }))
    .optional(),
  // 'offer' only. A catalog product slug, never a URL: the vendor destination
  // is the product's live affiliate offer, loaded when the block renders, so
  // an editor can only point at what the catalog already serves.
  product: z.string().optional(),
})

export const contentDetailSchema = contentSummarySchema.extend({
  content: z.array(contentBlockSchema),
  author: z.object({
    name: z.string(),
    slug: z.string(),
    bio: z.string(),
    avatar_url: z.string().nullable(),
  }),
  related_products: z.array(productSummarySchema),
  related_categories: z.array(categorySchema),
  related_content: z.array(contentSummarySchema),
  seo: z.object({
    title: z.string(),
    description: z.string(),
    canonical_url: z.string().url(),
  }),
})

export type ContentType = z.infer<typeof contentTypeSchema>
export type ContentSummary = z.infer<typeof contentSummarySchema>
export type ContentBlock = z.infer<typeof contentBlockSchema>
export type ContentDetail = z.infer<typeof contentDetailSchema>

export const contentAuthorPageSchema = z.object({
  author: z.object({
    name: z.string(),
    slug: z.string(),
    bio: z.string(),
    avatar_url: z.string().nullable(),
  }),
  entries: z.array(contentSummarySchema),
})

export type ContentAuthorPage = z.infer<typeof contentAuthorPageSchema>
