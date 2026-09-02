import { describe, expect, it } from 'vitest'

import {
  evidenceSummary,
  latestObservation,
  shortRevision,
} from './productRecord'
import { productDetailFixture } from './productDetailFixture'

describe('evidenceSummary', () => {
  it('counts the visible facts and takes the median confidence', () => {
    expect(evidenceSummary(productDetailFixture())).toEqual({
      count: 3,
      medianConfidence: 90,
    })
  })

  it('averages the two middle values for an even count', () => {
    const product = productDetailFixture()
    const [price, quality, description] = product.evidence
    if (!price || !quality || !description) throw new Error('fixture')
    expect(
      evidenceSummary(
        productDetailFixture({
          evidence: [
            price,
            quality,
            description,
            { ...description, fact_key: 'name', confidence: 60 },
          ],
        }),
      ),
    ).toEqual({ count: 4, medianConfidence: 80 })
  })

  // A fact the page does not show cannot be counted as evidence for it.
  it('ignores evidence for facts that do not apply to the product', () => {
    const product = productDetailFixture()
    const [price] = product.evidence
    if (!price) throw new Error('fixture')
    expect(
      evidenceSummary(
        productDetailFixture({
          evidence: [price, { ...price, fact_key: 'slug', confidence: 1 }],
        }),
      ),
    ).toEqual({ count: 1, medianConfidence: 100 })
  })

  it('is null with nothing to summarise', () => {
    expect(evidenceSummary(productDetailFixture({ evidence: [] }))).toBeNull()
  })
})

describe('latestObservation', () => {
  it('names the most recent day any visible fact was observed', () => {
    expect(latestObservation(productDetailFixture())).toBe('Aug 26, 2026')
  })

  it('is null when nothing is dated', () => {
    expect(latestObservation(productDetailFixture({ evidence: [] }))).toBeNull()
  })
})

describe('shortRevision', () => {
  it('prints the first eight characters', () => {
    expect(shortRevision('0f6b2c1e-9a3d-4e5f-8b7c-6d5e4f3a2b1c')).toBe(
      '0f6b2c1e',
    )
  })
})
