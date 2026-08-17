import type { PropsWithChildren } from 'react'

import { Heading } from '../ui/Heading'
import { Container } from '../ui/Container'
import { SkipLink } from '../ui/SkipLink'
import { BrandMark } from './BrandMark'
interface AuthLayoutProps extends PropsWithChildren {
  eyebrow: string
  title: string
  description: string
}

export function AuthLayout({
  children,
  eyebrow,
  title,
  description,
}: AuthLayoutProps) {
  return (
    <>
      <SkipLink />
      <main className="min-h-screen" id="main-content">
        <Container className="flex min-h-screen flex-col">
          <header className="flex h-20 items-center">
            <BrandMark />
          </header>

          <div className="grid flex-1 items-center py-12 lg:grid-cols-[1fr_0.8fr] lg:gap-24">
            <section className="max-w-xl">
              <p className="eyebrow">{eyebrow}</p>
              <Heading className="mt-5" level={1} size="display">
                {title}
              </Heading>
              <p className="mt-7 max-w-lg text-base leading-7 text-ink/60">
                {description}
              </p>
            </section>

            <section className="mt-12 border-t border-ink/15 pt-8 lg:mt-0 lg:border-l lg:border-t-0 lg:pl-16 lg:pt-0">
              {children}
            </section>
          </div>
        </Container>
      </main>
    </>
  )
}
