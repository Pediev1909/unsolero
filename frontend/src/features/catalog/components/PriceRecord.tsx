import { ArrowDown, ArrowUp } from 'lucide-react'

import { cn } from '../../../lib/styles/cn'
import { hasPriceRecord, priceRecordRows } from '../priceHistory'
import type { PriceChange } from '../priceHistory'
import type { PriceRecordEntry } from '../schemas'
import { productSectionIDs, sectionAnchorClass } from './productSections'

/**
 * What this product has cost, and when it changed.
 *
 * Dated prices are this site's claim on a reader's trust, and a competitor
 * survey found none of them publishes a price history at all — the number on
 * their page is the number today, with no way to tell whether it moved last
 * week. The record is the evidence for the date beside the price: two figures,
 * both read from the vendor's own page, both kept.
 *
 * It draws nothing at all where there is one figure or none, which is most of
 * the catalog. A one-row "history" is a decoration, not a record, and the
 * honest empty state for a price that has never moved is silence — the date it
 * was read is already printed beside the price.
 *
 * The change column is ink, not red and green. A price rising is bad news for
 * the reader and good news for nobody, and colouring it like a stock ticker
 * would editorialise a fact the arrow already states.
 */
export function PriceRecord({
  record,
}: {
  record: PriceRecordEntry[] | undefined
}) {
  // The same predicate the page asks before putting this in the jump row, so
  // an anchor can never point at a section that decided not to draw.
  if (!hasPriceRecord(record)) return null
  const rows = priceRecordRows(record)

  return (
    <section
      aria-labelledby="price-record-heading"
      className={cn('mt-10', sectionAnchorClass)}
      id={productSectionIDs.priceRecord}
    >
      <h2
        className="text-xs font-bold uppercase tracking-[0.14em]"
        id="price-record-heading"
      >
        Price record
      </h2>
      <div className="mt-4 overflow-x-auto">
        <table className="w-full min-w-[20rem] border-collapse text-sm">
          <caption className="sr-only">
            Every price this product has been listed at, newest first.
          </caption>
          <thead>
            <tr className="border-b border-ink/15 text-left text-[0.625rem] font-bold uppercase tracking-[0.13em] text-ink/65">
              <th className="py-2 pr-4" scope="col">
                Read on
              </th>
              <th className="py-2 pr-4" scope="col">
                Price
              </th>
              <th className="py-2 text-right" scope="col">
                Change
              </th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr className="border-b border-ink/10 align-top" key={row.key}>
                <th
                  className="whitespace-nowrap py-3 pr-4 text-left font-normal text-ink/70"
                  scope="row"
                >
                  {row.observedOn}
                  {row.isCurrent && (
                    <span className="mt-1 block text-[0.625rem] font-bold uppercase tracking-[0.13em] text-bronze-dark">
                      Current
                    </span>
                  )}
                </th>
                <td className="py-3 pr-4">
                  <span className="font-medium">{row.price}</span>
                  {row.basis && (
                    <span className="mt-0.5 block text-xs leading-5 text-ink/65">
                      {row.basis}
                    </span>
                  )}
                  {row.note && (
                    <span className="mt-1 block text-xs leading-5 text-ink/65">
                      {row.note}
                    </span>
                  )}
                </td>
                <td className="whitespace-nowrap py-3 text-right">
                  <PriceChangeCell change={row.change} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <p className="mt-4 text-xs leading-5 text-ink/65">
        Each figure was read from the vendor's own page on the date shown.
      </p>
    </section>
  )
}

function PriceChangeCell({ change }: { change: PriceChange | null }) {
  if (!change) {
    return (
      <span className="text-ink/65">
        <span className="sr-only">No earlier price recorded</span>
        <span aria-hidden="true">—</span>
      </span>
    )
  }
  const Arrow = change.direction === 'up' ? ArrowUp : ArrowDown
  return (
    <span className="inline-flex items-center gap-1 font-medium text-ink/70">
      <span className="sr-only">
        {change.direction === 'up' ? 'Up by' : 'Down by'}
      </span>
      <Arrow aria-hidden="true" size={14} />
      {change.amount}
    </span>
  )
}
