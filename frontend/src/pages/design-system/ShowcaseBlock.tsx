import type { ReactNode } from 'react'

import { Heading } from '../../components/ui/Heading'

interface ShowcaseBlockProps {
  eyebrow: string
  title: string
  description: string
  children: ReactNode
}

export function ShowcaseBlock({
  eyebrow,
  title,
  description,
  children,
}: ShowcaseBlockProps) {
  return (
    <section className="border-t border-ink/15 py-14 sm:py-18 lg:py-24">
      <div className="grid gap-10 lg:grid-cols-[0.35fr_1fr] lg:gap-16">
        <div>
          <p className="eyebrow">{eyebrow}</p>
          <Heading className="mt-4" level={2} size="title">
            {title}
          </Heading>
          <p className="mt-4 max-w-sm text-sm leading-6 text-ink/70">
            {description}
          </p>
        </div>
        <div className="min-w-0">{children}</div>
      </div>
    </section>
  )
}
