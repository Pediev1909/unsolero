import { cn } from '../../lib/styles/cn'
import { markColourFor, monogramFor, monogramGround } from './brandMonogram'

// Software has no product photograph, so every card in this catalog reaches the
// no-image branch. The previous placeholder was a generic icon above the vendor
// name, which reads as an absence rather than as a mark. This draws a nameplate
// instead: the vendor's own initials, set in the display face, on one ground.
//
// Nothing here is fetched. Vendor logos are registered trademarks that arrive
// with usage terms, and a catalog of 46 vendors cannot wait on 46 brand kits.
export function BrandMonogram({
  brand,
  category,
  className,
}: {
  brand: string
  category?: string
  className?: string
}) {
  const mark = markColourFor(category ?? brand)

  return (
    <div
      aria-hidden="true"
      className={cn(
        'flex size-full items-center gap-4 px-5 sm:px-6',
        className,
      )}
      style={{ backgroundColor: monogramGround }}
    >
      <span
        className="grid aspect-square h-[58%] min-h-11 shrink-0 place-items-center rounded-[3px] font-display text-[1.35rem] leading-none font-semibold tracking-[-0.03em]"
        style={{ backgroundColor: mark, color: '#ffffff' }}
      >
        {monogramFor(brand)}
      </span>
      <span className="truncate font-display text-[0.95rem] font-medium tracking-[-0.015em] text-ink">
        {brand}
      </span>
    </div>
  )
}
