import type { HTMLAttributes } from 'react'

import { cn } from '../../lib/styles/cn'

interface SkeletonProps extends HTMLAttributes<HTMLDivElement> {
  shape?: 'line' | 'block' | 'circle'
}

export function Skeleton({
  shape = 'block',
  className,
  ...props
}: SkeletonProps) {
  return (
    <div
      {...props}
      aria-hidden="true"
      className={cn(
        'skeleton-shimmer bg-paper',
        shape === 'line' && 'h-4 rounded-xs',
        shape === 'block' && 'min-h-24 rounded-sm',
        shape === 'circle' && 'aspect-square rounded-full',
        className,
      )}
    />
  )
}
