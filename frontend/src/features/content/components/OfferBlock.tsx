import { ExternalLink } from 'lucide-react'
import { Link } from 'react-router-dom'

import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { Skeleton } from '../../../components/ui/Skeleton'
import { affiliateClickPath } from '../../analytics/tracking'
import { useOffers } from '../../catalog/queries'
import type { ProductSummary } from '../../catalog/schemas'
import { affiliateDisclosure } from '../model'

interface OfferBlockProps {
  /** The catalog product slug the block names. */
  slug: string
  /**
   * The product's summary, from the piece's related products. Only the name
   * is read from it; when the piece does not list the product, the name is
   * recovered from the slug rather than fetched, because a second request
   * for one word is not worth a second loading state.
   */
  product?: ProductSummary
  heading?: string
  text?: string
  label?: string
}

/**
 * A vendor exit for one catalog product, inside the article.
 *
 * The block carries a product slug and nothing else that could become a URL.
 * The price, the date it was read and the tracked destination all come from
 * the product's live offer — the same query the product page runs, so the
 * two can never disagree — and when there is no live offer the block says so
 * with a plain link into the catalog instead of a button and a price we do
 * not have.
 *
 * Styled like the `cta` block: part of the writing, not a banner. It says in
 * its own text that it pays us and looks the same whether or not it does.
 */
export function OfferBlock({
  slug,
  product,
  heading,
  text,
  label,
}: OfferBlockProps) {
  const offers = useOffers(slug)
  const name = product?.name ?? nameFromSlug(slug)
  const offer = offers.data?.[0]
  const purchasePath = offer?.purchase_path

  return (
    <aside className="border border-ink/15 bg-paper px-5 py-6 sm:px-7">
      {heading && <h3 className="font-semibold text-ink">{heading}</h3>}
      {text && <p className="mt-2 text-base leading-7">{text}</p>}

      {offers.isPending ? (
        <div
          aria-label={`Loading the vendor link for ${name}`}
          className="mt-4"
          role="status"
        >
          <Skeleton className="h-11 min-h-0 w-full" />
        </div>
      ) : offer && purchasePath ? (
        <>
          <div className="mt-4 grid gap-4 border-t border-ink/15 pt-4 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
            <div className="min-w-0">
              <Link
                className="font-display text-lg font-medium leading-tight tracking-[-0.03em] text-ink hover:text-bronze-dark"
                to={`/products/${slug}`}
              >
                {name}
              </Link>
              <p className="mt-1 flex flex-wrap items-baseline gap-x-3 gap-y-1 text-xs text-ink/65">
                <PriceDisplay
                  amountMinor={offer.price.amount_minor}
                  currency={offer.price.currency}
                  size="sm"
                />
                <span>
                  {/* The date was always in the data. The word is what tells
                      a reader whether it is recent. */}
                  {offer.freshness_status === 'stale'
                    ? 'Price last read'
                    : 'Checked'}{' '}
                  {checkedDate(offer.last_checked_at)}
                </span>
              </p>
            </div>
            <a
              className="inline-flex min-h-11 items-center justify-center gap-2 bg-charcoal px-5 text-sm font-semibold text-canvas hover:bg-ink"
              href={affiliateClickPath(purchasePath, 'promotion')}
              rel="nofollow noopener sponsored"
              target="_blank"
            >
              {label || `View at ${offer.merchant.name}`}
              <ExternalLink aria-hidden="true" size={14} />
            </a>
          </div>
          <p className="mt-3 text-xs text-ink/65">{affiliateDisclosure}</p>
        </>
      ) : (
        // No live offer, or the lookup failed: the same honest state for both.
        // The editor's paragraph still stands, and the reader still gets a way
        // to the product — just not a button that implies a deal, or a price
        // we have not read today.
        <p className="mt-4 text-base leading-7">
          <Link
            className="underline decoration-ink/25 underline-offset-4 hover:decoration-ink"
            to={`/products/${slug}`}
          >
            See {name} in the catalog
          </Link>
        </p>
      )}
    </aside>
  )
}

function checkedDate(value: string) {
  return new Intl.DateTimeFormat('en-US', { dateStyle: 'medium' }).format(
    new Date(value),
  )
}

function nameFromSlug(slug: string) {
  return slug
    .split('-')
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}
