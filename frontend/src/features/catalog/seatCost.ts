import { isPerUser } from './billing'
import type { ProductSummary } from './schemas'

/**
 * Whether a listed price is charged once per seat or once for the whole team.
 * Read from the API's structured billing unit, which is the only place the
 * catalog states a basis. Neither the key-specification phrase nor the
 * description prose is consulted — a calculator that guesses a basis invents a
 * price — and a product with no billing object is treated as flat, so nothing
 * gets multiplied on a guess.
 */
export type BillingBasis = 'per_seat' | 'flat'

export const teamSizeRange = { min: 1, max: 50 } as const

export type PricedProduct = Pick<
  ProductSummary,
  'id' | 'name' | 'slug' | 'price' | 'billing'
>

export interface SeatCostLine<T extends PricedProduct = PricedProduct> {
  product: T
  basis: BillingBasis
  /** The listed price, in minor units, before any multiplication. */
  unitMinor: number
  /** What the team pays a month, in minor units. */
  totalMinor: number
  currency: string
}

export function billingBasis(
  product: Pick<ProductSummary, 'billing'>,
): BillingBasis {
  return isPerUser(product.billing) ? 'per_seat' : 'flat'
}

export function hasPerSeatPricing(
  products: readonly Pick<ProductSummary, 'billing'>[],
): boolean {
  return products.some((product) => billingBasis(product) === 'per_seat')
}

/** Whole seats between one and fifty; anything unreadable is one seat. */
export function clampTeamSize(value: number): number {
  if (!Number.isFinite(value)) return teamSizeRange.min
  return Math.min(
    teamSizeRange.max,
    Math.max(teamSizeRange.min, Math.trunc(value)),
  )
}

/**
 * Each product's monthly total for a team, cheapest first. Per-seat prices are
 * multiplied by the seat count; flat prices are carried through unchanged.
 * Everything stays in integer minor units — an integer times an integer needs
 * no rounding, and formatting is left to the display layer.
 */
export function seatCostLines<T extends PricedProduct>(
  products: readonly T[],
  teamSize: number,
): SeatCostLine<T>[] {
  const seats = clampTeamSize(teamSize)
  return products
    .map((product): SeatCostLine<T> => {
      const basis = billingBasis(product)
      const unitMinor = product.price.amount_minor
      return {
        product,
        basis,
        unitMinor,
        totalMinor: basis === 'per_seat' ? unitMinor * seats : unitMinor,
        currency: product.price.currency,
      }
    })
    .sort(
      (left, right) =>
        left.totalMinor - right.totalMinor ||
        left.product.name.localeCompare(right.product.name),
    )
}
