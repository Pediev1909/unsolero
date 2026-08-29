import { ExternalLink, X } from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router-dom'

import { Button } from '../../../components/ui/Button'
import { Checkbox } from '../../../components/ui/Checkbox'
import { cn } from '../../../lib/styles/cn'
import { MerchantAction } from './MerchantAction'
import {
  priceSource,
  rowIsIdentical,
  scoreBand,
  visibleComparisonRows,
  type ComparisonRow,
} from '../comparisonRows'
import type { ProductDetail } from '../schemas'

interface ComparisonTableProps {
  products: ProductDetail[]
  onRemove: (productID: string) => void
}

/**
 * The comparison, rebuilt around what the numbers mean.
 *
 * It used to print eight rows of bare fractions — 82/100 against 80/100 — and
 * leave the reader to decide whether two points mattered. Nielsen Norman's
 * finding is that the usual failure of a comparison table is the content
 * rather than the design: values with no context. So a score now shows a bar
 * and a word, every row says what it measures, and the rows that decide most
 * purchases — the billing basis and the date the price was read — are here
 * for the first time.
 *
 * No winner is declared anywhere, which is also their recommendation and
 * happens to be the whole position of this site: the reader's priorities
 * decide, and our job is to make the differences impossible to miss.
 */
export function ComparisonTable({ products, onRemove }: ComparisonTableProps) {
  const [differencesOnly, setDifferencesOnly] = useState(false)
  const allRows = visibleComparisonRows(products)
  const rows = differencesOnly
    ? allRows.filter((row) => !rowIsIdentical(row, products))
    : allRows
  const identicalCount = allRows.filter((row) =>
    rowIsIdentical(row, products),
  ).length

  const groups = ['Money', 'Suitability', 'Judgement'] as const

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <Checkbox
          checked={differencesOnly}
          label={`Show only where they differ${identicalCount > 0 ? ` (hides ${identicalCount})` : ''}`}
          onChange={(event) => setDifferencesOnly(event.target.checked)}
        />
        <p className="text-xs text-ink/60">
          Rows where every tool answers the same are tinted.
        </p>
      </div>

      <div
        aria-label="Product comparison"
        className="overflow-x-auto border border-ink/15"
        role="region"
        tabIndex={0}
      >
        <table className="w-full min-w-[48rem] border-collapse text-left text-sm">
          <caption className="sr-only">
            Structured facts and suitability scores for the selected tools
          </caption>
          <thead>
            {/* Sticky, because the whole point of scrolling down a comparison
                is to keep asking "which column is which". */}
            <tr>
              <th className="sticky top-0 left-0 z-30 w-44 border-r border-b border-ink/15 bg-paper p-4 text-xs tracking-[0.12em] text-ink/68 uppercase">
                Compare
              </th>
              {products.map((product) => (
                <th
                  className="sticky top-0 z-20 min-w-52 border-r border-b border-ink/15 bg-surface p-4 align-top last:border-r-0"
                  key={product.id}
                  scope="col"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0">
                      <p className="text-[0.625rem] tracking-[0.12em] text-ink/65 uppercase">
                        {product.brand.name}
                      </p>
                      <Link
                        className="mt-1.5 block font-display text-lg leading-tight font-medium hover:text-bronze-dark"
                        to={`/products/${product.slug}`}
                      >
                        {product.name}
                      </Link>
                      <p className="mt-2 font-display text-base font-semibold tabular-nums">
                        {product.price.amount_minor === 0
                          ? 'No monthly fee'
                          : new Intl.NumberFormat('en-US', {
                              style: 'currency',
                              currency: product.price.currency,
                            }).format(product.price.amount_minor / 100)}
                      </p>
                    </div>
                    <Button
                      aria-label={`Remove ${product.name} from comparison`}
                      onClick={() => onRemove(product.id)}
                      size="sm"
                      variant="quiet"
                    >
                      <X aria-hidden="true" size={15} />
                    </Button>
                  </div>
                </th>
              ))}
            </tr>
          </thead>

          {groups.map((group) => {
            const groupRows = rows.filter((row) => row.group === group)
            if (groupRows.length === 0) return null
            return (
              <tbody key={group}>
                <tr>
                  <th
                    className="border-b border-ink/15 bg-ink p-2.5 pl-4 text-[0.65rem] font-bold tracking-[0.16em] text-canvas uppercase"
                    colSpan={products.length + 1}
                    scope="colgroup"
                  >
                    {group}
                  </th>
                </tr>
                {groupRows.map((row) => (
                  <Row
                    key={row.key}
                    muted={!differencesOnly && rowIsIdentical(row, products)}
                    products={products}
                    row={row}
                  />
                ))}
              </tbody>
            )
          })}

          {/* The comparison used to end on a fact and leave the reader to find
              their own way to the vendor — back to the card, into the product
              page, then out. Three clicks from the moment they had decided.
              This row is the exit, placed where the decision actually happens.

              It is not a recommendation and it names no winner: every column
              that has a verified offer gets the same button, in the same
              place, whether or not the offer pays us. Columns without one show
              nothing rather than a disabled control, because an empty cell
              says "no offer" and a greyed button says "we are hiding
              something". */}
          <tfoot>
            <tr>
              <th
                className="sticky left-0 z-10 w-44 border-r border-t border-ink/15 bg-paper p-4 text-left align-middle text-xs tracking-[0.12em] text-ink/68 uppercase"
                scope="row"
              >
                Go to vendor
              </th>
              {products.map((product) => (
                <td
                  className="min-w-52 border-r border-t border-ink/15 p-4 align-middle last:border-r-0"
                  key={product.id}
                >
                  <MerchantAction
                    className=""
                    compact
                    slug={product.slug}
                    source="comparison"
                  />
                </td>
              ))}
            </tr>
          </tfoot>
        </table>
      </div>

      <p className="mt-4 text-xs leading-5 text-ink/65">
        Prices are read from each vendor&rsquo;s own page on the date shown.
        Scores are our judgement, not customer ratings. The vendor buttons above
        are affiliate links where we have a programme, and some of these tools
        pay us and some do not — neither the scores nor the order of these
        columns is affected by that. We deliberately do not name a winner —
        which of these matters is your decision, not ours.
      </p>
    </div>
  )
}

function Row({
  row,
  products,
  muted,
}: {
  row: ComparisonRow
  products: ProductDetail[]
  muted: boolean
}) {
  // Marked by tint, not by opacity. Dimming the row with opacity dropped its
  // text to a contrast ratio of 2.1 against a required 4.5 — the "these are
  // the same" signal was being paid for in legibility, by the reader who most
  // needs to read it.
  return (
    <tr>
      <th
        className={cn(
          'sticky left-0 z-10 border-r border-b border-ink/15 p-4 align-top font-semibold',
          muted ? 'bg-paper/60' : 'bg-paper',
        )}
        scope="row"
      >
        {row.label}
        {row.explains && (
          <span className="mt-1 block text-xs leading-4 font-normal text-ink/70">
            {row.explains}
          </span>
        )}
      </th>
      {products.map((product) => (
        <td
          className={cn(
            'border-r border-b border-ink/10 p-4 align-top last:border-r-0',
            muted ? 'bg-paper/40' : 'bg-surface',
          )}
          key={product.id}
        >
          {row.kind === 'score' ? (
            <ScoreCell value={row.score?.(product) ?? 0} />
          ) : (
            <FactCell product={product} row={row} />
          )}
        </td>
      ))}
    </tr>
  )
}

/**
 * A score as a bar and a word.
 *
 * The number stays, smaller, because somebody comparing four tools does want
 * to break a tie. But the word is what carries the meaning: nobody knows
 * whether 74 out of 100 is good, and the bands are coarse on purpose so a
 * two-point gap does not read as a difference when it is not one.
 */
function ScoreCell({ value }: { value: number }) {
  return (
    <div>
      <div className="flex items-baseline justify-between gap-2">
        <span className="font-display text-sm font-semibold">
          {scoreBand(value)}
        </span>
        <span className="text-xs text-ink/70 tabular-nums">{value}</span>
      </div>
      <div
        aria-hidden="true"
        className="mt-2 h-1.5 w-full rounded-full bg-line"
      >
        <div
          className="h-1.5 rounded-full bg-bronze"
          style={{ width: `${Math.min(100, Math.max(0, value))}%` }}
        />
      </div>
    </div>
  )
}

function FactCell({
  row,
  product,
}: {
  row: ComparisonRow
  product: ProductDetail
}) {
  const text = row.value?.(product) ?? '—'
  const source = row.key === 'price_checked' ? priceSource(product) : null

  return (
    <div className="text-ink">
      <span
        className={cn(
          row.key === 'price' && 'font-display text-base font-semibold',
        )}
      >
        {text}
      </span>
      {source && (
        <a
          className="mt-1.5 inline-flex items-center gap-1 text-xs text-bronze underline-offset-4 hover:underline"
          href={source}
          rel="nofollow noopener noreferrer"
          target="_blank"
        >
          The page it came from
          <ExternalLink aria-hidden="true" size={12} />
        </a>
      )}
    </div>
  )
}
