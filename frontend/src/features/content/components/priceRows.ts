import { useMemo } from 'react'

import type { ProductSummary } from '../../catalog/schemas'

// The hero of a comparison used to be an abstract illustration: decoration in
// the one place a reader most wants an answer. This draws the comparison
// itself — the products the piece covers, at their catalog prices, on one
// shared scale, cheapest first. The length of each bar against the others is
// the finding, and it differs on every page because the data does.
//
// One row per product rather than pins on a horizontal axis: pins collide as
// soon as two prices are close, and every fix for that either stacks labels
// into space this hero does not have or turns a scale into a chart.
//
// Nothing is authored per article. Every number comes from the catalog, so it
// cannot drift away from the product pages this piece links to.

type Row = {
  id: string
  name: string
  brand: string
  amountMinor: number
  currency: string
  width: number
  cheapest: boolean
  free: boolean
}

export function usePriceRows(products: ProductSummary[]) {
  return useMemo<{ rows: Row[]; currency: string } | null>(() => {
    const priced = products.filter((product) => product.price?.currency)
    // One product is not a comparison, and a set of identical prices says
    // nothing a bar can show — both are better served by the piece's own words.
    if (priced.length < 2) return null

    const currency = priced[0]?.price.currency
    if (!currency) return null
    if (priced.some((product) => product.price.currency !== currency))
      return null

    const amounts = priced.map((product) => product.price.amount_minor)
    const highest = Math.max(...amounts)
    const cheapest = Math.min(...amounts)
    if (highest === cheapest || highest === 0) return null

    const rows = [...priced]
      .sort((left, right) => left.price.amount_minor - right.price.amount_minor)
      .map((product) => ({
        id: product.id,
        name: product.name,
        brand: product.brand.name,
        amountMinor: product.price.amount_minor,
        currency,
        // A free tier still needs a visible mark, or the row reads as missing
        // data rather than as zero.
        width: Math.max((product.price.amount_minor / highest) * 100, 1.5),
        cheapest: product.price.amount_minor === cheapest,
        free: product.price.amount_minor === 0,
      }))
    return { rows, currency }
  }, [products])
}
