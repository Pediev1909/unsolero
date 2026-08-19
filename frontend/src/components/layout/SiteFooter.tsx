import { Link } from 'react-router-dom'

import { Container } from '../ui/Container'
import { BrandMark } from './BrandMark'
import { primaryNavigation } from './navigationItems'
import { openAnalyticsConsentPreferences } from '../../features/analytics/consent'

export function SiteFooter() {
  return (
    <footer className="border-t border-canvas/10 bg-ink py-12 text-canvas sm:py-16">
      <Container>
        <div className="grid gap-10 border-b border-canvas/15 pb-10 md:grid-cols-[1.2fr_0.8fr] md:gap-16">
          <div>
            <BrandMark inverse />
            <p className="mt-5 max-w-sm text-sm leading-6 text-canvas/60">
              Better software decisions for the business you actually run.
            </p>
          </div>
          <nav
            aria-label="Footer navigation"
            className="grid grid-cols-2 gap-x-8 gap-y-4 text-sm"
          >
            {primaryNavigation.map((item) => (
              <Link
                className="text-canvas/70 transition-colors duration-150 hover:text-canvas"
                key={item.to}
                to={item.to}
              >
                {item.label}
              </Link>
            ))}
            <Link
              className="text-canvas/70 transition-colors duration-150 hover:text-canvas"
              to="/guides"
            >
              Guides
            </Link>
            <Link
              className="text-canvas/70 transition-colors duration-150 hover:text-canvas"
              to="/articles"
            >
              Articles
            </Link>
            <Link
              className="text-canvas/70 transition-colors duration-150 hover:text-canvas"
              to="/privacy"
            >
              Privacy
            </Link>
            <Link
              className="text-canvas/70 transition-colors duration-150 hover:text-canvas"
              to="/affiliate-disclosure"
            >
              Affiliate disclosure
            </Link>
            <Link
              className="text-canvas/70 transition-colors duration-150 hover:text-canvas"
              to="/login"
            >
              Sign in
            </Link>
            <Link
              className="text-canvas/70 transition-colors duration-150 hover:text-canvas"
              to="/register"
            >
              Create account
            </Link>
          </nav>
        </div>
        <div className="flex flex-col gap-3 pt-6 text-[0.625rem] uppercase tracking-[0.14em] text-canvas/55 sm:flex-row sm:items-center sm:justify-between">
          <p>Independent by design</p>
          <div className="flex flex-wrap gap-x-5 gap-y-2">
            <button
              className="min-h-6 underline decoration-canvas/35 underline-offset-4 hover:text-canvas focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-canvas"
              onClick={openAnalyticsConsentPreferences}
              type="button"
            >
              Analytics preferences
            </button>
            <p>Affiliate commission never changes objective ranking.</p>
          </div>
        </div>
      </Container>
    </footer>
  )
}
