import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { renderWithProviders } from '../test/renderWithProviders'
import { NewsletterUnsubscribePage } from './NewsletterUnsubscribePage'

const unsubscriptionsPath = '/api/newsletter/unsubscriptions'

function jsonResponse(status: number, body?: unknown) {
  return new Response(body === undefined ? null : JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

// The page renders the site header, which loads categories for its menu, so
// fetch is routed by path: the unsubscribe call gets the scripted answers and
// everything else a harmless 404.
function stubFetch(...unsubscriptions: Response[]) {
  const fetchMock = vi.fn<
    (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>
  >((input) => {
    const url =
      typeof input === 'string'
        ? input
        : input instanceof URL
          ? input.href
          : input.url
    if (url === unsubscriptionsPath) {
      return Promise.resolve(
        unsubscriptions.shift() ??
          jsonResponse(500, {
            error: { code: 'unexpected', message: 'No scripted response.' },
          }),
      )
    }
    return Promise.resolve(
      jsonResponse(404, { error: { code: 'not_found', message: 'No route.' } }),
    )
  })
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

function unsubscribeCalls(fetchMock: ReturnType<typeof stubFetch>) {
  return fetchMock.mock.calls.filter(([input]) => input === unsubscriptionsPath)
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('NewsletterUnsubscribePage', () => {
  it('spends the fragment token once on mount', async () => {
    const fetchMock = stubFetch(jsonResponse(204))
    renderWithProviders(<NewsletterUnsubscribePage />, {
      route: '/newsletter/unsubscribe#raw-token',
    })

    expect(await screen.findByRole('status')).toHaveTextContent(
      /you are unsubscribed/i,
    )
    const calls = unsubscribeCalls(fetchMock)
    expect(calls).toHaveLength(1)
    const init = calls[0]?.[1] as RequestInit
    expect(init.method).toBe('POST')
    expect(JSON.parse(init.body as string) as unknown).toEqual({
      token: 'raw-token',
    })
  })

  it('accepts the token as a query parameter too', async () => {
    const fetchMock = stubFetch(jsonResponse(204))
    renderWithProviders(<NewsletterUnsubscribePage />, {
      route: '/newsletter/unsubscribe?token=query-token',
    })

    await screen.findByRole('status')
    const init = unsubscribeCalls(fetchMock)[0]?.[1] as RequestInit
    expect(JSON.parse(init.body as string) as unknown).toEqual({
      token: 'query-token',
    })
  })

  it('promises that nothing else arrives, and never offers to sign up again', async () => {
    stubFetch(jsonResponse(204))
    renderWithProviders(<NewsletterUnsubscribePage />, {
      route: '/newsletter/unsubscribe#raw-token',
    })

    expect(await screen.findByRole('status')).toHaveTextContent(
      /nothing else will arrive/i,
    )
    // Putting the sign-up form on this page would undo the click that got the
    // reader here, so no state of it may render one.
    expect(
      screen.queryByRole('button', { name: /send me the changes/i }),
    ).not.toBeInTheDocument()
  })

  it('explains an unknown link without inviting a re-subscription', async () => {
    stubFetch(
      jsonResponse(400, {
        error: {
          code: 'invalid_token',
          message: 'This link is invalid or has expired.',
        },
      }),
    )
    renderWithProviders(<NewsletterUnsubscribePage />, {
      route: '/newsletter/unsubscribe#stale',
    })

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent(/did not match a subscription/i)
    expect(alert).toHaveTextContent(/used already/i)
    expect(
      within(alert).queryByRole('button', { name: /send me the changes/i }),
    ).not.toBeInTheDocument()
    expect(
      within(alert).getByRole('link', { name: /hello@unsolero\.com/i }),
    ).toHaveAttribute('href', 'mailto:hello@unsolero.com')
  })

  it('asks for the link from the email when the token is missing, without calling the API', async () => {
    const fetchMock = stubFetch()
    renderWithProviders(<NewsletterUnsubscribePage />, {
      route: '/newsletter/unsubscribe',
    })

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent(/needs the link from the email/i)
    // A missing token is not a failed unsubscribe: nothing may be claimed
    // about an address the page was never given.
    expect(alert).toHaveTextContent(/nothing about your subscription has/i)
    expect(unsubscribeCalls(fetchMock)).toHaveLength(0)
  })

  it('keeps the link alive when the service fails and lets the reader retry', async () => {
    const fetchMock = stubFetch(
      jsonResponse(500, {
        error: { code: 'newsletter_unavailable', message: 'Try later.' },
      }),
      jsonResponse(204),
    )
    const user = userEvent.setup()
    renderWithProviders(<NewsletterUnsubscribePage />, {
      route: '/newsletter/unsubscribe#raw-token',
    })

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent(/could not unsubscribe you/i)
    await user.click(within(alert).getByRole('button', { name: /try again/i }))

    expect(await screen.findByRole('status')).toHaveTextContent(
      /you are unsubscribed/i,
    )
    expect(unsubscribeCalls(fetchMock)).toHaveLength(2)
  })

  it('asks search engines not to index the landing page', async () => {
    stubFetch(jsonResponse(204))
    renderWithProviders(<NewsletterUnsubscribePage />, {
      route: '/newsletter/unsubscribe#raw-token',
    })

    await screen.findByRole('status')
    expect(
      document.querySelector('meta[name="robots"]')?.getAttribute('content'),
    ).toBe('noindex, follow')
    expect(document.title).toBe('Unsubscribe from price alerts | UNSOLERO')
  })
})
