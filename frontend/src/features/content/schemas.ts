import { z } from 'zod'

import { categorySchema, productSummarySchema } from '../catalog/schemas'

export const contentTypeSchema = z.enum([
  'article',
  'guide',
  'buying_guide',
  'comparison',
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
})

const contentBlockSchema = z.object({
  type: z.enum([
    'paragraph',
    'heading',
    'unordered_list',
    'ordered_list',
    'quote',
    'callout',
  ]),
  heading: z.string().optional(),
  text: z.string().optional(),
  items: z.array(z.string()).optional(),
  attribution: z.string().optional(),
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
