import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import {
  affiliateClickPath,
  captureLandingAttribution,
  completeOnboarding,
  sendAnalyticsEvent,
  startOnboarding,
} from './tracking'
import { getAnalyticsConsent, setAnalyticsConsent } from './consent'

beforeEach(async () => {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          state: 'granted',
          policy_version: 'analytics-v1',
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    ),
  )
  await setAnalyticsConsent('granted')
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

  it('adds first-party attribution to a standalone promotion path', () => {
    window.history.replaceState({}, '', '/offers/funnel-hacking-secrets')
    const path = affiliateClickPath(
      '/api/affiliate/promotion/funnel-hacking-secrets-webinar',
      'promotion',
    )
    const url = new URL(path, window.location.origin)
    expect(url.pathname).toBe(
      '/api/affiliate/promotion/funnel-hacking-secrets-webinar',
    )
    expect(url.searchParams.get('source')).toBe('promotion')
    expect(url.searchParams.get('session_id')).toMatch(/^[0-9a-f-]{36}$/)
    expect(url.searchParams.get('recommendation_id')).toBeNull()
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
      consent_version: 'analytics-v1',
      properties: {
        product_id: '97bfb760-6d09-4b96-8a39-d2bb16445537',
      },
    })
    expect(body.context).toEqual({
      page_path: '/products/demo',
      traffic_source: 'editorial',
    })
    expect(body).not.toHaveProperty('user_id')
    expect(body.event_id).toMatch(/^[0-9a-f-]{36}$/)
  })

  it('does not send optional analytics after consent is declined', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          state: 'denied',
          policy_version: 'analytics-v1',
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )
    vi.stubGlobal('fetch', fetchMock)
    await setAnalyticsConsent('denied')
    fetchMock.mockClear()

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

  // A visitor who opens the site directly and only afterwards follows a campaign
  // link kept the empty first touch for the whole tab session, so every click
  // they made was recorded with no traffic source at all.
  it('keeps a later campaign when the first page view carried no attribution', () => {
    window.history.replaceState({}, '', '/')
    affiliateClickPath('/api/affiliate/click/offer-a', 'product_detail')

    window.history.replaceState({}, '', '/?utm_source=tiktok&utm_medium=social')
    const url = new URL(
      affiliateClickPath('/api/affiliate/click/offer-a', 'product_detail'),
      window.location.origin,
    )

    expect(url.searchParams.get('traffic_source')).toBe('tiktok')
    expect(url.searchParams.get('traffic_medium')).toBe('social')
  })

  it('keeps the first campaign when a later page view carries none', () => {
    window.history.replaceState({}, '', '/?utm_source=tiktok')
    affiliateClickPath('/api/affiliate/click/offer-a', 'product_detail')

    window.history.replaceState({}, '', '/products/demo')
    const url = new URL(
      affiliateClickPath('/api/affiliate/click/offer-a', 'product_detail'),
      window.location.origin,
    )

    expect(url.searchParams.get('traffic_source')).toBe('tiktok')
  })
})

describe('landing attribution captured before consent', () => {
  it('keeps the campaign across a client-side navigation without minting a session', () => {
    // No consent decision at all — the state a first-time visitor is in.
    window.localStorage.clear()
    expect(getAnalyticsConsent()).toBe('unknown')
    window.history.replaceState(
      {},
      '',
      '/guides/mailchimp-alternatives?utm_source=tiktok&utm_medium=bio&utm_campaign=2026-09-mailchimp-250',
    )
    captureLandingAttribution()

    expect(
      window.sessionStorage.getItem('rigmark:analytics-session:v1'),
    ).toBeNull()
    const stored = JSON.parse(
      window.sessionStorage.getItem('rigmark:analytics-attribution:v1') ?? '{}',
    ) as Record<string, string>
    expect(stored.traffic_source).toBe('tiktok')
    expect(stored.traffic_medium).toBe('bio')
    expect(stored.campaign).toBe('2026-09-mailchimp-250')

    // The visitor moves on inside the application; the URL no longer says
    // where they came from, and the vendor link must still know.
    window.history.replaceState({}, '', '/products/mailerlite-comfort')
    const path = affiliateClickPath(
      '/api/affiliate/click/97bfb760-6d09-4b96-8a39-d2bb16445537',
      'product_detail',
    )
    const url = new URL(path, window.location.origin)
    expect(url.searchParams.get('campaign')).toBe('2026-09-mailchimp-250')
    expect(url.searchParams.get('traffic_source')).toBe('tiktok')
    expect(url.searchParams.get('traffic_medium')).toBe('bio')
  })

  it('stores nothing when the landing URL carries no attribution', () => {
    window.history.replaceState({}, '', '/products')
    captureLandingAttribution()
    expect(
      window.sessionStorage.getItem('rigmark:analytics-attribution:v1'),
    ).toBeNull()
  })

  it('discards tokens that do not fit the bounded pattern', () => {
    window.history.replaceState(
      {},
      '',
      '/?utm_source=<script>&utm_campaign=ok-campaign',
    )
    captureLandingAttribution()
    const stored = JSON.parse(
      window.sessionStorage.getItem('rigmark:analytics-attribution:v1') ?? '{}',
    ) as Record<string, string>
    expect(stored.traffic_source).toBeUndefined()
    expect(stored.campaign).toBe('ok-campaign')
  })
})
