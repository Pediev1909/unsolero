import type { HTMLAttributes, ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'

type ContainerSize = 'copy' | 'reading' | 'content' | 'wide' | 'full'

const sizes: Record<ContainerSize, string> = {
  copy: 'max-w-copy',
  reading: 'max-w-reading',
  content: 'max-w-content',
  wide: 'max-w-wide',
  full: 'max-w-none',
}

interface ContainerProps extends HTMLAttributes<HTMLDivElement> {
  size?: ContainerSize
  children: ReactNode
}

export function Container({
  size = 'content',
  className,
  children,
  ...props
}: ContainerProps) {
  return (
    <div
      {...props}
      className={cn(
        'mx-auto w-full px-gutter xs:px-5 sm:px-gutter-md xl:px-gutter-lg',
        sizes[size],
        className,
      )}
    >
      {children}
    </div>
  )
}
