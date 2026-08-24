import { describe, expect, it } from 'vitest'

import { markColourFor, monogramFor, monogramMarks } from './brandMonogram'

describe('monogramFor', () => {
  // A vendor's capitalisation is part of its name. Uppercasing the mark would
  // misspell brands on a site whose argument is that its facts are checked.
  it('keeps a lowercase brand lowercase', () => {
    expect(monogramFor('monday.com')).toBe('m')
    expect(monogramFor('n8n')).toBe('n')
  })

  it('keeps an initialism the brand already uses', () => {
    expect(monogramFor('SE Ranking')).toBe('SE')
    expect(monogramFor('IBM Cloud')).toBe('IBM')
  })

  it('takes one initial from each of the first two words', () => {
    expect(monogramFor('Help Scout')).toBe('HS')
    expect(monogramFor('Google Workspace')).toBe('GW')
    expect(monogramFor('Simple Analytics')).toBe('SA')
  })

  it('takes a single initial from a one-word brand', () => {
    expect(monogramFor('Stripe')).toBe('S')
    expect(monogramFor('Zoho')).toBe('Z')
  })

  // A brand name should never be able to produce an empty tile.
  it('returns a mark for whitespace and empty input', () => {
    expect(monogramFor('   ')).toBe('—')
    expect(monogramFor('')).toBe('—')
  })
})

describe('markColourFor', () => {
  it('gives one category the same colour every time', () => {
    expect(markColourFor('Email marketing')).toBe(
      markColourFor('Email marketing'),
    )
  })

  it('ignores case, so a category is never split across two colours', () => {
    expect(markColourFor('CRM')).toBe(markColourFor('crm'))
  })

  it('always returns a colour from the palette', () => {
    for (const key of ['CRM', 'Analytics', 'Payments', 'SEO tools', '', 'x']) {
      expect(monogramMarks).toContain(markColourFor(key))
    }
  })
})
