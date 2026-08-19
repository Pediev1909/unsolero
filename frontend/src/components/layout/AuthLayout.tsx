import type { PropsWithChildren } from 'react'
import { Link } from 'react-router-dom'

import { Heading } from '../ui/Heading'
import { Container } from '../ui/Container'
import { SkipLink } from '../ui/SkipLink'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'
import { BrandMark } from './BrandMark'

interface AuthLayoutProps extends PropsWithChildren {
  eyebrow: string
  title: string
  description: string
  // The heading speaks to the reader ("Welcome back."); the document title has
  // to say which site and which page, because it is read out of context in a
  // browser tab, a bookmark and a shared link. Set here rather than on each
  // page so a new account screen cannot ship without one.
  documentTitle: string
  documentDescription?: string
}

export function AuthLayout({
  children,
  eyebrow,
  title,
  description,
  documentTitle,
  documentDescription,
}: AuthLayoutProps) {
  usePageMetadata({
    title: documentTitle,
    description: documentDescription ?? description,
    // Account screens carry no content worth ranking, and an indexed sign-in
    // page competes with the pages that do.
    robots: 'noindex, follow',
  })

  return (
    // The masthead and the footer sit outside <main> so they are announced as
    // the banner and contentinfo landmarks. Nested inside it, as they were,
    // they are ordinary sections and an account screen offers a screen reader
    // no landmarks at all to jump between.
    <div className="flex min-h-screen flex-col">
      <SkipLink />
      <Container>
        <header className="flex h-20 items-center">
          <BrandMark />
        </header>
      </Container>

      <main className="flex-1" id="main-content">
        <Container className="grid h-full items-center py-12 lg:grid-cols-[1fr_0.8fr] lg:gap-24">
          <section className="max-w-xl">
            <p className="eyebrow">{eyebrow}</p>
            <Heading className="mt-5" level={1} size="display">
              {title}
            </Heading>
            <p className="mt-7 max-w-lg text-base leading-7 text-ink/70">
              {description}
            </p>
          </section>

          <section className="mt-12 border-t border-ink/15 pt-8 lg:mt-0 lg:border-l lg:border-t-0 lg:pl-16 lg:pt-0">
            {children}
          </section>
        </Container>
      </main>

      {/* Deliberately not the site footer: an account flow should not be full
          of exits. But someone being asked to create an account is entitled to
          read what happens to their data and how the site earns before they
          agree to either, so those two links stay. */}
      <Container>
        <footer className="flex flex-wrap items-center gap-x-6 gap-y-1 border-t border-ink/10 py-6 text-caption text-ink/65">
          <p>© {new Date().getFullYear()} UNSOLERO</p>
          <Link
            className="flex min-h-6 items-center underline underline-offset-4 hover:text-ink"
            to="/privacy"
          >
            Privacy
          </Link>
          <Link
            className="flex min-h-6 items-center underline underline-offset-4 hover:text-ink"
            to="/affiliate-disclosure"
          >
            Affiliate disclosure
          </Link>
          <Link
            className="flex min-h-6 items-center underline underline-offset-4 hover:text-ink"
            to="/"
          >
            Back to UNSOLERO
          </Link>
        </footer>
      </Container>
    </div>
  )
}
