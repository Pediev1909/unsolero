import type { ContentSummary } from '../schemas'

// Thirteen comparisons shared one illustration and six buying guides shared
// another, so a grid of them read as one template repeated rather than a
// library of distinct pieces. The products each piece covers are already
// different on every card, so naming the card after them makes it distinct for
// free — and says more in the same space than any drawing could.
//
// Falls back to the illustration when a piece covers nothing, which is how
// pieces without a product set keep a picture.

function priceRange(covered: ContentSummary['covered']) {
  const amounts = covered.map((product) => product.price_minor)
  const currency = covered[0]?.currency ?? 'USD'
  const format = (amountMinor: number) =>
    amountMinor === 0
      ? 'Free'
      : new Intl.NumberFormat('en-US', {
          style: 'currency',
          currency,
          minimumFractionDigits: amountMinor % 100 === 0 ? 0 : 2,
          maximumFractionDigits: 2,
        }).format(amountMinor / 100)

  const lowest = Math.min(...amounts)
  const highest = Math.max(...amounts)
  return lowest === highest
    ? format(lowest)
    : `${format(lowest)}–${format(highest)}`
}

export function CardNameplate({ entry }: { entry: ContentSummary }) {
  const covered = entry.covered

  if (covered.length === 0) {
    return (
      <div className="aspect-[16/10] overflow-hidden bg-paper">
        <img
          alt={entry.hero_image.alt_text}
          className="size-full object-cover transition-transform duration-280 group-hover:scale-[1.015] motion-reduce:transform-none"
          loading="lazy"
          src={entry.hero_image.url}
        />
      </div>
    )
  }

  // Four names is where a card stops reading as a title and starts reading as a
  // list. The rest are counted rather than named.
  const named = covered.slice(0, 4)
  const remaining = covered.length - named.length

  return (
    <div className="flex min-h-[9.5rem] flex-col justify-center gap-2.5 bg-canvas px-5 py-6 sm:px-6">
      <p className="font-display text-xl leading-[1.16] tracking-[-0.025em] sm:text-[1.4rem]">
        {named.map((product, index) => (
          <span key={product.name}>
            {index > 0 && (
              <span className="px-1.5 text-[0.7rem] tracking-[0.08em] text-ink/35">
                VS
              </span>
            )}
            {product.name}
          </span>
        ))}
        {remaining > 0 && (
          <span className="text-ink/45"> and {remaining} more</span>
        )}
      </p>

      <p className="flex flex-wrap items-baseline gap-x-4 gap-y-1 border-t border-ink/12 pt-2.5 text-xs text-ink/60">
        <span>
          <span className="font-semibold text-ink">{covered.length}</span> tools
        </span>
        <span>
          <span className="text-sm font-bold text-bronze-dark tabular-nums">
            {priceRange(covered)}
          </span>{' '}
          a month
        </span>
      </p>
    </div>
  )
}
