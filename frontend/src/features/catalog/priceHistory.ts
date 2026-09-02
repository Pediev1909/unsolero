import { formatMinorCurrency } from '../../lib/money/format'
import { formatBillingBasis } from './billing'
import type { PriceRecordEntry } from './schemas'

/**
 * The price record, turned into rows a table can print.
 *
 * Every price on this site is dated, and that is the whole differentiator: no
 * competitor in this category publishes what a tool used to cost, only what it
 * costs today. The API keeps the history honest — it collapses revisions that
 * repeat a figure, so a row exists only where the number actually moved — and
 * this module does the arithmetic and the wording, so the component prints
 * strings and decides nothing.
 */

/** How this figure differs from the one below it in the table. */
export interface PriceChange {
  direction: 'up' | 'down'
  /** The absolute difference, formatted in the row's own currency. */
  amount: string
}

export interface PriceRecordRow {
  key: string
  /** The date the figure was read, formatted for display. */
  observedOn: string
  price: string
  /** The billing basis in words, or null where the revision stated none. */
  basis: string | null
  /** The reviewer's sentence, where there is one. */
  note: string | null
  /** Null on the oldest row: there is nothing older to compare it against. */
  change: PriceChange | null
  isCurrent: boolean
}

const observedFormat = new Intl.DateTimeFormat('en-US', { dateStyle: 'medium' })

/**
 * Rows for one product's price record, newest first, each carrying its change
 * against the next-older figure.
 *
 * A change is only computed within one currency. Subtracting a euro price from
 * a dollar one produces a number that means nothing, and a record that spans a
 * currency change is a change of price *and* of unit; the row prints both
 * figures and claims no delta.
 */
export function priceRecordRows(
  record: PriceRecordEntry[] | undefined,
): PriceRecordRow[] {
  if (!record) return []
  return record.map((entry, index) => ({
    key: `${entry.observed_at}-${entry.price_minor}`,
    observedOn: observedFormat.format(new Date(entry.observed_at)),
    price: formatMinorCurrency(entry.price_minor, entry.currency),
    // formatBillingBasis falls back to the key specification when a response
    // carries no structured basis; here there is nothing to fall back to, and
    // an entry that stated no basis must say nothing rather than borrow one.
    basis:
      formatBillingBasis(entry.billing ?? undefined, { value: '' }) || null,
    note: entry.note?.trim() || null,
    change: priceChange(entry, record[index + 1]),
    isCurrent: entry.is_current,
  }))
}

/**
 * Whether this record is worth drawing. One dated figure is not a history —
 * every product has one — so a single entry renders nothing at all rather than
 * a one-row table dressed up as a record.
 */
export function hasPriceRecord(
  record: PriceRecordEntry[] | undefined,
): boolean {
  return (record?.length ?? 0) > 1
}

function priceChange(
  entry: PriceRecordEntry,
  older: PriceRecordEntry | undefined,
): PriceChange | null {
  if (!older || older.currency !== entry.currency) return null
  const difference = entry.price_minor - older.price_minor
  if (difference === 0) return null
  return {
    direction: difference > 0 ? 'up' : 'down',
    amount: formatMinorCurrency(Math.abs(difference), entry.currency),
  }
}
