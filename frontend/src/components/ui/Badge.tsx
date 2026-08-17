import type { HTMLAttributes, ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'

type BadgeVariant =
  'neutral' | 'accent' | 'success' | 'warning' | 'error' | 'sponsored'

const variants: Record<BadgeVariant, string> = {
  neutral: 'border-ink/15 bg-paper text-ink',
  accent: 'border-bronze/30 bg-bronze-soft text-bronze-dark',
  success: 'border-moss/30 bg-moss-soft text-moss',
  warning: 'border-amber/30 bg-amber-soft text-amber',
  error: 'border-ember/30 bg-ember-soft text-ember',
  sponsored: 'border-ink/25 bg-transparent text-ink-soft',
}

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant
  children: ReactNode
}

export function Badge({
  variant = 'neutral',
  className,
  children,
  ...props
}: BadgeProps) {
  return (
    <span
      {...props}
      className={cn(
        'inline-flex min-h-6 items-center rounded-full border px-2.5 py-0.5 text-[0.625rem] font-bold uppercase tracking-[0.12em]',
        variants[variant],
        className,
      )}
    >
      {children}
    </span>
  )
}
