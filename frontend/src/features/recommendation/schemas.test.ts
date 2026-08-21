import { describe, expect, it } from 'vitest'

import {
  recommendationInputSchema,
  recommendationPreviewInputSchema,
  recommendationResultSchema,
} from './schemas'

const brief = {
  goal: 'local_business' as const,
  experience: 'beginner' as const,
  budget_minor: 12_000,
  currency: 'USD' as const,
  existing_equipment: [],
  training_preferences: [],
  priorities: [],
  free_text: '',
}

describe('recommendation schemas', () => {
  // The wizard asks for both, so a finished brief should carry both.
  it('requires a preference and a priority on submit', () => {
    expect(recommendationInputSchema.safeParse(brief).success).toBe(false)
  })

  // A preview runs from the third question, before either has been asked.
  it('accepts a half-answered brief for preview', () => {
    expect(recommendationPreviewInputSchema.safeParse(brief).success).toBe(true)
  })

  // The result echoes the input back. Validating that echo with the stricter
  // schema turned a working 200 into a parse failure, and the panel rendered
  // nothing while the network showed success.
  it('parses a result whose echoed input is half answered', () => {
    const result = {
      recommendation_id: null,
      setup_id: null,
      setup_name: null,
      saved: false,
      status: 'complete' as const,
      total_cost: { amount_minor: 5642, currency: 'USD' },
      recommendation_score: 87,
      fit: {
        goal_match: 90,
        budget_match: 97,
        space_match: 100,
        experience_match: 80,
        preference_match: 50,
        quality: 85,
        value: 80,
        durability: 88,
        compatibility: 70,
        portability: 75,
        noise: 100,
      },
      recommended_products: [],
      alternatives: [],
      rejected_alternatives: [],
      policy_version: 'saas-v10',
      engine_version: 'test',
      input: brief,
    }
    const parsed = recommendationResultSchema.safeParse(result)
    expect(parsed.success).toBe(true)
  })
})
