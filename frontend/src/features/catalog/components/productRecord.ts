import { visibleEvidence } from '../evidence'
import type { ProductDetail } from '../schemas'

export interface EvidenceSummary {
  count: number
  medianConfidence: number
}

/**
 * How much stands behind the page, in one line: how many visible facts, and
 * the median of their confidence scores.
 *
 * Median rather than mean, because a mean lets one confident fact paper over
 * a doubtful one. Counted over the evidence the page actually shows, so the
 * number in the strip is the number of cards in the record it links to.
 */
export function evidenceSummary(
  product: ProductDetail,
): EvidenceSummary | null {
  const confidences = visibleEvidence(product)
    .map((entry) => entry.confidence)
    .sort((left, right) => left - right)
  if (confidences.length === 0) return null
  const middle = Math.floor(confidences.length / 2)
  const upper = confidences[middle] ?? 0
  const lower = confidences[middle - 1] ?? upper
  const median =
    confidences.length % 2 === 1 ? upper : Math.round((lower + upper) / 2)
  return { count: confidences.length, medianConfidence: median }
}

/**
 * The most recent day any visible fact was observed, or null when nothing is
 * dated. The revision line in the evidence record prints it: a revision id
 * says which record this is, the date says how old.
 *
 * This is deliberately not a price history. The API returns the observations
 * attached to the published fact revision only, and a price correction opens a
 * new revision without carrying the obsolete price observation forward, so
 * each fact arrives with one date. Earlier prices exist in superseded
 * revisions the API does not expose.
 */
export function latestObservation(product: ProductDetail): string | null {
  const times = visibleEvidence(product)
    .map((entry) => new Date(entry.observed_at).getTime())
    .filter((time) => Number.isFinite(time))
  if (times.length === 0) return null
  return new Intl.DateTimeFormat('en-US', { dateStyle: 'medium' }).format(
    new Date(Math.max(...times)),
  )
}

/** The first eight characters: enough to tell two revisions apart, short enough to print. */
export function shortRevision(id: string): string {
  return id.slice(0, 8)
}
