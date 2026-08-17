import { Dumbbell } from 'lucide-react'
import { Link } from 'react-router-dom'

import { cn } from '../../lib/styles/cn'

interface BrandMarkProps {
  inverse?: boolean
  compact?: boolean
  className?: string
}

export function BrandMark({
  inverse = false,
  compact = false,
  className,
}: BrandMarkProps) {
  return (
    <Link
      aria-label="UNSOLERO home"
      className={cn('inline-flex items-center gap-2.5', className)}
      to="/"
    >
      <span
        className={cn(
          'grid size-8 place-items-center',
          inverse ? 'bg-canvas text-ink' : 'bg-ink text-canvas',
        )}
      >
        <Dumbbell aria-hidden="true" size={17} strokeWidth={1.75} />
      </span>
      {!compact && (
        <span className="text-lg font-semibold tracking-[-0.03em]">
          UNSOLERO
        </span>
      )}
    </Link>
  )
}
