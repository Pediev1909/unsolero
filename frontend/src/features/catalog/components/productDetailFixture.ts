import type { ProductDetail } from '../schemas'

/**
 * One complete product for component tests, with the evidence a real SaaS
 * product carries: a dated, sourced price plus a few scored facts.
 */
export function productDetailFixture(
  overrides: Partial<ProductDetail> = {},
): ProductDetail {
  return {
    id: 'a7f1c2d4-5b6e-4a8c-9d0e-1f2a3b4c5d6e',
    name: 'Mailchimp Standard',
    slug: 'mailchimp-standard',
    brand: { name: 'Mailchimp', slug: 'mailchimp' },
    category: { name: 'Email marketing', slug: 'email-marketing' },
    price: { amount_minor: 2000, currency: 'USD' },
    primary_image: null,
    // The string is what the server derives from the object below; a fixture
    // where the two disagree would test a state the API never produces.
    key_specification: {
      label: 'Billing',
      value: 'Flat rate, monthly billing',
    },
    billing: {
      period: 'monthly',
      unit: 'flat',
      unit_note: null,
      annual_price_minor: null,
    },
    suitability: [{ key: 'beginner', label: 'Easy to adopt', score: 84 }],
    scores: {
      quality: 80,
      value: 72,
      durability: 88,
      beginner: 84,
      advanced: 61,
      apartment: 0,
      noise: 0,
      portability: 70,
    },
    is_demo: false,
    description: 'Email marketing with automation for small teams.',
    images: [],
    dimensions: { length_mm: 0, width_mm: 0, height_mm: 0 },
    weight_grams: 0,
    max_capacity_grams: null,
    material: '',
    warranty_months: 0,
    attributes: [],
    strengths: [
      { key: 'templates', label: 'Template library', score: 86 },
      { key: 'deliverability', label: 'Deliverability', score: 81 },
    ],
    weaknesses: [{ key: 'price_scaling', label: 'Price at scale', score: 74 }],
    use_cases: [
      { key: 'newsletters', label: 'Small-list newsletters', score: 90 },
      { key: 'ecommerce', label: 'Shop follow-ups', score: 70 },
    ],
    alternatives: [],
    fact_revision_id: '0f6b2c1e-9a3d-4e5f-8b7c-6d5e4f3a2b1c',
    score_revision_id: 'c1b2a3f4-5e6d-4c7b-8a9f-0e1d2c3b4a5f',
    evidence: [
      {
        fact_key: 'price',
        classification: 'manufacturer_claim',
        source_type: 'manufacturer_documentation',
        source_title: 'Mailchimp pricing',
        source_url: 'https://mailchimp.com/pricing/',
        observed_at: '2026-08-26T12:00:00Z',
        expires_at: null,
        confidence: 100,
        is_fictional: false,
      },
      {
        fact_key: 'score.quality',
        classification: 'editorial_assessment',
        source_type: 'editorial_assessment',
        source_title: 'UNSOLERO editorial review',
        source_url: null,
        observed_at: '2026-08-20T12:00:00Z',
        expires_at: null,
        confidence: 70,
        is_fictional: false,
      },
      {
        fact_key: 'description',
        classification: 'verified_fact',
        source_type: 'manufacturer_documentation',
        source_title: 'Mailchimp product page',
        source_url: null,
        observed_at: '2026-08-10T12:00:00Z',
        expires_at: null,
        confidence: 90,
        is_fictional: false,
      },
    ],
    ...overrides,
  }
}
