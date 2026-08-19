import { Link } from 'react-router-dom'

import { Container } from '../ui/Container'
import { BrandMark } from './BrandMark'
import { openAnalyticsConsentPreferences } from '../../features/analytics/consent'

// Grouped rather than listed. A single alphabet-soup column of ten links makes
// the reader scan all of them to find one; under a heading they can skip to the
// group they want. The grouping is editorial and deliberately differs from the
// header's navigation, which carries only the primary journeys.
const footerGroups: { heading: string; links: { label: string; to: string }[] }[] = [
  {
    heading: 'Decide',
    links: [
      { label: 'Build my setup', to: '/build' },
      { label: 'Software', to: '/products' },
      { label: 'Compare', to: '/compare' },
      { label: 'Wishlist', to: '/wishlist' },
    ],
  },
  {
    heading: 'Learn',
    links: [
      { label: 'How it works', to: '/#method' },
      { label: 'Guides', to: '/guides' },
      { label: 'Articles', to: '/articles' },
      { label: 'About', to: '/about' },
    ],
  },
  {
    heading: 'Account',
    links: [
      { label: 'Sign in', to: '/login' },
      { label: 'Create account', to: '/register' },
      { label: 'Privacy', to: '/privacy' },
      { label: 'Affiliate disclosure', to: '/affiliate-disclosure' },
    ],
  },
]

export function SiteFooter() {
  return (
    <footer className="border-t border-canvas/10 bg-ink py-10 text-canvas">
      <Container>
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
                      <Link
                        className="flex min-h-6 items-center text-sm text-canvas/75 transition-colors duration-150 hover:text-canvas"
                        to={link.to}
                      >
                        {link.label}
                      </Link>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </nav>
        </div>
        <div className="flex flex-col gap-2 pt-5 text-[0.625rem] uppercase tracking-[0.12em] text-canvas/70 sm:flex-row sm:items-center sm:justify-between">
          <p>
            © {new Date().getFullYear()} UNSOLERO · Independent by design
          </p>
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
