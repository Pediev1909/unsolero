import type { ProductDetail } from './schemas'

/**
 * What a comparison row is, and what it knows about itself.
 *
 * The old table rendered eight rows of bare numbers — "82/100" against
 * "80/100" — and left the reader to work out whether two points meant
 * anything. Nielsen Norman's finding on comparison tables is that the usual
 * failure is not the design but the content: values without context. So each
 * row now carries what it measures, and each score carries a plain-English
 * sense of the number rather than a fraction.
 */
export interface ComparisonRow {
  key: string
  label: string
  /** One line saying what this row measures. Shown under the label. */
  explains?: string
  /** Grouped under a heading so the table can be read in sections. */
  group: 'Money' | 'Suitability' | 'Judgement'
  /** A score row draws a bar; a fact row prints its value. */
  kind: 'score' | 'fact'
  /** For score rows: the 0–100 number. */
  score?: (product: ProductDetail) => number
  /** For fact rows: the text. */
  value?: (product: ProductDetail) => string
  /** A row no selected product can answer is left out entirely. */
  applies?: (products: ProductDetail[]) => boolean
}

/**
 * Turns a 0–100 score into words.
 *
 * A number out of a hundred looks precise and communicates nothing: nobody
 * knows whether 74 is good. These bands are deliberately coarse, because the
 * underlying judgement is coarse, and a five-band scale stops a two-point gap
 * reading as a meaningful difference when it is not.
 */
export function scoreBand(value: number): string {
  if (value >= 90) return 'Exceptional'
  if (value >= 78) return 'Strong'
  if (value >= 65) return 'Adequate'
  if (value >= 50) return 'Limited'
  return 'Weak'
}

/** The date the price on this product was last read, if it is recorded. */
export function priceCheckedOn(product: ProductDetail): string | null {
  const record = product.evidence.find((item) => item.fact_key === 'price')
  if (!record) return null
  return new Intl.DateTimeFormat('en-US', { dateStyle: 'medium' }).format(
    new Date(record.observed_at),
  )
}

/** The vendor page a price came from, if one is recorded. */
export function priceSource(product: ProductDetail): string | null {
  return (
    product.evidence.find((item) => item.fact_key === 'price')?.source_url ??
    null
  )
}

function insightLine(items: { label: string }[], limit = 2): string {
  if (items.length === 0) return '—'
  return items
    .slice(0, limit)
    .map((item) => item.label)
    .join(', ')
}

export const comparisonRows: ComparisonRow[] = [
  {
    key: 'price',
    group: 'Money',
    kind: 'fact',
    label: 'Entry price',
    explains: 'The cheapest paid tier, as the vendor lists it.',
    value: (product) =>
      product.price.amount_minor === 0
        ? 'No monthly fee'
        : new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: product.price.currency,
          }).format(product.price.amount_minor / 100),
  },
  {
    key: 'billing',
    group: 'Money',
    kind: 'fact',
    // The single most useful row on the page and the one no competitor shows.
    // A monthly rate set against an annual one is not a comparison, and the
    // difference is usually 20–25% — larger than most of the score gaps people
    // agonise over.
    label: 'Billing basis',
    explains: 'Monthly and annual rates are not comparable. This says which.',
    value: (product) => product.key_specification.value || '—',
  },
  {
    key: 'price_checked',
    group: 'Money',
    kind: 'fact',
    label: 'Price read on',
    explains: 'The day we opened the vendor page and read this number.',
    value: (product) => priceCheckedOn(product) ?? 'Not recorded',
  },
  {
    key: 'category',
    group: 'Money',
    kind: 'fact',
    label: 'Category',
    value: (product) => product.category.name,
    // Only worth a row when the products are not all from one category, which
    // is the usual case and would make it a row of identical values.
    applies: (products) =>
      new Set(products.map((product) => product.category.slug)).size > 1,
  },
  {
    key: 'quality',
    group: 'Suitability',
    kind: 'score',
    label: 'Product quality',
    explains: 'How complete and well-made the tool is for its own job.',
    score: (product) => product.scores.quality,
  },
  {
    key: 'value',
    group: 'Suitability',
    kind: 'score',
    label: 'Value for money',
    explains: 'What the price buys relative to the alternatives here.',
    score: (product) => product.scores.value,
  },
  {
    key: 'beginner',
    group: 'Suitability',
    kind: 'score',
    label: 'Easy to adopt',
    explains: 'How much work before it is useful, with nobody to configure it.',
    score: (product) => product.scores.beginner,
  },
  {
    key: 'advanced',
    group: 'Suitability',
    kind: 'score',
    label: 'Depth for power users',
    explains: 'How far it goes once the obvious features run out.',
    score: (product) => product.scores.advanced,
  },
  {
    key: 'durability',
    group: 'Suitability',
    kind: 'score',
    label: 'Vendor stability',
    explains:
      'How likely the company is to still be running this in five years.',
    score: (product) => product.scores.durability,
  },
  {
    key: 'portability',
    group: 'Suitability',
    kind: 'score',
    label: 'Data portability',
    explains: 'How easily your data comes back out if you leave.',
    score: (product) => product.scores.portability,
  },
  {
    key: 'strengths',
    group: 'Judgement',
    kind: 'fact',
    label: 'Strongest at',
    value: (product) => insightLine(product.strengths),
    applies: (products) => products.some((p) => p.strengths.length > 0),
  },
  {
    key: 'weaknesses',
    group: 'Judgement',
    kind: 'fact',
    label: 'Weakest at',
    // Named plainly rather than softened. A comparison that only lists what
    // each option is good at is a brochure for all of them.
    value: (product) => insightLine(product.weaknesses),
    applies: (products) => products.some((p) => p.weaknesses.length > 0),
  },
  {
    key: 'use_cases',
    group: 'Judgement',
    kind: 'fact',
    label: 'Best suited to',
    value: (product) => insightLine(product.use_cases),
    applies: (products) => products.some((p) => p.use_cases.length > 0),
  },
]

/**
 * True when every selected product answers this row identically.
 *
 * Used to fade those rows and to power the "differences only" filter. Nielsen
 * Norman's guidance is that a table must let the reader skim the differences;
 * a row where all four products say the same thing is the noise that hides
 * them.
 */
export function rowIsIdentical(
  row: ComparisonRow,
  products: ProductDetail[],
): boolean {
  if (products.length < 2) return false
  const values = products.map((product) =>
    row.kind === 'score'
      ? scoreBand(row.score?.(product) ?? 0)
      : (row.value?.(product) ?? ''),
  )
  return values.every((value) => value === values[0])
}

export function visibleComparisonRows(
  products: ProductDetail[],
): ComparisonRow[] {
  return comparisonRows.filter((row) => !row.applies || row.applies(products))
}
