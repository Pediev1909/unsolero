import { ExternalLink } from 'lucide-react'
import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { buttonStyles } from '../../../components/ui/buttonStyles'
import { cn } from '../../../lib/styles/cn'
import { affiliateClickPath } from '../../analytics/tracking'
import { priceCheckedOn, priceSource } from '../comparisonRows'
import { useOffers } from '../queries'
import type { ProductDetail } from '../schemas'
import { evidenceSummary } from './productRecord'
import { productSectionIDs, sectionAnchorClass } from './productSections'

/**
 * The first screen's answers: how much, for whom, how well sourced, and where.
 *
 * Every site in this category opens a product page with one strip — price,
 * free plan, best for, a button — and the reason is not fashion: those are the
 * questions a reader arrives with. This page answered them too, but after a
 * scroll, in a price card, a badge row and an evidence record the competition
 * has no equivalent of. The strip puts the answers first and links down to the
 * parts that back them. Nothing here is printed a second time below: the price
 * card that used to hold the price is gone (see ProductGallery).
 *
 * Two cells are honest by omission. The catalog has no free-plan or trial
 * field, so no cell claims one. And when the vendor has no affiliate programme
 * the last cell says so and offers the brand page, rather than a button
 * dressed up as an offer.
 */
export function ProductAtAGlance({ product }: { product: ProductDetail }) {
  const bestFor = product.use_cases[0]
  const evidence = evidenceSummary(product)
  const checked = priceCheckedOn(product)
  const source = priceSource(product)

  return (
    <section
      aria-labelledby="at-a-glance-heading"
      className={cn('mt-8', sectionAnchorClass)}
      id={productSectionIDs.glance}
    >
      <h2 className="sr-only" id="at-a-glance-heading">
        At a glance
      </h2>
      <dl className="grid gap-px border border-ink/15 bg-ink/15 sm:grid-cols-2 lg:grid-cols-4">
        <Cell label="Entry price">
          <PriceDisplay
            amountMinor={product.price.amount_minor}
            currency={product.price.currency}
            size="lg"
          />
          <span className="mt-1 block text-sm text-ink/70">
            {product.key_specification.value || 'Billing basis not recorded'}
          </span>
          <span className="mt-2 block text-xs leading-5 text-ink/65">
            {/* The caveat travelled here with the price. A price is as good
                as its date, and the date is only useful next to the number. */}
            {checked
              ? `Read from the vendor on ${checked} — not a live quote.`
              : 'Not a live quote. The day it was read is not recorded.'}
            {source && (
              <>
                {' '}
                <a
                  className="inline-flex items-center gap-1 text-bronze underline-offset-4 hover:underline"
                  href={source}
                  rel="nofollow noopener noreferrer"
                  target="_blank"
                >
                  Open the page it came from
                  <ExternalLink aria-hidden="true" size={12} />
                </a>
              </>
            )}
          </span>
        </Cell>

        {bestFor && (
          <Cell label="Best for">
            <span className="block font-display text-xl font-medium leading-tight tracking-[-0.03em]">
              {bestFor.label}
            </span>
            <span className="mt-1 block text-xs leading-5 text-ink/65">
              Strongest use case, {bestFor.score}/100
            </span>
          </Cell>
        )}

        {evidence && (
          <Cell label="Evidence">
            <span className="block font-display text-xl font-medium leading-tight tracking-[-0.03em]">
              {evidence.count} {evidence.count === 1 ? 'fact' : 'facts'}
            </span>
            <span className="mt-1 block text-xs leading-5 text-ink/65">
              Median confidence {evidence.medianConfidence}/100.{' '}
              <a
                className="text-bronze underline-offset-4 hover:underline"
                href={`#${productSectionIDs.evidence}`}
              >
                See the record
              </a>
            </span>
          </Cell>
        )}

        <Cell label="Where to get it">
          <VendorControl product={product} />
        </Cell>
      </dl>
    </section>
  )
}

function Cell({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="bg-surface p-4 sm:p-5">
      <dt className="text-[0.625rem] font-bold uppercase tracking-[0.13em] text-ink/65">
        {label}
      </dt>
      <dd className="mt-2">{children}</dd>
    </div>
  )
}

/**
 * The one control on the strip. The button is the one ProductOffers draws
 * lower down — same relationship attributes, same disclosure, and a click is
 * attributed to the product page in the same way.
 */
function VendorControl({ product }: { product: ProductDetail }) {
  const offers = useOffers(product.slug)
  if (offers.isPending) {
    return (
      <span className="block text-sm leading-6 text-ink/65" role="status">
        Checking for a vendor offer…
      </span>
    )
  }
  const offer = offers.isError
    ? undefined
    : offers.data.find((item) => item.purchase_path)
  if (!offer?.purchase_path) {
    return (
      <>
        <span className="block text-sm leading-6 text-ink/70">
          No affiliate offer for this product.
        </span>
        <Link
          className="mt-1 inline-flex min-h-11 items-center text-sm font-semibold text-bronze-dark hover:text-ink"
          to={`/brands/${product.brand.slug}`}
        >
          More from {product.brand.name}
        </Link>
      </>
    )
  }
  return (
    <>
      <a
        className={buttonStyles({ fullWidth: true, size: 'sm' })}
        href={affiliateClickPath(offer.purchase_path, 'product_detail')}
        rel="nofollow noopener sponsored"
        target="_blank"
      >
        View at {offer.merchant.name}{' '}
        <ExternalLink aria-hidden="true" size={15} />
      </a>
      <span className="mt-2 block text-xs leading-5 text-ink/65">
        Affiliate link. Commission never changes recommendation scoring.
      </span>
    </>
  )
}
