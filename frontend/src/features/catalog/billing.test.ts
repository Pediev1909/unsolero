import { describe, expect, it } from 'vitest'

import {
  annualPriceMinor,
  annualSaving,
  formatBillingBasis,
  isAnnualOnly,
  isPerUser,
} from './billing'
import type { Billing } from './schemas'

function billing(overrides: Partial<Billing> = {}): Billing {
  return {
    period: 'monthly',
    unit: 'flat',
    unit_note: null,
    annual_price_minor: null,
    ...overrides,
  }
}

const legacy = { value: 'Per month' }

describe('formatBillingBasis', () => {
  // The same table the server uses to derive key_specification, so a page
  // with the object and a page with only the string agree word for word.
  it.each<[Partial<Billing>, string]>([
    [{ unit: 'flat', period: 'monthly' }, 'Flat rate, monthly billing'],
    [{ unit: 'per_user', period: 'monthly' }, 'Per user, monthly billing'],
    [{ unit: 'flat', period: 'annual' }, 'Flat rate, billed yearly'],
    [{ unit: 'per_user', period: 'annual' }, 'Per user, billed yearly'],
    [{ unit: 'flat', period: 'free' }, 'Flat rate, free plan'],
    [
      { unit: 'per_contacts', period: 'monthly' },
      'Per contact tier, monthly billing',
    ],
    [
      {
        unit: 'per_contacts',
        period: 'monthly',
        unit_note: 'Up to 500 contacts',
      },
      'Up to 500 contacts, monthly billing',
    ],
    // The note is stored as it reads mid-sentence and opens the phrase here,
    // so its first letter is raised — the API's own worked example.
    [
      {
        unit: 'per_contacts',
        period: 'monthly',
        unit_note: 'at 1,000 contacts',
      },
      'At 1,000 contacts, monthly billing',
    ],
    [{ unit: 'per_transaction', period: 'usage' }, 'Per transaction'],
    [
      {
        unit: 'per_transaction',
        period: 'usage',
        unit_note: '2.9% + 30¢ per card payment',
      },
      '2.9% + 30¢ per card payment',
    ],
    [{ unit: 'usage', period: 'usage' }, 'Usage-based'],
    [
      { unit: 'usage', period: 'usage', unit_note: 'Per 1,000 tasks' },
      'Per 1,000 tasks',
    ],
  ])('%o reads as %s', (overrides, expected) => {
    expect(formatBillingBasis(billing(overrides), legacy)).toBe(expected)
  })

  // A note on a flat or per-user price is not a substitute for the unit: the
  // reader still needs to know it is per user.
  it('keeps the generic phrase for flat and per-user units even with a note', () => {
    expect(
      formatBillingBasis(
        billing({ unit: 'per_user', unit_note: 'Minimum 3 seats' }),
        legacy,
      ),
    ).toBe('Per user, monthly billing')
  })

  it('ignores a blank note', () => {
    expect(
      formatBillingBasis(billing({ unit: 'usage', unit_note: '   ' }), legacy),
    ).toBe('Usage-based, monthly billing')
  })

  // A response cached before the field existed still has the server's string,
  // and that is better than a blank.
  it('falls back to the key specification when there is no billing object', () => {
    expect(formatBillingBasis(undefined, legacy)).toBe('Per month')
  })
})

describe('isAnnualOnly', () => {
  it('is true only for a vendor that sells nothing but yearly contracts', () => {
    expect(isAnnualOnly(billing({ period: 'annual' }))).toBe(true)
    expect(isAnnualOnly(billing({ period: 'monthly' }))).toBe(false)
    // A monthly product with a cheaper yearly option is still sold monthly.
    expect(
      isAnnualOnly(billing({ period: 'monthly', annual_price_minor: 1500 })),
    ).toBe(false)
    expect(isAnnualOnly(undefined)).toBe(false)
  })
})

describe('isPerUser', () => {
  it('reads the unit, not any words', () => {
    expect(isPerUser(billing({ unit: 'per_user' }))).toBe(true)
    expect(isPerUser(billing({ unit: 'flat' }))).toBe(false)
    expect(
      isPerUser(billing({ unit: 'per_contacts', unit_note: 'per user' })),
    ).toBe(false)
    expect(isPerUser(undefined)).toBe(false)
  })
})

describe('annual option', () => {
  const price = { amount_minor: 2000 }

  it('gives the annual figure and the monthly saving when the vendor sells both', () => {
    const both = billing({ period: 'monthly', annual_price_minor: 1500 })
    expect(annualPriceMinor(both)).toBe(1500)
    expect(annualSaving(price, both)).toBe(500)
  })

  it('has nothing to say when there is no annual rate', () => {
    expect(annualPriceMinor(billing())).toBeNull()
    expect(annualSaving(price, billing())).toBeNull()
    expect(annualSaving(price, undefined)).toBeNull()
  })

  // An annual-only product's price already is the yearly rate; there is no
  // second figure to offer, whatever the data happens to carry.
  it('ignores an annual figure on a product that is not sold monthly', () => {
    const annualOnly = billing({ period: 'annual', annual_price_minor: 1500 })
    expect(annualPriceMinor(annualOnly)).toBeNull()
    expect(annualSaving(price, annualOnly)).toBeNull()
  })

  it('does not call a dearer or identical yearly rate a saving', () => {
    expect(
      annualSaving(price, billing({ annual_price_minor: 2000 })),
    ).toBeNull()
    expect(
      annualSaving(price, billing({ annual_price_minor: 2500 })),
    ).toBeNull()
  })
})
