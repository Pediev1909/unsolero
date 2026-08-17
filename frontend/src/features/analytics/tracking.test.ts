import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import {
  affiliateClickPath,
  completeOnboarding,
  sendAnalyticsEvent,
  startOnboarding,
} from './tracking'
import { setAnalyticsConsent } from './consent'

beforeEach(() => {
  setAnalyticsConsent('granted')
})

afterEach(() => {
  vi.unstubAllGlobals()
  window.sessionStorage.clear()
  window.localStorage.clear()
  window.history.replaceState({}, '', '/')
})

describe('affiliate and analytics tracking', () => {
  it('adds bounded first-party attribution to an offer path', () => {
    window.history.replaceState(
      {},
      '',
      '/products/demo?utm_campaign=strength_launch',
    )
    const path = affiliateClickPath(
      '/api/affiliate/click/97bfb760-6d09-4b96-8a39-d2bb16445537',
      'product_detail',
      '4ba7d524-9fd5-4d18-8c42-778c42d996f3',
    )
    const url = new URL(path, window.location.origin)
    expect(url.pathname).toBe(
      '/api/affiliate/click/97bfb760-6d09-4b96-8a39-d2bb16445537',
    )
    expect(url.searchParams.get('source')).toBe('product_detail')
    expect(url.searchParams.get('campaign')).toBe('strength_launch')
    expect(url.searchParams.get('traffic_source')).toBeNull()
    expect(url.searchParams.get('recommendation_id')).toBe(
      '4ba7d524-9fd5-4d18-8c42-778c42d996f3',
    )
    expect(url.searchParams.get('session_id')).toMatch(/^[0-9a-f-]{36}$/)
  })

  it('does not modify external or legacy destinations', () => {
    expect(
      affiliateClickPath('https://merchant.example/item', 'wishlist'),
    ).toBe('https://merchant.example/item')
    expect(affiliateClickPath('/api/out/link-id', 'wishlist')).toBe(
      '/api/out/link-id',
    )
  })

  it('sends only the typed event envelope', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', fetchMock)
    window.history.replaceState({}, '', '/products/demo?utm_source=editorial')
    await sendAnalyticsEvent('product_viewed', 'product_detail', {
      product_id: '97bfb760-6d09-4b96-8a39-d2bb16445537',
    })
    const [, options] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(typeof options.body).toBe('string')
    const body = JSON.parse(options.body as string) as Record<string, unknown>
    expect(body).toMatchObject({
      name: 'product_viewed',
      surface: 'product_detail',
      consent_state: 'granted',
      properties: {
        product_id: '97bfb760-6d09-4b96-8a39-d2bb16445537',
      },
    })
    expect(body.context).toEqual({
      page_path: '/products/demo',
      traffic_source: 'editorial',
    })
    expect(body).not.toHaveProperty('user_id')
  })

  it('does not send optional analytics after consent is declined', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    setAnalyticsConsent('denied')

    await sendAnalyticsEvent('page_view', 'route', {})

    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('pairs one onboarding start and completion per attempt', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', fetchMock)

    const firstID = startOnboarding()
    expect(startOnboarding()).toBe(firstID)
    completeOnboarding('complete')
    completeOnboarding('complete')
    await vi.waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2))

    const events = fetchMock.mock.calls.map(([, rawOptions]) => {
      const body = (rawOptions as RequestInit).body
      if (typeof body !== 'string') throw new Error('Expected a JSON body.')
      return JSON.parse(body) as {
        name: string
        properties: { onboarding_id: string }
      }
    })
    expect(events.map((event) => event.name).sort()).toEqual([
      'onboarding_completed',
      'onboarding_started',
    ])
    expect(
      events.every((event) => event.properties.onboarding_id === firstID),
    ).toBe(true)
  })
})
