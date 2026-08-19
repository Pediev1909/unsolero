import { ExternalLink, Truck } from 'lucide-react'

import { Badge } from '../../../components/ui/Badge'
import { ErrorState } from '../../../components/ui/ErrorState'
import { EmptyState } from '../../../components/ui/EmptyState'
import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { Skeleton } from '../../../components/ui/Skeleton'
import { buttonStyles } from '../../../components/ui/buttonStyles'
import { formatMinorCurrency } from '../../../lib/money/format'
import { useOffers } from '../queries'
import { affiliateClickPath } from '../../analytics/tracking'

export function ProductOffers({
  slug,
  isDemo,
}: {
  slug: string
  isDemo: boolean
}) {
  const offers = useOffers(slug)

  if (offers.isPending) {
    return (
      <div
        aria-label="Loading merchant offers"
        className="space-y-3"
        role="status"
      >
        {[0, 1].map((item) => (
          <Skeleton className="h-28 w-full" key={item} />
        ))}
      </div>
    )
  }
  if (offers.isError) {
    return (
      <ErrorState
        compact
        description="Merchant offers could not be loaded."
        onRetry={() => void offers.refetch()}
        title="Offers unavailable"
      />
    )
  }
  if (offers.data.length === 0) {
    return (
      <EmptyState
        compact
        description={`No verified merchant offers are available for this ${isDemo ? 'demo ' : ''}product.`}
        title="No current offers"
      />
    )
  }

  return (
    <div>
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
                Checked{' '}
                {new Intl.DateTimeFormat('en-US', {
                  dateStyle: 'medium',
                }).format(new Date(offer.last_checked_at))}
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
        Only fresh offer observations are shown; prices and availability are not
        live checkout quotes. Outbound links are tracked for attribution.
        Affiliate commission never changes product ranking or suitability
        scores.
      </p>
    </div>
  )
}
