import type { ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'

interface ProductGridProps {
  children: ReactNode
  columns?: 2 | 3 | 4
  className?: string
}

const columns = {
  2: 'md:grid-cols-2',
  3: 'md:grid-cols-2 lg:grid-cols-3',
  4: 'md:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4',
}

export function ProductGrid({
  children,
  columns: columnCount = 3,
  className,
}: ProductGridProps) {
  return (
    <div
      className={cn(
        'grid grid-cols-1 overflow-hidden border-l border-t border-ink/15',
        columns[columnCount],
        className,
      )}
    >
      {children}
    </div>
  )
}
