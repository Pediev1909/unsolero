import { describe, expect, it } from 'vitest'

import {
  defaultBuilderValues,
  fromDraft,
  fromRecommendationInput,
  toDraft,
  toRecommendationInput,
} from './model'

describe('recommendation builder model', () => {
  it('requires identity and space answers before producing an engine input', () => {
    expect(toRecommendationInput(defaultBuilderValues)).toBeNull()
  })

  it('maps UI space presets to objective dimensions', () => {
    const input = toRecommendationInput({
      ...defaultBuilderValues,
      goal: 'build_muscle',
      experience: 'beginner',
      space_preset: 'small_apartment',
      training_preferences: ['dumbbells'],
      priorities: ['compact'],
    })
    expect(input?.available_space).toEqual({
      length_mm: 2400,
      width_mm: 1800,
      height_mm: 2400,
      apartment_living: true,
    })
  })

  it('round-trips a structured authenticated draft', () => {
    const values = {
      ...defaultBuilderValues,
      goal: 'strength' as const,
      experience: 'advanced' as const,
      space_preset: 'full_garage' as const,
      existing_equipment: [{ name: 'Barbell', category_slug: 'barbells' }],
      training_preferences: ['barbell' as const],
      priorities: ['durability' as const],
    }
    expect(fromDraft(toDraft(values, 6))).toEqual(values)
  })

  it('reopens a saved recommendation input for editing', () => {
    const values = {
      ...defaultBuilderValues,
      goal: 'strength' as const,
      experience: 'advanced' as const,
      space_preset: 'full_garage' as const,
      training_preferences: ['barbell' as const],
      priorities: ['quality' as const],
    }
    const input = toRecommendationInput(values)
    expect(input && fromRecommendationInput(input)).toEqual(values)
  })
})
