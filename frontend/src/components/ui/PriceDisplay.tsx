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

export function PriceDisplay({
  amountMinor,
  currency,
  originalAmountMinor,
  locale,
  size = 'md',
  className,
  ...props
}: PriceDisplayProps) {
  const current = formatMinorCurrency(amountMinor, currency, locale)
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
          'font-display font-medium tabular-nums tracking-[-0.035em]',
          sizes[size],
        )}
      >
        {current}
      </span>
      {original && originalAmountMinor !== amountMinor && (
        <span className="text-xs text-ink/45 line-through">
          <span className="sr-only">Previously </span>
          {original}
        </span>
      )}
    </span>
  )
}
