import { describe, expect, it } from 'vitest'

import { hasPriceRecord, priceRecordRows } from './priceHistory'
import type { PriceRecordEntry } from './schemas'

function entry(overrides: Partial<PriceRecordEntry> = {}): PriceRecordEntry {
  return {
    observed_at: '2026-09-02T09:28:00Z',
    price_minor: 3900,
    currency: 'USD',
    billing: {
      period: 'monthly',
      unit: 'flat',
      unit_note: null,
      annual_price_minor: 2900,
    },
    is_current: true,
    ...overrides,
  }
}

describe('priceRecordRows', () => {
  it('prints each figure with its basis and its change against the older one', () => {
    const rows = priceRecordRows([
      entry(),
      entry({
        observed_at: '2026-08-21T11:05:00Z',
        price_minor: 2900,
        // Revisions published before the billing columns existed state no
        // basis. The row says nothing rather than borrowing today's.
        billing: null,
        note: 'Price read from the vendor pricing page on 2026-08-21.',
        is_current: false,
      }),
    ])

    expect(rows).toHaveLength(2)
    expect(rows[0]).toMatchObject({
      observedOn: 'Sep 2, 2026',
      price: '$39.00',
      basis: 'Flat rate, monthly billing',
      note: null,
      change: { direction: 'up', amount: '$10.00' },
      isCurrent: true,
    })
    expect(rows[1]).toMatchObject({
      observedOn: 'Aug 21, 2026',
      price: '$29.00',
      basis: null,
      note: 'Price read from the vendor pricing page on 2026-08-21.',
      // Nothing older to compare the oldest figure against.
      change: null,
      isCurrent: false,
    })
  })

  it('reports a price that came down as a fall of the same size', () => {
    const rows = priceRecordRows([
      entry({ price_minor: 1200 }),
      entry({
        observed_at: '2026-08-18T09:00:00Z',
        price_minor: 2000,
        is_current: false,
      }),
    ])

    expect(rows[0]?.change).toEqual({ direction: 'down', amount: '$8.00' })
  })

  // Subtracting a dollar figure from a euro one produces a number that means
  // nothing. The record still shows both prices; it just claims no delta.
  it('claims no change across a currency change', () => {
    const rows = priceRecordRows([
      entry({ price_minor: 3900, currency: 'USD' }),
      entry({
        observed_at: '2026-08-21T11:05:00Z',
        price_minor: 3900,
        currency: 'EUR',
        is_current: false,
      }),
    ])

    expect(rows[0]?.change).toBeNull()
    expect(rows[1]?.price).toBe('€39.00')
  })

  it('has nothing to say about a product with no record', () => {
    expect(priceRecordRows(undefined)).toEqual([])
    expect(priceRecordRows([])).toEqual([])
  })
})

describe('hasPriceRecord', () => {
  // One dated figure is not a history: every product in the catalog has one.
  it('is true only where a second figure exists to compare against', () => {
    expect(hasPriceRecord(undefined)).toBe(false)
    expect(hasPriceRecord([])).toBe(false)
    expect(hasPriceRecord([entry()])).toBe(false)
    expect(
      hasPriceRecord([
        entry(),
        entry({ price_minor: 2900, is_current: false }),
      ]),
    ).toBe(true)
  })
})
