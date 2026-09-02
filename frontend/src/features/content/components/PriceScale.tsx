import type { ProductSummary } from '../../catalog/schemas'
import { usePriceRows } from './priceRows'

function formatPrice(amountMinor: number, currency: string) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: amountMinor % 100 === 0 ? 0 : 2,
    maximumFractionDigits: 2,
  }).format(amountMinor / 100)
}

export function PriceScale({ products }: { products: ProductSummary[] }) {
  const scale = usePriceRows(products)
  if (!scale) return null

  return (
    <figure className="border-y border-ink/12 bg-canvas px-5 py-8 sm:px-9 sm:py-10">
      <figcaption className="text-[0.625rem] font-bold uppercase tracking-[0.14em] text-ink/65">
        What each one costs, entry paid tier
      </figcaption>

      <ul className="mt-6 space-y-px">
        {scale.rows.map((row, index) => (
          <li
            className="grid grid-cols-[minmax(6.5rem,11rem)_minmax(0,1fr)_auto] items-center gap-4 py-2.5 sm:gap-6"
            key={row.id}
          >
            <span className="truncate text-[0.8125rem] font-semibold leading-tight sm:text-sm">
              {row.name}
            </span>

            <span className="relative block h-2.5 bg-ink/8">
              {/* One authored moment: the bars grow to their real length, in
                  price order, once. Their resting state is the finished
                  length, so reduced motion shows the comparison outright. */}
              <span
                className={`absolute inset-y-0 left-0 origin-left animate-[price-bar_760ms_cubic-bezier(.16,1,.3,1)_both] motion-reduce:animate-none ${
                  row.cheapest ? 'bg-bronze' : 'bg-ink/45'
                }`}
                style={{
                  width: `${row.width}%`,
                  animationDelay: `${120 + index * 110}ms`,
                }}
              />
            </span>

            <span
              className={`text-right text-base leading-tight tracking-[-0.02em] tabular-nums sm:text-lg ${
                row.cheapest ? 'text-bronze' : 'text-ink/70'
              }`}
            >
              {row.free ? 'Free' : formatPrice(row.amountMinor, row.currency)}
              {row.annualOnly && (
                <span className="block text-[0.625rem] leading-4 tracking-normal text-ink/60">
                  billed yearly
                </span>
              )}
            </span>
          </li>
        ))}
      </ul>

      <p className="mt-5 border-t border-ink/12 pt-4 text-body-sm leading-6 text-ink/68">
        Prices are the entry paid tier recorded in the catalog, per month. Where
        a vendor sells only yearly contracts the row says so. What the cheapest
        option costs you in setup, migration or a missing feature is what the
        rest of this piece is about.
      </p>
    </figure>
  )
}

// The page cannot ask "did the scale render?" without calling a hook
// conditionally, so the decision lives here: a piece with a real spread of
// prices gets the comparison, and everything else keeps its illustration.
export function EditorialHero({
  products,
  image,
}: {
  products: ProductSummary[]
  image: { url: string; alt_text: string }
}) {
  const scale = usePriceRows(products)

  if (scale) return <PriceScale products={products} />

  return (
    <div className="aspect-[16/9] max-h-[48rem] overflow-hidden bg-paper">
      <img
        alt={image.alt_text}
        className="size-full object-cover"
        fetchPriority="high"
        src={image.url}
      />
    </div>
  )
}
