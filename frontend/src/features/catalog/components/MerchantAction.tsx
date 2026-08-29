import { ExternalLink } from 'lucide-react'

import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { Skeleton } from '../../../components/ui/Skeleton'
import { useOffers } from '../queries'
import {
  affiliateClickPath,
  type AffiliateSource,
} from '../../analytics/tracking'

/**
 * The merchant call to action.
 *
 * `compact` exists for surfaces that already carry their own affiliate
 * disclosure and their own separators — the comparison table's footer row is
 * the first — where repeating the rule under every column would say the same
 * sentence four times in one screen. It drops the frame and the sentence, not
 * the `sponsored` relationship or the tracked path: those are what make the
 * link honest and what make it pay, and neither is a styling concern.
 *
 * A surface with no disclosure of its own must use the default variant.
 */
export function MerchantAction({
  slug,
  className = 'mt-5',
  source = 'product_detail',
  recommendationID,
  compact = false,
}: {
  slug: string
  className?: string
  source?: AffiliateSource
  recommendationID?: string | null
  compact?: boolean
}) {
  const offers = useOffers(slug)
  if (offers.isPending)
    return <Skeleton className={`${className} h-11 w-full`} />
  if (offers.isError || offers.data.length === 0)
    return compact ? null : (
      <p className={`${className} text-xs text-ink/68`}>
        No verified merchant offer is currently available.
      </p>
    )
  const offer = offers.data[0]
  if (!offer?.purchase_path) return null
  const action = (
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
  )
  if (compact) return <div className={className}>{action}</div>
  return (
    <div className={`${className} border-t border-ink/10 pt-5`}>
      {action}
      <p className="mt-2 text-xs text-ink/65">
        Affiliate link. Commission never changes recommendation scoring.
      </p>
    </div>
  )
}
