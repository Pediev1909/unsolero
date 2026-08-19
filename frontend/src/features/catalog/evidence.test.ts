import { describe, expect, it } from 'vitest'

import { evidenceFactLabel, visibleEvidence } from './evidence'
import { productDetailSchema } from './schemas'

// The evidence section is the strongest claim this site makes: every number is
// tied to a reviewed source. It was labelling those sources with raw fact keys,
// so it offered "Score.Durability", "Warranty" and "Slug" as the proof.
const softwareProduct = {
  id: 'b8efd843-65a1-4847-99dc-5ad79ee83eca',
  name: 'ClickUp Unlimited',
  slug: 'clickup-unlimited',
  brand: { name: 'ClickUp', slug: 'clickup' },
  category: { name: 'Project management', slug: 'project-management' },
  price: { amount_minor: 1000, currency: 'USD' },
  primary_image: null,
  key_specification: { label: 'Billing', value: 'Per month' },
  suitability: [{ key: 'beginner', label: 'Easy to adopt', score: 74 }],
  scores: {
    quality: 82,
    value: 90,
    durability: 84,
    beginner: 74,
    advanced: 88,
    apartment: 0,
    noise: 0,
    portability: 78,
  },
  is_demo: false,
  description: 'Project and task tracking with unlimited use.',
  images: [],
  dimensions: { length_mm: 0, width_mm: 0, height_mm: 0 },
  weight_grams: 0,
  max_capacity_grams: null,
  material: '',
  warranty_months: 0,
  attributes: [],
  strengths: [{ key: 'value', label: 'Value', score: 90 }],
  weaknesses: [],
  use_cases: [{ key: 'advanced', label: 'Depth for power users', score: 88 }],
  alternatives: [],
  fact_revision_id: '2b1c5e0a-4d3f-4a1b-8c2d-9e0f1a2b3c4d',
  score_revision_id: '3c2d6f1b-5e4a-4b2c-9d3e-0f1a2b3c4d5e',
  evidence: [],
}

const withEvidence = (keys: string[]) => ({
  ...softwareProduct,
  evidence: keys.map((fact_key) => ({
    fact_key,
    classification: 'editorial_assessment',
    source_type: 'editorial_assessment',
    source_title: 'UNSOLERO suitability assessment',
    source_url: null,
    observed_at: '2026-08-18T00:00:00.000Z',
    expires_at: null,
    confidence: 80,
    is_fictional: false,
  })),
})

describe('evidenceFactLabel', () => {
  it('names a score in the words the rest of the page uses', () => {
    expect(evidenceFactLabel('score.durability')).toBe('Vendor stability')
    expect(evidenceFactLabel('score.portability')).toBe('Data portability')
  })

  it('falls back to a readable form of an unmapped key', () => {
    expect(evidenceFactLabel('some_new_fact')).toBe('some new fact')
  })
})

describe('visibleEvidence', () => {
  it('drops facts a software product does not have', () => {
    const product = productDetailSchema.parse(
      withEvidence([
        'score.value',
        'score.apartment',
        'score.noise',
        'warranty',
        'slug',
      ]),
    )
    expect(visibleEvidence(product).map((e) => e.fact_key)).toEqual([
      'score.value',
    ])
  })

  it('keeps a physical fact when the product actually has one', () => {
    const product = productDetailSchema.parse({
      ...withEvidence(['warranty']),
      warranty_months: 24,
    })
    expect(visibleEvidence(product)).toHaveLength(1)
  })
})
