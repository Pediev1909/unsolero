import { ExternalLink, Truck } from 'lucide-react'
import type { ReactNode } from 'react'

import { Badge } from '../../../components/ui/Badge'
import { Container } from '../../../components/ui/Container'
import { Heading } from '../../../components/ui/Heading'
import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { Skeleton } from '../../../components/ui/Skeleton'
import { buttonStyles } from '../../../components/ui/buttonStyles'
import { formatMinorCurrency } from '../../../lib/money/format'
import { useOffers } from '../queries'
import { affiliateClickPath } from '../../analytics/tracking'

/**
 * Where to get the product, when there is somewhere to send people.
 *
 * This used to render a "Merchant comparison / Available offers" section on
 * every product page, and show an empty state on the 45 of 53 where there is
 * nothing to show. That was the equipment catalog talking: comparing offers
 * across merchants is what you do for a treadmill five shops sell. Software
 * has exactly one seller -- the vendor -- so the section can never fill for
 * a product whose vendor has no programme, and a heading called "Available
 * offers" comparing a single merchant was never going to make sense either.
 *
 * The component owns its whole section now and renders nothing at all when
 * there is nothing to say.
 */
export function ProductOffers({ slug }: { slug: string }) {
  const offers = useOffers(slug)

  if (offers.isPending) {
    return (
      <Section>
        <div
          aria-label="Loading vendor link"
          className="space-y-3"
          role="status"
        >
          <Skeleton className="h-28 w-full" />
        </div>
      </Section>
    )
  }

  // Nothing, on both empty and error. This section is supplementary: the page
  // it sits on has already delivered the price, the scores and the evidence.
  // Putting an error panel below all that, for a section most products do not
  // have, tells the reader something is broken when nothing is.
  if (offers.isError || offers.data.length === 0) return null

  return (
    <Section>
      <div className="space-y-3">
        {offers.data.map((offer) => (
          <article
            className="grid gap-5 border border-ink/15 p-4 sm:grid-cols-[1fr_auto] sm:items-center sm:p-5"
            key={offer.id}
          >
            <div>
              <div className="flex flex-wrap items-center gap-2">
                <h3 className="font-semibold">{offer.merchant.name}</h3>
                {/* "In stock" is a warehouse's answer. A subscription is
                    always available, so the badge only earns its place when it
                    is saying something the reader does not already assume. */}
                {offer.availability !== 'in_stock' && (
                  <Badge variant="warning">
                    {offer.availability.replace('_', ' ')}
                  </Badge>
                )}
                {offer.disclosure_label && (
                  <Badge variant="sponsored">Affiliate link</Badge>
                )}
              </div>
              <div className="mt-3 flex flex-wrap items-baseline gap-x-4 gap-y-1">
                <PriceDisplay
                  amountMinor={offer.price.amount_minor}
                  currency={offer.price.currency}
                  size="md"
                />
                {/* Nothing is shipped to anyone. A lorry icon and the words
                    "Shipping included" under a monthly subscription price is
                    the equipment catalog talking. */}
                {offer.shipping_minor > 0 && (
                  <span className="inline-flex items-center gap-1 text-xs text-ink/70">
                    <Truck aria-hidden="true" size={14} />{' '}
                    {formatMinorCurrency(
                      offer.shipping_minor,
                      offer.price.currency,
                    )}{' '}
                    shipping
                  </span>
                )}
              </div>
              <p className="mt-2 text-xs text-ink/65">
                {/* A landed price that equals the price, and a condition of
                    "new", are both answers to questions nobody asks about
                    software. */}
                {offer.landed_price_minor !== offer.price.amount_minor && (
                  <>
                    Estimated total{' '}
                    {formatMinorCurrency(
                      offer.landed_price_minor,
                      offer.price.currency,
                    )}{' '}
                    ·{' '}
                  </>
                )}
                {offer.condition !== 'new' && <>{offer.condition} · </>}
                {/* The date was always here. What was missing is the word
                    that tells a reader whether it is recent, which they
                    should not have to work out by subtracting dates. */}
                {offer.freshness_status === 'stale'
                  ? 'Price last read'
                  : 'Checked'}{' '}
                {new Intl.DateTimeFormat('en-US', {
                  dateStyle: 'medium',
                }).format(new Date(offer.last_checked_at))}
                {offer.freshness_status === 'stale' && (
                  <> · not re-verified since</>
                )}
                {offer.expires_at
                  ? ` · Valid until ${new Intl.DateTimeFormat('en-US', { dateStyle: 'medium' }).format(new Date(offer.expires_at))}`
                  : ''}
              </p>
            </div>
            {offer.purchase_path && (
              <a
                className={buttonStyles({
                  className: 'w-full sm:w-auto',
                  size: 'sm',
                  variant: 'primary',
                })}
                href={affiliateClickPath(offer.purchase_path, 'product_detail')}
                rel="nofollow noopener sponsored"
                target="_blank"
              >
                View at {offer.merchant.name}{' '}
                <ExternalLink aria-hidden="true" size={15} />
              </a>
            )}
          </article>
        ))}
      </div>
      <p className="mt-4 text-xs leading-5 text-ink/68">
        The price the vendor lists today, with the date it was last checked —
        not a live checkout quote. Outbound links are tracked so we know a click
        came from here. Commission never changes where a product ranks or what
        it scores.
      </p>
    </Section>
  )
}

/**
 * The section chrome, kept in one place so the heading, the copy and the list
 * appear and disappear together. Splitting them across the page and this
 * component is how a heading ended up sitting above an empty state.
 */
function Section({ children }: { children: ReactNode }) {
  return (
    <section className="py-16 sm:py-24">
      <Container>
        <p className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark">
          Where to get it
        </p>
        <Heading className="mt-4" level={2} size="title">
          Straight from the vendor
        </Heading>
        <p className="mt-4 max-w-2xl text-sm leading-6 text-ink/70">
          Software has one seller, so this is not a price comparison. It is the
          vendor&rsquo;s own page, and the date we last read the price on it.
        </p>
        <div className="mt-9">{children}</div>
      </Container>
    </section>
  )
}
