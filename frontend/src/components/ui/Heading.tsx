import type { HTMLAttributes, ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'

type HeadingLevel = 1 | 2 | 3 | 4
type HeadingSize = 'hero' | 'display' | 'section' | 'title' | 'subtitle'

const sizes: Record<HeadingSize, string> = {
  hero: 'text-[clamp(3.25rem,8vw,7.5rem)] leading-[0.86] tracking-[-0.065em]',
  display: 'text-[clamp(3rem,6vw,6.5rem)] leading-[0.92] tracking-[-0.06em]',
  section: 'text-[clamp(2.5rem,5vw,5.5rem)] leading-[0.98] tracking-[-0.055em]',
  title: 'text-3xl leading-tight tracking-[-0.04em] sm:text-4xl',
  subtitle: 'text-xl leading-tight tracking-[-0.03em] sm:text-2xl',
}

interface HeadingProps extends HTMLAttributes<HTMLHeadingElement> {
  level?: HeadingLevel
  size?: HeadingSize
  balance?: boolean
  children: ReactNode
}

export function Heading({
  level = 2,
  size = 'section',
  balance = true,
  className,
  children,
  ...props
}: HeadingProps) {
  const Element = `h${level}` as const
  return (
    <Element
      {...props}
      className={cn(
        'font-display font-medium',
        sizes[size],
        balance && 'text-balance',
        className,
      )}
    >
      {children}
    </Element>
  )
}
