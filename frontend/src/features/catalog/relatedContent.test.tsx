import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useProductEditorial } from './relatedContent'

const entry = {
  id: '7da06367-69db-4fac-b1c4-e878b64a6a60',
  type: 'comparison',
  title: 'Mailchimp vs ActiveCampaign',
  slug: 'mailchimp-vs-activecampaign',
  path: '/compare/mailchimp-vs-activecampaign',
  description: 'Two email tools, one budget.',
  hero_image: {
    url: '/images/saas-agency-stack.svg',
    alt_text: 'Diagram of a connected software stack',
    is_primary: false,
  },
  author_name: 'Andon Pediev',
  published_at: '2026-08-20T09:00:00Z',
  updated_at: '2026-08-26T09:00:00Z',
  covered: [
    { name: 'Mailchimp Standard', price_minor: 2000, currency: 'USD' },
    { name: 'ActiveCampaign Plus', price_minor: 4900, currency: 'USD' },
  ],
}

function wrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  })
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>
  }
}

function jsonResponse(body: unknown) {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  })
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('useProductEditorial', () => {
  it('asks for the entries that reference the product and parses them as content summaries', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse([entry]))
    vi.stubGlobal('fetch', fetchMock)

    const { result } = renderHook(
      () => useProductEditorial('mailchimp-standard'),
      { wrapper: wrapper() },
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(fetchMock.mock.calls[0]?.[0]).toBe(
      '/api/content?product=mailchimp-standard',
    )
    expect(result.current.data).toHaveLength(1)
    expect(result.current.data?.[0]?.path).toBe(
      '/compare/mailchimp-vs-activecampaign',
    )
    expect(result.current.data?.[0]?.covered).toHaveLength(2)
  })

  // A payload that is not a list of content summaries is an error, not a
  // list of undefined titles rendered as empty cards.
  it('refuses a payload that does not match the content summary shape', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(jsonResponse([{ id: 'not-a-summary' }])),
    )

    const { result } = renderHook(() => useProductEditorial('anything'), {
      wrapper: wrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))
  })

  it('does not fetch without a slug', () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)

    const { result } = renderHook(() => useProductEditorial(''), {
      wrapper: wrapper(),
    })

    expect(result.current.fetchStatus).toBe('idle')
    expect(fetchMock).not.toHaveBeenCalled()
  })
})
