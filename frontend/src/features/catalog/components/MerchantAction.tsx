import { ExternalLink } from 'lucide-react'

import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { Skeleton } from '../../../components/ui/Skeleton'
import { useOffers } from '../queries'
import {
  affiliateClickPath,
  type AffiliateSource,
} from '../../analytics/tracking'

export function MerchantAction({
  slug,
  className = 'mt-5',
  source = 'product_detail',
  recommendationID,
}: {
  slug: string
  className?: string
  source?: AffiliateSource
  recommendationID?: string | null
}) {
  const offers = useOffers(slug)
  if (offers.isPending)
    return <Skeleton className={`${className} h-11 w-full`} />
  if (offers.isError || offers.data.length === 0)
    return (
      <p className={`${className} text-xs text-ink/50`}>
        No verified merchant offer is currently available.
      </p>
    )
  const offer = offers.data[0]
  if (!offer?.purchase_path) return null
  return (
    <div className={`${className} border-t border-ink/10 pt-5`}>
      <a
        className="inline-flex min-h-11 w-full items-center justify-center gap-2 bg-charcoal px-4 text-sm font-semibold text-canvas hover:bg-ink"
        href={affiliateClickPath(offer.purchase_path, source, recommendationID)}
        rel="nofollow noopener sponsored"
        target="_blank"
      >
        View at {offer.merchant.name} ·{' '}
        <PriceDisplay
          amountMinor={offer.landed_price_minor}
          currency={offer.price.currency}
          size="sm"
        />
        <ExternalLink aria-hidden="true" size={14} />
      </a>
      <p className="mt-2 text-xs text-ink/45">
        Affiliate link. Commission never changes recommendation scoring.
      </p>
    </div>
  )
}
