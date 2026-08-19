import type { ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'

interface StatePanelProps {
  icon?: ReactNode
  eyebrow?: string
  title: string
  description?: string
  action?: ReactNode
  compact?: boolean
  live?: 'polite' | 'assertive' | 'off'
  className?: string
}

export function StatePanel({
  icon,
  eyebrow,
  title,
  description,
  action,
  compact = false,
  live = 'off',
  className,
}: StatePanelProps) {
  return (
    <div
      aria-live={live}
      className={cn(
        'border border-ink/12 bg-surface text-center',
        compact ? 'p-6' : 'px-6 py-12 sm:px-10 sm:py-16',
        className,
      )}
    >
      {icon && (
        <div className="mx-auto mb-5 flex size-10 items-center justify-center border border-ink/15 text-bronze">
          {icon}
        </div>
      )}
      {eyebrow && <p className="eyebrow">{eyebrow}</p>}
      <h2 className="mt-3 font-display text-2xl font-medium tracking-[-0.04em]">
        {title}
      </h2>
      {description && (
        <p className="mx-auto mt-3 max-w-md text-sm leading-6 text-ink/70">
          {description}
        </p>
      )}
      {action && <div className="mt-7">{action}</div>}
    </div>
  )
}
