import { describe, expect, it } from 'vitest'

import type { Billing } from './schemas'
import {
  billingBasis,
  clampTeamSize,
  hasPerSeatPricing,
  seatCostLines,
  type PricedProduct,
} from './seatCost'

const perUser: Billing = {
  period: 'monthly',
  unit: 'per_user',
  unit_note: null,
  annual_price_minor: null,
}
const flatRate: Billing = { ...perUser, unit: 'flat' }

function product(
  name: string,
  amountMinor: number,
  billing: Billing | undefined,
  currency = 'USD',
): PricedProduct {
  return {
    id: name.toLowerCase(),
    name,
    slug: name.toLowerCase(),
    price: { amount_minor: amountMinor, currency },
    billing,
  }
}

describe('billingBasis', () => {
  it('reads per-seat pricing from the structured unit', () => {
    expect(billingBasis(product('A', 1, perUser))).toBe('per_seat')
    expect(billingBasis(product('B', 1, flatRate))).toBe('flat')
    // A per-seat price on a yearly contract is still charged per seat.
    expect(
      billingBasis(product('C', 1, { ...perUser, period: 'annual' })),
    ).toBe('per_seat')
  })

  // Words are not a basis. A note that happens to say "per user" on a
  // contact-tier price does not make it one, and a product the API has not
  // yet described is not multiplied on a guess.
  it('never infers a basis from words or from silence', () => {
    expect(
      billingBasis(
        product('D', 1, {
          ...flatRate,
          unit: 'per_contacts',
          unit_note: 'per user allowance',
        }),
      ),
    ).toBe('flat')
    expect(billingBasis(product('E', 1, undefined))).toBe('flat')
  })
})

describe('seatCostLines', () => {
  const perSeat = product('Seat CRM', 1990, perUser)
  const flat = product('Flat Suite', 2500, flatRate)
  const cheapSeat = product('Lite CRM', 400, perUser)

  it('multiplies per-seat prices by the team and leaves flat prices alone', () => {
    const lines = seatCostLines([perSeat, flat, cheapSeat], 5)
    expect(lines.map((line) => [line.product.name, line.totalMinor])).toEqual([
      ['Lite CRM', 2000],
      ['Flat Suite', 2500],
      ['Seat CRM', 9950],
    ])
    expect(lines.find((line) => line.product === flat)?.basis).toBe('flat')
    expect(lines.find((line) => line.product === flat)?.unitMinor).toBe(2500)
  })

  it('keeps every total an integer in minor units, whatever the team size input', () => {
    for (const line of seatCostLines([perSeat, cheapSeat], 7.9)) {
      expect(Number.isSafeInteger(line.totalMinor)).toBe(true)
    }
    // 7.9 seats is seven seats; a fraction of a seat is not a thing anyone
    // is billed for.
    expect(seatCostLines([perSeat], 7.9)[0]?.totalMinor).toBe(1990 * 7)
  })

  it('carries each product currency through untouched', () => {
    const euro = product('Euro Seat', 1000, perUser, 'EUR')
    const lines = seatCostLines([perSeat, euro], 2)
    expect(lines.find((line) => line.product === euro)?.currency).toBe('EUR')
    expect(lines.find((line) => line.product === perSeat)?.currency).toBe('USD')
  })

  it('orders ties by name so the list is stable', () => {
    const lines = seatCostLines(
      [product('Zed', 1000, flatRate), product('Alpha', 1000, flatRate)],
      3,
    )
    expect(lines.map((line) => line.product.name)).toEqual(['Alpha', 'Zed'])
  })
})

describe('team size', () => {
  it('clamps to whole seats between one and fifty', () => {
    expect(clampTeamSize(0)).toBe(1)
    expect(clampTeamSize(-3)).toBe(1)
    expect(clampTeamSize(51)).toBe(50)
    expect(clampTeamSize(12.7)).toBe(12)
    expect(clampTeamSize(Number.NaN)).toBe(1)
  })

  it('reports whether a list has anything worth multiplying', () => {
    expect(hasPerSeatPricing([product('A', 1, flatRate)])).toBe(false)
    expect(
      hasPerSeatPricing([product('A', 1, flatRate), product('B', 1, perUser)]),
    ).toBe(true)
    expect(hasPerSeatPricing([product('A', 1, undefined)])).toBe(false)
    expect(hasPerSeatPricing([])).toBe(false)
  })
})
