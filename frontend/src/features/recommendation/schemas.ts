import { z } from 'zod'

import { productSummarySchema } from '../catalog/schemas'

// These must match the vocabulary the active recommendation policy declares.
// The engine rejects any value it has not declared, so a mismatch here shows
// up as a rejected brief rather than as a silently ignored field.
export const goals = [
  'client_services',
  'sell_products_online',
  'creator_business',
  'software_product',
  'solo_consulting',
  'local_business',
] as const
export const experiences = ['beginner', 'intermediate', 'advanced'] as const
export const trainingPreferences = [
  'all_in_one',
  'best_of_breed',
  'open_source',
  'no_code',
  'api_first',
  'privacy_focused',
  'eu_hosted',
] as const
export const priorities = [
  'budget',
  'ease_of_use',
  'integrations',
  'reliability',
  'vendor_stability',
  'data_portability',
] as const

const equipmentSchema = z.object({
  name: z.string().min(1).max(120),
  category_slug: z.string().min(1),
})

export const builderValuesSchema = z.object({
  goal: z.enum(goals).or(z.literal('')),
  experience: z.enum(experiences).or(z.literal('')),
  budget_minor: z.number().int().min(10_000).max(2_000_000),
  existing_equipment: z.array(equipmentSchema),
  training_preferences: z.array(z.enum(trainingPreferences)),
  priorities: z.array(z.enum(priorities)),
  free_text: z.string().max(1000),
})

export const recommendationInputSchema = z.object({
  goal: z.enum(goals),
  experience: z.enum(experiences),
  budget_minor: z.number().int().positive(),
  currency: z.literal('USD'),
  existing_equipment: z.array(equipmentSchema),
  training_preferences: z.array(z.enum(trainingPreferences)).min(1),
  priorities: z.array(z.enum(priorities)).min(1),
  free_text: z.string().max(1000),
})

export const draftSchema = z.object({
  current_step: z.number().int().min(1).max(7),
  goal: z.enum(goals).nullable(),
  experience: z.enum(experiences).nullable(),
  budget_minor: z.number().int().positive().nullable(),
  currency: z.literal('USD').nullable(),
  // The API still returns this field for verticals that use it. A
  // non-spatial deployment neither sends nor reads it.
  available_space: z.unknown().nullable().optional(),
  existing_equipment: z
    .array(equipmentSchema)
    .nullable()
    .transform((v) => v ?? []),
  training_preferences: z
    .array(z.enum(trainingPreferences))
    .nullable()
    .transform((v) => v ?? []),
  priorities: z
    .array(z.enum(priorities))
    .nullable()
    .transform((v) => v ?? []),
  free_text: z.string(),
})

const moneySchema = z.object({
  amount_minor: z.number().int().nonnegative(),
  currency: z.string().length(3),
})
const breakdownSchema = z.object({
  goal_match: z.number().int().min(0).max(100),
  budget_match: z.number().int().min(0).max(100),
  space_match: z.number().int().min(0).max(100),
  experience_match: z.number().int().min(0).max(100),
  preference_match: z.number().int().min(0).max(100),
  quality: z.number().int().min(0).max(100),
  value: z.number().int().min(0).max(100),
  durability: z.number().int().min(0).max(100),
  compatibility: z.number().int().min(0).max(100),
  portability: z.number().int().min(0).max(100),
  noise: z.number().int().min(0).max(100),
})
const reasonSchema = z.object({
  code: z.string(),
  message: z.string(),
  dimension: z.string(),
  score: z.number().int().min(0).max(100),
})

export const recommendationResultSchema = z.object({
  recommendation_id: z.string().nullable(),
  setup_id: z.string().nullable(),
  setup_name: z.string().nullable(),
  saved: z.boolean(),
  status: z.enum(['complete', 'no_suitable_products']),
  total_cost: moneySchema,
  recommendation_score: z.number().int().min(0).max(100),
  fit: breakdownSchema,
  recommended_products: z.array(
    z.object({
      rank: z.number().int().positive(),
      quantity: z.number().int().positive(),
      score: z.number().int().min(0).max(100),
      breakdown: breakdownSchema,
      reasons: z.array(reasonSchema),
      product: productSummarySchema,
    }),
  ),
  alternatives: z.array(
    z.object({
      for_product_id: z.string(),
      type: z.enum(['cheaper', 'premium']),
      price_difference_minor: z.number().int(),
      score: z.number().int().min(0).max(100),
      reasons: z.array(reasonSchema),
      product: productSummarySchema,
    }),
  ),
  rejected_alternatives: z.array(
    z.object({
      code: z.string(),
      reason: z.string(),
      product: productSummarySchema,
    }),
  ),
  policy_version: z.string(),
  engine_version: z.string(),
  input: recommendationInputSchema,
})

export const setupListSchema = z.object({
  setups: z.array(
    z.object({
      id: z.string(),
      name: z.string(),
      item_count: z.number().int().nonnegative(),
      total_cost: moneySchema,
      recommendation_score: z.number().int().min(0).max(100),
      created_at: z.string(),
      updated_at: z.string(),
    }),
  ),
  page: z.number().int().positive(),
  page_size: z.number().int().positive().max(100),
  total: z.number().int().nonnegative(),
  total_pages: z.number().int().nonnegative(),
})

export type BuilderValues = z.infer<typeof builderValuesSchema>
export type RecommendationInput = z.infer<typeof recommendationInputSchema>
export type RecommendationDraft = z.output<typeof draftSchema>
export type RecommendationResult = z.infer<typeof recommendationResultSchema>
export type SetupList = z.infer<typeof setupListSchema>
