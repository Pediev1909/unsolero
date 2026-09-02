import type { Billing, ProductSummary } from './schemas'

/**
 * Words for the structured billing basis.
 *
 * These are the server's own rules for deriving `key_specification` from the
 * `billing` object, repeated here so a page holding the object and a page
 * holding only the older string print the same phrase. Anything that needs the
 * shape rather than the words — is this per seat, is this yearly-only, is
 * there a cheaper annual rate — asks the predicates below instead of reading
 * the phrase back.
 */

const unitPhrases: Record<Billing['unit'], string> = {
  flat: 'Flat rate',
  per_user: 'Per user',
  per_contacts: 'Per contact tier',
  per_transaction: 'Per transaction',
  usage: 'Usage-based',
}

// Units whose generic phrase gives way to the vendor's own wording when the
// catalog recorded one: "Up to 500 contacts" says more than "Per contact tier".
const notedUnits: ReadonlySet<Billing['unit']> = new Set([
  'per_contacts',
  'per_transaction',
  'usage',
])

const periodPhrases: Record<Billing['period'], string | null> = {
  monthly: 'monthly billing',
  annual: 'billed yearly',
  free: 'free plan',
  usage: null,
}

function unitPhrase(billing: Billing): string {
  const note = billing.unit_note?.trim()
  if (note && notedUnits.has(billing.unit)) return upperFirst(note)
  return unitPhrases[billing.unit]
}

// The note is recorded as it would read mid-sentence ("at 1,000 contacts")
// and opens the phrase here, so its first character is raised the way the
// server raises it. Spreading the string walks code points, not UTF-16 units,
// so a note that opens with a symbol is not torn in half.
function upperFirst(value: string): string {
  const [first = ''] = value
  return first.toUpperCase() + value.slice(first.length)
}

/**
 * The basis as one phrase — "Per user, monthly billing" — or, when the API
 * response predates the structured field, whatever the server printed.
 */
export function formatBillingBasis(
  billing: Billing | undefined,
  keySpecification: Pick<ProductSummary['key_specification'], 'value'>,
): string {
  if (!billing) return keySpecification.value
  return [unitPhrase(billing), periodPhrases[billing.period]]
    .filter((part): part is string => part !== null)
    .join(', ')
}

/** The vendor sells only yearly contracts; `price` is the per-month equivalent. */
export function isAnnualOnly(billing: Billing | undefined): boolean {
  return billing?.period === 'annual'
}

/** The listed price is charged once per seat rather than once for the team. */
export function isPerUser(billing: Billing | undefined): boolean {
  return billing?.unit === 'per_user'
}

/**
 * The per-month figure on annual billing, where the vendor also sells month
 * to month. Null on every other basis: an annual-only product has no second
 * price, and a free or usage-priced one has no yearly contract to quote.
 */
export function annualPriceMinor(billing: Billing | undefined): number | null {
  if (!billing || billing.period !== 'monthly') return null
  return billing.annual_price_minor
}

/**
 * What a month costs less on the yearly contract, in minor units. Null when
 * there is no annual option, and also when the annual figure is not lower —
 * a "saving" of nothing, or of less than nothing, is not one.
 */
export function annualSaving(
  price: Pick<ProductSummary['price'], 'amount_minor'>,
  billing: Billing | undefined,
): number | null {
  const annual = annualPriceMinor(billing)
  if (annual === null) return null
  const saving = price.amount_minor - annual
  return saving > 0 ? saving : null
}
