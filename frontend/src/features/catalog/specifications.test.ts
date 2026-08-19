import { describe, expect, it } from 'vitest'

import { productDetailSchema } from './schemas'
import { specifications } from './specifications'

// Every product in a software catalog reports zero for the physical fields the
// equipment vertical needed. Rendering them produced a "Specifications" heading
// followed by "0 × 0 × 0 mm", "0 kg", a blank material and "Not specified" — on
// the page a vendor reviewing an affiliate application is most likely to open.
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

describe('specifications', () => {
  it('lists nothing for a product with no physical facts', () => {
    const product = productDetailSchema.parse(softwareProduct)
    expect(specifications(product)).toEqual([])
  })

  it('lists the physical facts a product does have', () => {
    const product = productDetailSchema.parse({
      ...softwareProduct,
      dimensions: { length_mm: 1200, width_mm: 800, height_mm: 2100 },
      weight_grams: 64000,
      material: 'powder-coated steel',
      warranty_months: 24,
    })
    expect(specifications(product).map(([label]) => label)).toEqual([
      'Dimensions',
      'Weight',
      'Material',
      'Warranty',
    ])
  })
})
