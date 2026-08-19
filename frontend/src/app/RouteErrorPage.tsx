import { AlertTriangle, RefreshCw } from 'lucide-react'
import { isRouteErrorResponse, Link, useRouteError } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Button } from '../components/ui/Button'
import { Container } from '../components/ui/Container'
import { Heading } from '../components/ui/Heading'

export function RouteErrorPage() {
  const error = useRouteError()
  const notFound = isRouteErrorResponse(error) && error.status === 404

  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <Container className="py-20 sm:py-28">
          <div
            className="mx-auto max-w-2xl border-y border-ink/15 py-12 text-center"
            role="alert"
          >
            <AlertTriangle
              aria-hidden="true"
              className="mx-auto text-bronze-dark"
              size={28}
            />
            <p className="eyebrow mt-6">
              {notFound ? 'Page unavailable' : 'Application error'}
            </p>
            <Heading className="mt-4" level={1} size="title">
              {notFound
                ? 'We could not find this page.'
                : 'Something interrupted this page.'}
            </Heading>
            <p className="mx-auto mt-5 max-w-lg text-sm leading-6 text-ink/70">
              {notFound
                ? 'The address may have changed or the page may no longer be published.'
                : 'Your account and saved data have not been changed. Reload the page to try again.'}
            </p>
            <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
              {!notFound && (
                <Button onClick={() => window.location.reload()}>
                  <RefreshCw aria-hidden="true" size={16} /> Reload page
                </Button>
              )}
              <Link
                className="inline-flex min-h-11 items-center justify-center border border-ink/20 px-5 text-sm font-semibold hover:border-ink"
                to="/"
              >
                Return home
              </Link>
            </div>
          </div>
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
