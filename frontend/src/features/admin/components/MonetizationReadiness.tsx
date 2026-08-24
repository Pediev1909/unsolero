import { Link } from 'react-router-dom'

import { Badge } from '../../../components/ui/Badge'
import type { DashboardData } from '../schemas'

// Inventory totals say how much exists. They do not say how much of it can take
// money, which is the only number that decides whether traffic is worth
// sending. A published product with no active offer, or an offer with no active
// affiliate link, is a page that spends attention and returns nothing.

type Readiness = DashboardData['readiness']

const reasonLabels: Record<Readiness['blocked'][number]['reason'], string> = {
  no_active_offer: 'No active offer',
  no_affiliate_link: 'No affiliate link',
}

export function MonetizationReadiness({ readiness }: { readiness: Readiness }) {
  const { published_products: published, earning_ready: ready } = readiness
  const share = published === 0 ? 0 : Math.round((ready / published) * 100)
  const blocked =
    readiness.without_active_offer + readiness.without_affiliate_link

  return (
    <section>
      <h2 className="font-editorial text-xl">Can the catalog earn?</h2>
      <p className="mt-2 max-w-2xl text-body-sm text-ink/70">
        A published product earns only through an active offer carrying an
        active affiliate link. Everything else is a page that costs traffic and
        returns nothing.
      </p>

      <div className="mt-5 grid gap-5 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)]">
        <div className="border border-ink/12 bg-surface p-6">
          <p className="text-[0.625rem] font-bold uppercase tracking-[0.14em] text-ink/65">
            Earning ready
          </p>
          <p className="mt-3 font-editorial text-4xl tabular-nums">
            {ready.toLocaleString()}
            <span className="text-ink/45"> / {published.toLocaleString()}</span>
          </p>
          <div
            aria-hidden="true"
            className="mt-4 h-2 w-full overflow-hidden bg-paper"
          >
            <div
              className="h-full bg-moss"
              style={{ width: `${Math.max(share, published === 0 ? 0 : 1)}%` }}
            />
          </div>
          <p className="mt-3 text-sm text-ink/70">
            {published === 0
              ? 'Nothing is published yet.'
              : `${share}% of the published catalog can take money.`}
          </p>
        </div>

        <div className="grid gap-px bg-ink/12 sm:grid-cols-2">
          <GapTile
            action="Add an offer"
            description="Published, but no active merchant offer exists."
            label="Without an offer"
            to="/admin/offers"
            value={readiness.without_active_offer}
          />
          <GapTile
            action="Add a link"
            description="An offer exists, but no active affiliate link is attached."
            label="Without a link"
            to="/admin/affiliate-links"
            value={readiness.without_affiliate_link}
          />
        </div>
      </div>

      {readiness.blocked.length > 0 && (
        <div className="mt-6 border border-ink/12 bg-surface">
          <div className="flex flex-wrap items-baseline justify-between gap-3 border-b border-ink/12 px-5 py-4">
            <h3 className="font-semibold">Next products to fix</h3>
            <p className="text-sm text-ink/65">
              {readiness.blocked.length === blocked
                ? `${blocked} product${blocked === 1 ? '' : 's'}`
                : `First ${readiness.blocked.length} of ${blocked}`}
            </p>
          </div>
          <ul className="divide-y divide-ink/8">
            {readiness.blocked.map((item) => (
              <li
                className="flex flex-wrap items-center justify-between gap-3 px-5 py-3"
                key={item.id}
              >
                <Link
                  className="font-semibold text-bronze-dark hover:text-ink"
                  to={`/admin/products/${item.id}`}
                >
                  {item.name}
                </Link>
                <Badge
                  variant={
                    item.reason === 'no_active_offer' ? 'neutral' : 'warning'
                  }
                >
                  {reasonLabels[item.reason]}
                </Badge>
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
  )
}

function GapTile({
  label,
  value,
  description,
  action,
  to,
}: {
  label: string
  value: number
  description: string
  action: string
  to: string
}) {
  return (
    <div className="bg-surface p-6">
      <p className="text-[0.625rem] font-bold uppercase tracking-[0.14em] text-ink/65">
        {label}
      </p>
      <p className="mt-3 font-editorial text-3xl tabular-nums">
        {value.toLocaleString()}
      </p>
      <p className="mt-2 text-sm leading-6 text-ink/70">{description}</p>
      {value > 0 && (
        <Link
          className="mt-3 inline-block text-sm font-semibold text-bronze-dark hover:text-ink"
          to={to}
        >
          {action}
        </Link>
      )}
    </div>
  )
}
