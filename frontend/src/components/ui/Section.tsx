import type { HTMLAttributes, ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'

type SectionSpace = 'sm' | 'md' | 'lg'
type SectionSurface = 'canvas' | 'surface' | 'paper' | 'ink'

const spaces: Record<SectionSpace, string> = {
  sm: 'py-section-sm',
  md: 'py-section-sm md:py-section',
  lg: 'py-section md:py-section-lg',
}

const surfaces: Record<SectionSurface, string> = {
  canvas: 'bg-canvas text-ink',
  surface: 'bg-surface text-ink',
  paper: 'bg-paper text-ink',
  ink: 'bg-ink text-canvas',
}

interface SectionProps extends HTMLAttributes<HTMLElement> {
  space?: SectionSpace
  surface?: SectionSurface
  children: ReactNode
}

export function Section({
  space = 'md',
  surface = 'canvas',
  className,
  children,
  ...props
}: SectionProps) {
  return (
    <section
      {...props}
      className={cn(spaces[space], surfaces[surface], className)}
    >
      {children}
    </section>
  )
}
