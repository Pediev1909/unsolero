import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { renderWithProviders } from '../test/renderWithProviders'
import { NewsletterConfirmPage } from './NewsletterConfirmPage'

const confirmationsPath = '/api/newsletter/confirmations'

function jsonResponse(status: number, body?: unknown) {
  return new Response(body === undefined ? null : JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

// The page renders the site header, which loads categories for its menu, so
// fetch is routed by path: the confirmation call gets the scripted answers and
// everything else a harmless 404.
function stubFetch(...confirmations: Response[]) {
  const fetchMock = vi.fn<
    (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>
  >((input) => {
    const url =
      typeof input === 'string'
        ? input
        : input instanceof URL
          ? input.href
          : input.url
    if (url === confirmationsPath) {
      return Promise.resolve(
        confirmations.shift() ??
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

function confirmationCalls(fetchMock: ReturnType<typeof stubFetch>) {
  return fetchMock.mock.calls.filter(([input]) => input === confirmationsPath)
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('NewsletterConfirmPage', () => {
  it('confirms the fragment token once on mount', async () => {
    const fetchMock = stubFetch(jsonResponse(204))
    renderWithProviders(<NewsletterConfirmPage />, {
      route: '/newsletter/confirm#raw-token',
    })

    expect(await screen.findByRole('status')).toHaveTextContent(/on the list/i)
    const calls = confirmationCalls(fetchMock)
    expect(calls).toHaveLength(1)
    const init = calls[0]?.[1] as RequestInit
    expect(init.method).toBe('POST')
    expect(JSON.parse(init.body as string) as unknown).toEqual({
      token: 'raw-token',
    })
  })

  it('accepts the token as a query parameter too', async () => {
    const fetchMock = stubFetch(jsonResponse(204))
    renderWithProviders(<NewsletterConfirmPage />, {
      route: '/newsletter/confirm?token=query-token',
    })

    await screen.findByRole('status')
    const init = confirmationCalls(fetchMock)[0]?.[1] as RequestInit
    expect(JSON.parse(init.body as string) as unknown).toEqual({
      token: 'query-token',
    })
  })

  it('explains an expired link and offers a fresh sign-up', async () => {
    stubFetch(
      jsonResponse(400, {
        error: {
          code: 'invalid_token',
          message: 'This link is invalid or has expired.',
        },
      }),
    )
    renderWithProviders(<NewsletterConfirmPage />, {
      route: '/newsletter/confirm#stale',
    })

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent(/expired or was already used/i)
    expect(
      within(alert).getByRole('button', { name: /send me the changes/i }),
    ).toBeInTheDocument()
  })

  it('treats a missing token as an invalid link without calling the API', async () => {
    const fetchMock = stubFetch()
    renderWithProviders(<NewsletterConfirmPage />, {
      route: '/newsletter/confirm',
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      /expired or was already used/i,
    )
    expect(confirmationCalls(fetchMock)).toHaveLength(0)
  })

  it('keeps the link alive when the service fails and lets the reader retry', async () => {
    const fetchMock = stubFetch(
      jsonResponse(500, {
        error: { code: 'newsletter_unavailable', message: 'Try later.' },
      }),
      jsonResponse(204),
    )
    const user = userEvent.setup()
    renderWithProviders(<NewsletterConfirmPage />, {
      route: '/newsletter/confirm#raw-token',
    })

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent(/could not confirm/i)
    await user.click(within(alert).getByRole('button', { name: /try again/i }))

    expect(await screen.findByRole('status')).toHaveTextContent(/on the list/i)
    expect(confirmationCalls(fetchMock)).toHaveLength(2)
  })

  it('asks search engines not to index the landing page', async () => {
    stubFetch(jsonResponse(204))
    renderWithProviders(<NewsletterConfirmPage />, {
      route: '/newsletter/confirm#raw-token',
    })

    await screen.findByRole('status')
    expect(
      document.querySelector('meta[name="robots"]')?.getAttribute('content'),
    ).toBe('noindex, follow')
    expect(document.title).toBe('Confirm your price alerts | UNSOLERO')
  })
})
