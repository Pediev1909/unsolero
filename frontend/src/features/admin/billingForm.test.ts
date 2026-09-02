import { describe, expect, it } from 'vitest'

import {
  fromBilling,
  parseBillingForm,
  unitNoteMaxLength,
  type BillingFormValues,
} from './billingForm'

function values(overrides: Partial<BillingFormValues> = {}): BillingFormValues {
  return {
    billing_period: 'monthly',
    pricing_unit: 'flat',
    unit_note: '',
    annual_price_minor: '',
    ...overrides,
  }
}

describe('parseBillingForm', () => {
  it('turns the controls into the object the API takes', () => {
    expect(
      parseBillingForm(
        values({
          pricing_unit: 'per_user',
          unit_note: '  Minimum 3 seats  ',
          annual_price_minor: ' 1500 ',
        }),
        2000,
      ),
    ).toEqual({
      ok: true,
      billing: {
        period: 'monthly',
        unit: 'per_user',
        unit_note: 'Minimum 3 seats',
        annual_price_minor: 1500,
      },
    })
  })

  it('sends an empty note and an empty annual rate as null, not as empty strings', () => {
    const result = parseBillingForm(values(), 2000)
    expect(result.ok).toBe(true)
    if (result.ok) {
      expect(result.billing.unit_note).toBeNull()
      expect(result.billing.annual_price_minor).toBeNull()
    }
  })

  // The yearly rate is a second price beside a monthly one. On any other
  // period there is nothing for it to sit beside.
  it('refuses a yearly rate on a product not sold month to month', () => {
    const result = parseBillingForm(
      values({ billing_period: 'annual', annual_price_minor: '1500' }),
      2000,
    )
    expect(result.ok).toBe(false)
    if (!result.ok) {
      expect(result.errors.annual_price_minor).toMatch(/month to month/)
      expect(result.errors.billing_period).toBeUndefined()
    }
  })

  it('refuses a free tier priced above zero, and allows one at zero', () => {
    const priced = parseBillingForm(values({ billing_period: 'free' }), 2000)
    expect(priced.ok).toBe(false)
    if (!priced.ok) expect(priced.errors.billing_period).toMatch(/price of 0/)

    expect(parseBillingForm(values({ billing_period: 'free' }), 0).ok).toBe(
      true,
    )
    // With no price to check against, the rule cannot fire.
    expect(parseBillingForm(values({ billing_period: 'free' }), null).ok).toBe(
      true,
    )
  })

  it('refuses a yearly rate that is not whole minor units', () => {
    for (const bad of ['15.00', '-1500', 'abc', '1 500']) {
      const result = parseBillingForm(values({ annual_price_minor: bad }), 2000)
      expect(result.ok).toBe(false)
      if (!result.ok) expect(result.errors.annual_price_minor).toMatch(/minor/)
    }
  })

  it('caps the unit note at the length the server allows', () => {
    const result = parseBillingForm(
      values({ unit_note: 'x'.repeat(unitNoteMaxLength + 1) }),
      2000,
    )
    expect(result.ok).toBe(false)
    if (!result.ok) expect(result.errors.unit_note).toMatch(/120/)
    expect(
      parseBillingForm(
        values({ unit_note: 'x'.repeat(unitNoteMaxLength) }),
        2000,
      ).ok,
    ).toBe(true)
  })

  it('rejects a period or unit outside the vocabulary', () => {
    const result = parseBillingForm(
      values({
        billing_period: 'weekly' as BillingFormValues['billing_period'],
        pricing_unit: 'per_seat' as BillingFormValues['pricing_unit'],
      }),
      2000,
    )
    expect(result.ok).toBe(false)
    if (!result.ok) {
      expect(result.errors.billing_period).toBe('Choose a billing period.')
      expect(result.errors.pricing_unit).toBe('Choose a pricing unit.')
    }
  })
})

describe('fromBilling', () => {
  it('starts a product without the object as a monthly flat rate', () => {
    expect(fromBilling(undefined)).toEqual(values())
  })

  it('round-trips a product that has one', () => {
    const billing = {
      period: 'monthly' as const,
      unit: 'per_contacts' as const,
      unit_note: 'Up to 500 contacts',
      annual_price_minor: 1300,
    }
    const form = fromBilling(billing)
    expect(form).toEqual({
      billing_period: 'monthly',
      pricing_unit: 'per_contacts',
      unit_note: 'Up to 500 contacts',
      annual_price_minor: '1300',
    })
    expect(parseBillingForm(form, 2000)).toEqual({ ok: true, billing })
  })
})
