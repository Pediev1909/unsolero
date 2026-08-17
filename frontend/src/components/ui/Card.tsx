import type { HTMLAttributes, ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'

type CardVariant = 'plain' | 'outlined' | 'raised' | 'dark'

const variants: Record<CardVariant, string> = {
  plain: 'bg-surface',
  outlined: 'border border-ink/15 bg-surface',
  raised: 'border border-ink/10 bg-surface shadow-raised',
  dark: 'bg-charcoal text-canvas',
}

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  variant?: CardVariant
  padding?: 'none' | 'sm' | 'md' | 'lg'
  children: ReactNode
}

const paddings = {
  none: '',
  sm: 'p-4',
  md: 'p-5 sm:p-6',
  lg: 'p-6 sm:p-8',
}

export function Card({
  variant = 'outlined',
  padding = 'md',
  className,
  children,
  ...props
}: CardProps) {
  return (
    <div
      {...props}
      className={cn(
        'rounded-sm',
        variants[variant],
        paddings[padding],
        className,
      )}
    >
      {children}
    </div>
  )
}
