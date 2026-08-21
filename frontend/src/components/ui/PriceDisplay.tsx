import type { HTMLAttributes } from 'react'

import { formatMinorCurrency } from '../../lib/money/format'
import { cn } from '../../lib/styles/cn'

interface PriceDisplayProps extends HTMLAttributes<HTMLSpanElement> {
  amountMinor: number
  currency: string
  originalAmountMinor?: number
  locale?: string
  size?: 'sm' | 'md' | 'lg'
}

const sizes = {
  sm: 'text-sm',
  md: 'text-xl',
  lg: 'text-3xl',
}

// A phrase needs less room than a figure at the same visual weight, so the
// free label steps down one size and keeps a comparison row from reflowing.
const freeSizes = {
  sm: 'text-sm',
  md: 'text-base',
  lg: 'text-xl',
}

export function PriceDisplay({
  amountMinor,
  currency,
  originalAmountMinor,
  locale,
  size = 'md',
  className,
  ...props
}: PriceDisplayProps) {
  // Zero is a real price here, not missing data. Six products in the catalog
  // charge nothing per month and take a percentage of each sale instead, and
  // "$0.00" in a comparison column reads as a bug rather than as a fact. The
  // wording stays deliberately narrow: "No monthly fee" is true of a payment
  // processor that still takes 2.9%, where "Free" would not be.
  const isFree = amountMinor === 0
  const current = isFree
    ? 'No monthly fee'
    : formatMinorCurrency(amountMinor, currency, locale)
  const original =
    originalAmountMinor === undefined
      ? null
      : formatMinorCurrency(originalAmountMinor, currency, locale)

  return (
    <span
      {...props}
      className={cn('inline-flex items-baseline gap-2', className)}
    >
      {/* Tabular figures keep prices aligned when several sit in a column,
          which is most of this site. Proportional digits make a comparison
          table look like it was typeset by accident. */}
      <span
        className={cn(
          'font-display font-medium tracking-[-0.035em]',
          // Tabular figures only help when the content is figures. Applying
          // them to a phrase just spaces the letters oddly.
          isFree ? 'text-ink/80' : 'tabular-nums',
          isFree ? freeSizes[size] : sizes[size],
        )}
      >
        {current}
      </span>
      {original && originalAmountMinor !== amountMinor && (
        <span className="text-xs text-ink/65 line-through">
          <span className="sr-only">Previously </span>
          {original}
        </span>
      )}
    </span>
  )
}
