import { ArrowLeft } from 'lucide-react'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { ButtonLink } from '../components/ui/ButtonLink'
import { Heading } from '../components/ui/Heading'

export function NotFoundPage() {
  return (
    // A reader who lands here arrived from a stale link or a typo, so this is
    // the page most in need of a way onward. It previously filled the viewport
    // with an apology and a single link home, and offered no navigation at all.
    <>
      <SiteHeader />
      <main className="grid place-items-center bg-canvas px-6 py-24" id="main-content">
        <div className="max-w-xl text-center">
          <p className="eyebrow">404 / Page not found</p>
          <Heading className="mt-5" level={1} size="display">
            This path needs a rethink.
          </Heading>
          <p className="mx-auto mt-6 max-w-md leading-7 text-ink/70">
            The page you requested does not exist or has moved. The catalog and
            the guides are the usual places to pick the thread back up.
          </p>
          <div className="mt-8 flex flex-wrap justify-center gap-3">
            <ButtonLink to="/">
              <ArrowLeft aria-hidden="true" size={16} />
              Back to UNSOLERO
            </ButtonLink>
            <ButtonLink to="/products" variant="secondary">
              Browse software
            </ButtonLink>
          </div>
        </div>
      </main>
      <SiteFooter />
    </>
  )
}
