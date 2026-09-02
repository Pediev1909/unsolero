import { ExternalLink } from 'lucide-react'
import { Link } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { ButtonLink } from '../components/ui/ButtonLink'
import { Container } from '../components/ui/Container'
import { EmptyState } from '../components/ui/EmptyState'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { LoadingState } from '../components/ui/LoadingState'
import { PriceDisplay } from '../components/ui/PriceDisplay'
import { affiliateClickPath } from '../features/analytics/tracking'
import { BrandMark } from '../features/catalog/components/BrandMark'
import { useLiveOffers } from '../features/catalog/queries'
import type { LiveOffer } from '../features/catalog/schemas'
import { formatEditorialDate } from '../features/content/model'
import { usePageMetadata } from '../lib/seo/usePageMetadata'
import { useStructuredData } from '../lib/seo/useStructuredData'

const description =
  'Every live vendor offer in the UNSOLERO catalog, with the price and the date it was last read. Affiliate links, and the ranking does not know they exist.'

interface OfferGroup {
  name: string
  slug: string
  items: LiveOffer[]
}

/**
 * One group per category, in the order the API sent them (sorted by category,
 * catalog order within it). Fifteen rows under one heading is a list; the same
 * fifteen under "Email marketing", "CRM" and so on is something a reader can
 * skip through to the job they are hiring software for.
 */
function groupByCategory(items: LiveOffer[]): OfferGroup[] {
  const groups = new Map<string, OfferGroup>()
  for (const item of items) {
    const { name, slug } = item.product.category
    const group = groups.get(slug)
    if (group) group.items.push(item)
    else groups.set(slug, { name, slug, items: [item] })
  }
  return [...groups.values()]
}

export function OffersPage() {
  const offers = useLiveOffers()
  usePageMetadata({
    title: 'Live vendor offers and trials | UNSOLERO',
    description,
  })
  useStructuredData('live-offers', {
    '@context': 'https://schema.org',
    '@type': 'CollectionPage',
    name: 'Live vendor offers',
    description,
    url: window.location.href,
  })

  const groups = groupByCategory(offers.data?.items ?? [])
  const total = offers.data?.items.length ?? 0

  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <section className="border-b border-ink/15 py-16 sm:py-24 lg:py-28">
          <Container>
            <p className="eyebrow">Where to get it</p>
            <Heading className="mt-5 max-w-4xl" level={1} size="display">
              Live vendor offers
            </Heading>
            <p className="mt-7 max-w-2xl text-lg leading-8 text-ink/65">
              Every product in the catalog with a working link to its vendor
              today. The price beside each is what the vendor&rsquo;s own page
              said on the date shown, not a live checkout quote. These are
              affiliate links: we may earn a commission if you subscribe, and
              that commission has no say in{' '}
              <Link
                className="underline underline-offset-4 hover:text-bronze-dark"
                to="/articles/how-unsolero-ranks-software"
              >
                how we rank software
              </Link>
              . Read the{' '}
              <Link
                className="underline underline-offset-4 hover:text-bronze-dark"
                to="/affiliate-disclosure"
              >
                affiliate disclosure
              </Link>
              .
            </p>
          </Container>
        </section>

        <Container className="py-12 sm:py-18 lg:py-24">
          {offers.isPending && (
            <LoadingState
              description="Checking which vendor links are live."
              title="Loading offers"
            />
          )}
          {offers.isError && (
            <ErrorState
              description="The list of live offers could not be loaded. Every product page still shows its own price and vendor link."
              onRetry={() => void offers.refetch()}
              title="Offers unavailable"
            />
          )}
          {offers.isSuccess && total === 0 && (
            <EmptyState
              action={
                <ButtonLink to="/products" variant="secondary">
                  Browse all software
                </ButtonLink>
              }
              description="Every vendor link we show is re-checked against the vendor's own page before it is drawn, and none passes right now. The catalog, the prices and the rankings are unaffected."
              title="No live offers right now"
            />
          )}

          {offers.isSuccess && total > 0 && (
            <div className="flex flex-col gap-14">
              {groups.map((group) => (
                <section
                  aria-labelledby={`offers-${group.slug}`}
                  key={group.slug}
                >
                  <div className="flex items-baseline justify-between gap-4 border-b border-ink/15 pb-3">
                    <h2
                      className="scroll-mt-24 font-display text-2xl font-medium tracking-[-0.03em]"
                      id={`offers-${group.slug}`}
                    >
                      {group.name}
                    </h2>
                    <Link
                      className="shrink-0 text-sm font-semibold text-ink/70 hover:text-bronze-dark"
                      to={`/categories/${group.slug}`}
                    >
                      All {group.name.toLowerCase()}
                    </Link>
                  </div>
                  <ul>
                    {group.items.map((item) => (
                      <OfferRow item={item} key={item.product.id} />
                    ))}
                  </ul>
                </section>
              ))}

              <p className="border-t border-ink/15 pt-8 text-sm leading-6 text-ink/60">
                {total} {total === 1 ? 'product' : 'products'} with a working
                vendor link as of{' '}
                {formatEditorialDate(offers.data.generated_at)}. Outbound links
                are tracked so we know a click came from here. Commission never
                changes where a product ranks or what it scores.
              </p>
            </div>
          )}
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}

/**
 * One product, one vendor, one dated price.
 *
 * The button is the same tracked, `sponsored` exit the catalog card draws,
 * not MerchantAction: that component fetches the product's offers itself, and
 * fifteen rows would mean fifteen requests for facts this page already holds.
 */
function OfferRow({ item }: { item: LiveOffer }) {
  const { product, offer } = item
  const stale = offer.freshness_status === 'stale'
  return (
    <li className="border-b border-ink/15 py-5 sm:grid sm:grid-cols-[1fr_auto] sm:items-center sm:gap-6">
      <div className="flex items-start gap-3">
        <BrandMark
          brandName={product.brand.name}
          brandSlug={product.brand.slug}
          size="md"
        />
        <div className="min-w-0 flex-1">
          <h3 className="font-display text-lg leading-tight font-medium tracking-[-0.03em]">
            <Link
              className="hover:text-bronze-dark"
              to={`/products/${product.slug}`}
            >
              {product.name}
            </Link>
          </h3>
          <p className="mt-2 flex flex-wrap items-baseline gap-x-2 gap-y-1">
            <PriceDisplay
              amountMinor={offer.price.amount_minor}
              currency={offer.price.currency}
              size="sm"
            />
            <span className="text-sm text-ink/65">
              {product.key_specification.value}
            </span>
          </p>
          {/* The date was always the point. What a reader should not have to
              work out by subtracting dates is whether it is recent, so a
              price older than the editorial window says so in words. */}
          <p className="mt-1 text-xs leading-5 text-ink/65">
            {stale ? 'Price last read' : 'Checked'}{' '}
            {formatEditorialDate(offer.last_checked_at)}
            {stale && ' · not re-verified since'}
            {' · '}
            {product.disclosure_label ?? 'Affiliate link'}
          </p>
        </div>
      </div>
      {product.purchase_path && (
        <a
          className="mt-4 inline-flex min-h-11 w-full items-center justify-center gap-2 bg-charcoal px-4 text-sm font-semibold text-canvas hover:bg-ink sm:mt-0 sm:w-auto"
          href={affiliateClickPath(product.purchase_path, 'product_detail')}
          rel="nofollow noopener sponsored"
          target="_blank"
        >
          View at {offer.merchant_name}
          <ExternalLink aria-hidden="true" size={14} />
        </a>
      )}
    </li>
  )
}
