import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Button } from '../components/ui/Button'
import { Container } from '../components/ui/Container'
import { Heading } from '../components/ui/Heading'
import { LoadingState } from '../components/ui/LoadingState'
import { unsubscribeFromNewsletter } from '../features/newsletter/api'
import { ApiError } from '../lib/api/client'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

// 'missing' is a separate state from 'invalid', which the confirmation page
// does not need: there the remedy for both is the same, so both show the
// sign-up form. Here the two differ. No token means the reader arrived without
// the link (typed the address, or a mail client that strips fragments), so the
// answer is where to find the link. A token the server does not know means the
// link itself is spent or mangled, and the only remedy left is a person.
type UnsubscribeState =
  | 'pending'
  | 'unsubscribed'
  | 'invalid'
  | 'missing'
  | 'unavailable'

const contactAddress = 'hello@unsolero.com'

export function NewsletterUnsubscribePage() {
  const location = useLocation()
  // The email puts the token in the fragment so it never reaches the edge's
  // access log; a query parameter is accepted too for clients that drop it.
  const token = useMemo(() => {
    const fromFragment = location.hash.slice(1)
    if (fromFragment) return fromFragment
    return new URLSearchParams(location.search).get('token') ?? ''
  }, [location.hash, location.search])
  const attempt = useRef<Promise<void> | null>(null)
  const [state, setState] = useState<UnsubscribeState>(
    token ? 'pending' : 'missing',
  )

  usePageMetadata({
    title: 'Unsubscribe from price alerts | UNSOLERO',
    description: 'Unsubscribe from the UNSOLERO price-change emails.',
    // A one-time link's landing page has nothing to rank.
    robots: 'noindex, follow',
  })

  useEffect(() => {
    if (!token || state !== 'pending') return
    attempt.current ??= unsubscribeFromNewsletter(token)
    let active = true
    void attempt.current
      .then(() => {
        if (!active) return
        window.history.replaceState(null, '', window.location.pathname)
        setState('unsubscribed')
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
                description="Removing your address from the list."
                title="Unsubscribing you"
              />
            </div>
          )}
          {state === 'unsubscribed' && (
            <div role="status">
              <Heading className="mt-5" level={1} size="title">
                You are unsubscribed.
              </Heading>
              <p className="mt-5 max-w-2xl text-base leading-7 text-ink/70">
                Nothing else will arrive. The address stays on a do-not-send
                list so a later sign-up cannot quietly add it back; to have it
                deleted outright instead, write to{' '}
                <a
                  className="underline underline-offset-4"
                  href={`mailto:${contactAddress}`}
                >
                  {contactAddress}
                </a>
                .
              </p>
              <Link
                className="mt-8 inline-block font-semibold underline underline-offset-4"
                to="/products"
              >
                Browse the software prices
              </Link>
            </div>
          )}
          {/* No sign-up form here. On the confirmation page a fresh address is
              exactly what the reader wants; on an unsubscribe page it is the
              opposite of what they came to do. */}
          {state === 'invalid' && (
            <div role="alert">
              <Heading className="mt-5" level={1} size="title">
                This link did not match a subscription.
              </Heading>
              <p className="mt-5 max-w-2xl text-base leading-7 text-ink/70">
                It may have been used already, or altered somewhere between the
                email and this page. If emails keep arriving, forward one to{' '}
                <a
                  className="underline underline-offset-4"
                  href={`mailto:${contactAddress}`}
                >
                  {contactAddress}
                </a>{' '}
                and the address will be removed by hand.
              </p>
            </div>
          )}
          {state === 'missing' && (
            <div role="alert">
              <Heading className="mt-5" level={1} size="title">
                This page needs the link from the email.
              </Heading>
              <p className="mt-5 max-w-2xl text-base leading-7 text-ink/70">
                The unsubscribe link at the foot of any UNSOLERO email carries
                the one-time token this page spends, so open it from there.
                Nothing about your subscription has changed. If you cannot find
                an email, write to{' '}
                <a
                  className="underline underline-offset-4"
                  href={`mailto:${contactAddress}`}
                >
                  {contactAddress}
                </a>
                .
              </p>
            </div>
          )}
          {state === 'unavailable' && (
            <div role="alert">
              <Heading className="mt-5" level={1} size="title">
                We could not unsubscribe you right now.
              </Heading>
              <p className="mt-5 max-w-2xl text-base leading-7 text-ink/70">
                Nothing about your subscription has changed and the link is
                still valid — unsubscribe links do not expire. Try again in a
                moment.
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
