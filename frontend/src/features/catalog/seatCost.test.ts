import { describe, expect, it } from 'vitest'

import {
  billingBasis,
  clampTeamSize,
  hasPerSeatPricing,
  seatCostLines,
  type PricedProduct,
} from './seatCost'

function product(
  name: string,
  amountMinor: number,
  specification: string,
  currency = 'USD',
): PricedProduct {
  return {
    id: name.toLowerCase(),
    name,
    slug: name.toLowerCase(),
    price: { amount_minor: amountMinor, currency },
    key_specification: { label: 'Billing', value: specification },
  }
}

describe('billingBasis', () => {
  it('reads per-seat pricing only from the structured specification', () => {
    expect(billingBasis(product('A', 1, 'Per user per month'))).toBe('per_seat')
    expect(billingBasis(product('B', 1, '12 USD per seat, monthly'))).toBe(
      'per_seat',
    )
    expect(billingBasis(product('C', 1, 'Per month'))).toBe('flat')
    // Substrings are not a basis: "supervisor" contains "per" and "user" is
    // in "users included", and neither says anything about seats.
    expect(billingBasis(product('D', 1, 'Supervisor dashboard'))).toBe('flat')
    expect(billingBasis(product('E', 1, '3 users included'))).toBe('flat')
  })
})

describe('seatCostLines', () => {
  const perSeat = product('Seat CRM', 1990, 'Per user per month')
  const flat = product('Flat Suite', 2500, 'Per month')
  const cheapSeat = product('Lite CRM', 400, 'Per seat per month')

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
    const euro = product('Euro Seat', 1000, 'Per user per month', 'EUR')
    const lines = seatCostLines([perSeat, euro], 2)
    expect(lines.find((line) => line.product === euro)?.currency).toBe('EUR')
    expect(lines.find((line) => line.product === perSeat)?.currency).toBe('USD')
  })

  it('orders ties by name so the list is stable', () => {
    const lines = seatCostLines(
      [product('Zed', 1000, 'Per month'), product('Alpha', 1000, 'Per month')],
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
    expect(hasPerSeatPricing([product('A', 1, 'Per month')])).toBe(false)
    expect(
      hasPerSeatPricing([
        product('A', 1, 'Per month'),
        product('B', 1, 'Per user per month'),
      ]),
    ).toBe(true)
    expect(hasPerSeatPricing([])).toBe(false)
  })
})
