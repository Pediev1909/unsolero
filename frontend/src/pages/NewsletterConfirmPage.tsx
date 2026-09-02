import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Button } from '../components/ui/Button'
import { Container } from '../components/ui/Container'
import { Heading } from '../components/ui/Heading'
import { LoadingState } from '../components/ui/LoadingState'
import { confirmNewsletterSubscription } from '../features/newsletter/api'
import { NewsletterForm } from '../features/newsletter/components/NewsletterForm'
import { ApiError } from '../lib/api/client'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

type ConfirmationState = 'pending' | 'confirmed' | 'invalid' | 'unavailable'

export function NewsletterConfirmPage() {
  const location = useLocation()
  // The email puts the token in the fragment so it never reaches the edge's
  // access log; a query parameter is accepted too for clients that drop it.
  const token = useMemo(() => {
    const fromFragment = location.hash.slice(1)
    if (fromFragment) return fromFragment
    return new URLSearchParams(location.search).get('token') ?? ''
  }, [location.hash, location.search])
  const attempt = useRef<Promise<void> | null>(null)
  const [state, setState] = useState<ConfirmationState>(
    token ? 'pending' : 'invalid',
  )

  usePageMetadata({
    title: 'Confirm your price alerts | UNSOLERO',
    description: 'Confirm your UNSOLERO newsletter subscription.',
    // A one-time link's landing page has nothing to rank.
    robots: 'noindex, follow',
  })

  useEffect(() => {
    if (!token || state !== 'pending') return
    attempt.current ??= confirmNewsletterSubscription(token)
    let active = true
    void attempt.current
      .then(() => {
        if (!active) return
        window.history.replaceState(null, '', window.location.pathname)
        setState('confirmed')
      })
      .catch((error: unknown) => {
        if (!active) return
        attempt.current = null
        setState(
          error instanceof ApiError && error.code === 'invalid_token'
            ? 'invalid'
            : 'unavailable',
        )
      })
    return () => {
      active = false
    }
  }, [state, token])

  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <Container className="py-16 sm:py-24" size="reading">
          <p className="eyebrow">Price alerts</p>
          {state === 'pending' && (
            <div className="mt-8">
              <LoadingState
                description="Checking the one-time confirmation link."
                title="Confirming your address"
              />
            </div>
          )}
          {state === 'confirmed' && (
            <div role="status">
              <Heading className="mt-5" level={1} size="title">
                You&rsquo;re on the list.
              </Heading>
              <p className="mt-5 max-w-2xl text-base leading-7 text-ink/70">
                From now on you get one email when a price you follow changes,
                and nothing else. Every email carries a one-click unsubscribe
                link.
              </p>
              <Link
                className="mt-8 inline-block font-semibold underline underline-offset-4"
                to="/products"
              >
                Browse the software prices
              </Link>
            </div>
          )}
          {state === 'invalid' && (
            <div role="alert">
              <Heading className="mt-5" level={1} size="title">
                This link has expired or was already used.
              </Heading>
              <p className="mt-5 max-w-2xl text-base leading-7 text-ink/70">
                Confirmation links work once and for 48 hours. Enter your
                address again and we will send a fresh one.
              </p>
              <div className="mt-10 border-t border-ink/15 pt-8">
                <NewsletterForm source="newsletter-confirm" />
              </div>
            </div>
          )}
          {state === 'unavailable' && (
            <div role="alert">
              <Heading className="mt-5" level={1} size="title">
                We could not confirm your address right now.
              </Heading>
              <p className="mt-5 max-w-2xl text-base leading-7 text-ink/70">
                Nothing about your subscription has changed and the link is
                still valid. Try again in a moment.
              </p>
              <Button
                className="mt-8"
                onClick={() => setState('pending')}
                variant="secondary"
              >
                Try again
              </Button>
            </div>
          )}
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
