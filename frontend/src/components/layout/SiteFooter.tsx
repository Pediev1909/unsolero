import { Link } from 'react-router-dom'

import { Container } from '../ui/Container'
import { BrandMark } from './BrandMark'
import { openAnalyticsConsentPreferences } from '../../features/analytics/consent'
import { NewsletterForm } from '../../features/newsletter/components/NewsletterForm'

// Grouped rather than listed. A single alphabet-soup column of ten links makes
// the reader scan all of them to find one; under a heading they can skip to the
// group they want. The grouping is editorial and deliberately differs from the
// header's navigation, which carries only the primary journeys.
const footerGroups: {
  heading: string
  links: { label: string; to: string }[]
}[] = [
  {
    heading: 'Decide',
    links: [
      { label: 'Build my setup', to: '/build' },
      { label: 'Software categories', to: '/categories' },
      { label: 'All software', to: '/products' },
      { label: 'All vendors', to: '/brands' },
      { label: 'Live offers', to: '/offers' },
      { label: 'Compare', to: '/compare' },
      { label: 'Wishlist', to: '/wishlist' },
    ],
  },
  {
    heading: 'Learn',
    links: [
      { label: 'How it works', to: '/how-it-works' },
      { label: 'Comparisons', to: '/comparisons' },
      { label: 'Stacks', to: '/stacks' },
      { label: 'Guides', to: '/guides' },
      { label: 'Articles', to: '/articles' },
      {
        label: 'Free funnel training',
        to: '/offers/funnel-hacking-secrets',
      },
      { label: 'About', to: '/about' },
      // Partner programmes and press open the site looking for a human to
      // write to, and the address existed only on the About page.
      { label: 'Contact', to: 'mailto:hello@unsolero.com' },
    ],
  },
  {
    heading: 'Account',
    links: [
      { label: 'Sign in', to: '/login' },
      { label: 'Create account', to: '/register' },
      { label: 'Privacy', to: '/privacy' },
      { label: 'Terms', to: '/terms' },
      { label: 'Affiliate disclosure', to: '/affiliate-disclosure' },
    ],
  },
]

const footerLinkStyle =
  'flex min-h-6 items-center text-sm text-canvas/75 transition-colors duration-150 hover:text-canvas'

/**
 * `newsletter` is false on the pages that are the end of a newsletter
 * decision. Offering "one email when a price changes" directly beneath "You
 * are unsubscribed." reads as not having listened, and a sign-up form on a
 * confirmation page competes with the thing the reader just did.
 */
export function SiteFooter({ newsletter = true }: { newsletter?: boolean }) {
  return (
    <footer className="border-t border-canvas/10 bg-ink py-10 text-canvas">
      <Container>
        {/* The newsletter panel is light on the dark footer so the shared Input
            and Button keep the contrast they were designed for: the footer's
            inverse palette has no error colour that reads on ink. */}
        {newsletter && (
          <div className="mb-8 border-b border-canvas/15 pb-8">
            <div className="bg-canvas px-5 py-6 text-ink sm:px-7 sm:py-7">
              <NewsletterForm compact source="footer" />
            </div>
          </div>
        )}
        <div className="grid gap-8 border-b border-canvas/15 pb-8 md:grid-cols-[1fr_auto] md:gap-16">
          <div>
            <BrandMark inverse />
            <p className="mt-3 max-w-xs text-sm leading-6 text-canvas/75">
              Better software decisions for the business you actually run.
            </p>
          </div>
          <nav
            aria-label="Footer navigation"
            className="grid grid-cols-2 gap-x-10 gap-y-7 sm:grid-cols-3 md:gap-x-14"
          >
            {footerGroups.map((group) => (
              <div key={group.heading}>
                <p className="text-[0.6875rem] font-semibold uppercase tracking-[0.14em] text-canvas/55">
                  {group.heading}
                </p>
                {/* The links are 14px text, so their own box was 17px tall and
                    a thumb had 17px of vertical room to hit one. `min-h-6`
                    raises each hit box to the 24px WCAG 2.2 asks for, and
                    because it grows into the gap that was already there, the
                    footer keeps the height it was shrunk to. */}
                <ul className="mt-3 space-y-2">
                  {group.links.map((link) => (
                    <li key={link.to}>
                      {/* A mailto: is not a route. Handing it to Link makes
                          the router treat it as a path and navigate nowhere. */}
                      {link.to.startsWith('mailto:') ? (
                        <a className={footerLinkStyle} href={link.to}>
                          {link.label}
                        </a>
                      ) : (
                        <Link className={footerLinkStyle} to={link.to}>
                          {link.label}
                        </Link>
                      )}
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </nav>
        </div>
        <div className="flex flex-col gap-2 pt-5 text-[0.625rem] uppercase tracking-[0.12em] text-canvas/70 sm:flex-row sm:items-center sm:justify-between">
          <p>© {new Date().getFullYear()} UNSOLERO · Independent by design</p>
          <div className="flex flex-wrap items-center gap-x-5 gap-y-1">
            <p>Commission never changes ranking</p>
            <button
              className="min-h-6 underline decoration-canvas/35 underline-offset-4 hover:text-canvas focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-canvas"
              onClick={openAnalyticsConsentPreferences}
              type="button"
            >
              Analytics preferences
            </button>
          </div>
        </div>
      </Container>
    </footer>
  )
}
