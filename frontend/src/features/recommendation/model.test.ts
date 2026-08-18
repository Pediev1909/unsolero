import { describe, expect, it } from 'vitest'

import {
  defaultBuilderValues,
  fromDraft,
  fromRecommendationInput,
  toDraft,
  toRecommendationInput,
} from './model'

describe('recommendation builder model', () => {
  it('requires identity answers before producing an engine input', () => {
    expect(toRecommendationInput(defaultBuilderValues)).toBeNull()
  })

  it('produces an engine input without any spatial answer', () => {
    const input = toRecommendationInput({
      ...defaultBuilderValues,
      goal: 'client_services',
      experience: 'beginner',
      training_preferences: ['best_of_breed'],
      priorities: ['budget'],
    })

    expect(input).not.toBeNull()
    expect(input?.goal).toBe('client_services')
    // A software stack has no footprint, so the brief carries no room
    // measurement at all rather than an invented one.
    expect(input).not.toHaveProperty('available_space')
  })

  it('round-trips a structured authenticated draft', () => {
    const values = {
      ...defaultBuilderValues,
      goal: 'solo_consulting' as const,
      experience: 'advanced' as const,
      existing_equipment: [{ name: 'CRM', category_slug: 'crm' }],
      training_preferences: ['api_first' as const],
      priorities: ['data_portability' as const],
    }
    expect(fromDraft(toDraft(values, 6))).toEqual(values)
  })

  it('reopens a saved recommendation input for editing', () => {
    const values = {
      ...defaultBuilderValues,
      goal: 'software_product' as const,
      experience: 'advanced' as const,
      training_preferences: ['api_first' as const],
      priorities: ['reliability' as const],
    }
    const input = toRecommendationInput(values)
    expect(input && fromRecommendationInput(input)).toEqual(values)
  })
})
