import { Star } from 'lucide-react'
import type { HTMLAttributes } from 'react'

import { cn } from '../../lib/styles/cn'

interface RatingProps extends HTMLAttributes<HTMLDivElement> {
  value: number
  max?: number
  label?: string
  showValue?: boolean
  size?: 'sm' | 'md'
}

export function Rating({
  value,
  max = 5,
  label = 'Rating',
  showValue = true,
  size = 'sm',
  className,
  ...props
}: RatingProps) {
  const normalizedMax = Math.max(1, Math.floor(max))
  const boundedValue = Math.min(Math.max(value, 0), normalizedMax)
  const iconSize = size === 'sm' ? 14 : 18

  return (
    <div
      {...props}
      aria-label={`${label}: ${boundedValue.toFixed(1)} out of ${normalizedMax}`}
      className={cn('inline-flex items-center gap-2', className)}
      role="img"
    >
      <span aria-hidden="true" className="flex gap-0.5 text-bronze">
        {Array.from({ length: normalizedMax }, (_, index) => {
          const fill = Math.min(Math.max(boundedValue - index, 0), 1)
          return (
            <span className="relative" key={index}>
              <Star className="text-line" size={iconSize} strokeWidth={1.5} />
              {fill > 0 && (
                <span
                  className="absolute inset-0 overflow-hidden"
                  style={{ width: `${fill * 100}%` }}
                >
                  <Star fill="currentColor" size={iconSize} strokeWidth={1.5} />
                </span>
              )}
            </span>
          )
        })}
      </span>
      {showValue && (
        <span aria-hidden="true" className="text-xs font-semibold text-ink/60">
          {boundedValue.toFixed(1)}
        </span>
      )}
    </div>
  )
}
