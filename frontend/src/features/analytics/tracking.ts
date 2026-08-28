import { apiRequest } from '../../lib/api/client'
import { analyticsConsentPolicyVersion, getAnalyticsConsent } from './consent'

const sessionStorageKey = 'rigmark:analytics-session:v1'
const attributionStorageKey = 'rigmark:analytics-attribution:v1'
const onboardingStorageKey = 'rigmark:onboarding-attempt:v1'
const onboardingStartedKey = 'rigmark:onboarding-started:v1'
const onboardingCompletedKey = 'rigmark:onboarding-completed:v1'
const tokenPattern = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/
const uuidPattern =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/
let memorySessionID: string | undefined

type AnalyticsEventMap = {
  page_view: Record<string, never>
  onboarding_started: { onboarding_id: string }
  onboarding_completed: {
    onboarding_id: string
    outcome: 'complete' | 'no_suitable_products'
  }
  recommendation_generated: {
    status: 'complete' | 'no_suitable_products'
    persistence: 'account' | 'browser'
  }
  product_viewed: { product_id: string }
  product_saved: {
    product_id: string
    persistence: 'account' | 'browser'
  }
  comparison_created: {
    product_count: number
    persistence: 'account' | 'browser'
  }
  setup_saved: {
    setup_id: string
    persistence: 'account' | 'browser'
  }
}

type AnalyticsContext = {
  page_path: string
  traffic_source?: string
  traffic_medium?: string
  campaign?: string
  referrer_host?: string
}

type AnalyticsEnvelope = {
  [Name in keyof AnalyticsEventMap]: {
    name: Name
    event_id: string
    surface: string
    session_id: string
    consent_version: string
    properties: AnalyticsEventMap[Name]
    context: AnalyticsContext
  }
}[keyof AnalyticsEventMap]

export interface AnalyticsProvider {
  send(event: AnalyticsEnvelope): Promise<void>
}

class FirstPartyAnalyticsProvider implements AnalyticsProvider {
  send(event: AnalyticsEnvelope) {
    return apiRequest(
      '/analytics/events',
      { method: 'POST', body: event },
      () => undefined,
    )
  }
}

const firstPartyProvider = new FirstPartyAnalyticsProvider()
const dispatch = createAnalyticsDispatcher([firstPartyProvider])

export type AffiliateSource =
  | 'product_detail'
  | 'wishlist'
  | 'recommendation'
  | 'comparison'
  | 'setup'
  | 'promotion'

export function createAnalyticsDispatcher(
  providers: readonly AnalyticsProvider[],
) {
  return async (event: AnalyticsEnvelope) => {
    const results = await Promise.allSettled(
      providers.map((provider) => provider.send(event)),
    )
    if (results.every((result) => result.status === 'rejected')) {
      throw new Error('No analytics provider accepted the event.')
    }
  }
}

export function trackEvent<Name extends keyof AnalyticsEventMap>(
  name: Name,
  surface: string,
  properties: AnalyticsEventMap[Name],
) {
  if (getAnalyticsConsent() !== 'granted') return
  void dispatch(envelope(name, surface, properties)).catch(() => {
    // Analytics must never block or change the underlying user action.
  })
}

export function sendAnalyticsEvent<Name extends keyof AnalyticsEventMap>(
  name: Name,
  surface: string,
  properties: AnalyticsEventMap[Name],
) {
  if (getAnalyticsConsent() !== 'granted') return Promise.resolve()
  return firstPartyProvider.send(envelope(name, surface, properties))
}

function envelope<Name extends keyof AnalyticsEventMap>(
  name: Name,
  surface: string,
  properties: AnalyticsEventMap[Name],
): AnalyticsEnvelope {
  return {
    event_id: crypto.randomUUID(),
    name,
    surface,
    session_id: analyticsSessionID(),
    consent_version: analyticsConsentPolicyVersion,
    properties,
    context: analyticsContext(),
  } as AnalyticsEnvelope
}

export function startOnboarding() {
  let onboardingID = storedUUID(onboardingStorageKey)
  if (!onboardingID || storedUUID(onboardingCompletedKey) === onboardingID) {
    onboardingID = crypto.randomUUID()
    safeSessionSet(onboardingStorageKey, onboardingID)
  }
  if (storedUUID(onboardingStartedKey) !== onboardingID) {
    safeSessionSet(onboardingStartedKey, onboardingID)
    trackEvent('onboarding_started', 'recommendation', {
      onboarding_id: onboardingID,
    })
  }
  return onboardingID
}

export function completeOnboarding(
  outcome: 'complete' | 'no_suitable_products',
) {
  const onboardingID = storedUUID(onboardingStorageKey)
  if (!onboardingID || storedUUID(onboardingCompletedKey) === onboardingID)
    return
  safeSessionSet(onboardingCompletedKey, onboardingID)
  trackEvent('onboarding_completed', 'recommendation', {
    onboarding_id: onboardingID,
    outcome,
  })
}

export function affiliateClickPath(
  purchasePath: string,
  source: AffiliateSource,
  recommendationID?: string | null,
) {
  const destination = new URL(purchasePath, window.location.origin)
  if (
    destination.origin !== window.location.origin ||
    (!destination.pathname.startsWith('/api/affiliate/click/') &&
      !destination.pathname.startsWith('/api/affiliate/promotion/'))
  ) {
    return purchasePath
  }
  destination.searchParams.set('source', source)
  destination.searchParams.set('session_id', analyticsSessionID())
  const attribution = analyticsAttribution()
  setQueryAttribution(destination, 'campaign', attribution.campaign)
  setQueryAttribution(destination, 'traffic_source', attribution.traffic_source)
  setQueryAttribution(destination, 'traffic_medium', attribution.traffic_medium)
  if (recommendationID && uuidPattern.test(recommendationID)) {
    destination.searchParams.set('recommendation_id', recommendationID)
  }
  return `${destination.pathname}${destination.search}`
}

function setQueryAttribution(
  destination: URL,
  key: string,
  value: string | undefined,
) {
  if (value && tokenPattern.test(value))
    destination.searchParams.set(key, value)
}

function analyticsContext(): AnalyticsContext {
  return { page_path: window.location.pathname, ...analyticsAttribution() }
}

function analyticsAttribution(): Omit<AnalyticsContext, 'page_path'> {
  const stored = safeSessionGet(attributionStorageKey)
  if (stored) {
    try {
      const parsed = parseAttribution(
        JSON.parse(stored) as Record<string, unknown>,
      )
      // A visitor who opened the site directly and only afterwards followed a
      // campaign link would otherwise keep the empty first touch for the whole
      // tab session, losing the campaign that actually brought them back.
      // First touch still wins; an empty one is not a touch.
      if (Object.keys(parsed).length > 0) return parsed
    } catch {
      // Replace invalid browser state with current bounded attribution.
    }
  }
  const search = new URLSearchParams(window.location.search)
  const current = parseAttribution({
    traffic_source: search.get('utm_source'),
    traffic_medium: search.get('utm_medium'),
    campaign: search.get('utm_campaign'),
    referrer_host: externalReferrerHost(),
  })
  if (Object.keys(current).length > 0) {
    safeSessionSet(attributionStorageKey, JSON.stringify(current))
  }
  return current
}

function parseAttribution(value: Record<string, unknown>) {
  const result: Omit<AnalyticsContext, 'page_path'> = {}
  for (const key of ['traffic_source', 'traffic_medium', 'campaign'] as const) {
    const candidate = value[key]
    if (typeof candidate === 'string' && tokenPattern.test(candidate)) {
      result[key] = candidate
    }
  }
  const referrer = value.referrer_host
  if (typeof referrer === 'string' && referrer.length <= 253) {
    result.referrer_host = referrer.toLowerCase()
  }
  return result
}

function externalReferrerHost() {
  try {
    const referrer = new URL(document.referrer)
    return referrer.origin === window.location.origin
      ? undefined
      : referrer.hostname.toLowerCase()
  } catch {
    return undefined
  }
}

export function analyticsSessionID() {
  if (memorySessionID) return memorySessionID
  const stored = safeSessionGet(sessionStorageKey)
  if (stored && uuidPattern.test(stored)) {
    memorySessionID = stored
    return stored
  }
  memorySessionID = crypto.randomUUID()
  safeSessionSet(sessionStorageKey, memorySessionID)
  return memorySessionID
}

function storedUUID(key: string) {
  const value = safeSessionGet(key)
  return value && uuidPattern.test(value) ? value : undefined
}

function safeSessionGet(key: string) {
  try {
    return window.sessionStorage.getItem(key)
  } catch {
    return null
  }
}

function safeSessionSet(key: string, value: string) {
  try {
    window.sessionStorage.setItem(key, value)
  } catch {
    // Memory-backed session identity remains available when storage is blocked.
  }
}
