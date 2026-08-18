import { apiRequest } from '../../lib/api/client'

export type AnalyticsConsent = 'unknown' | 'granted' | 'denied'
export const analyticsConsentPolicyVersion = 'analytics-v1'

const consentStorageKey = `unsolero:analytics-consent:${analyticsConsentPolicyVersion}`
export const analyticsConsentChangedEvent = 'unsolero:analytics-consent-changed'
export const openAnalyticsConsentEvent = 'unsolero:analytics-consent-open'

type ConsentResponse = {
  state: 'unknown' | 'granted' | 'denied' | 'withdrawn'
  policy_version?: string
}

export function getAnalyticsConsent(): AnalyticsConsent {
  try {
    const value = window.localStorage.getItem(consentStorageKey)
    return value === 'granted' || value === 'denied' ? value : 'unknown'
  } catch {
    return 'unknown'
  }
}

export async function setAnalyticsConsent(
  consent: Exclude<AnalyticsConsent, 'unknown'>,
  source: 'banner' | 'preferences' | 'account_sync' = 'preferences',
) {
  const persisted = await apiRequest(
    '/analytics/consent',
    {
      method: 'PUT',
      body: {
        state: consent,
        policy_version: analyticsConsentPolicyVersion,
        source,
      },
    },
    parseConsent,
  )
  const localState = persisted.state === 'granted' ? 'granted' : 'denied'
  try {
    window.localStorage.setItem(consentStorageKey, localState)
  } catch {
    // A blocked preference store leaves future analytics disabled by default.
  }
  window.dispatchEvent(
    new CustomEvent(analyticsConsentChangedEvent, { detail: localState }),
  )
  return localState
}

export async function synchronizeAnalyticsConsentAfterAuthentication() {
  const consent = getAnalyticsConsent()
  if (consent === 'unknown') return
  await setAnalyticsConsent(consent, 'account_sync')
  if (consent === 'granted') {
    await apiRequest(
      '/analytics/identity/claim',
      { method: 'POST' },
      () => undefined,
    ).catch(() => {
      // Claiming is privacy-enhancing but optional. Server-side account consent
      // still governs all new authenticated events if the old browser identity
      // is absent, already claimed, or deliberately revoked.
    })
  }
}

function parseConsent(value: unknown): ConsentResponse {
  if (!value || typeof value !== 'object') throw new Error('Invalid consent')
  const state = (value as Record<string, unknown>).state
  const policy = (value as Record<string, unknown>).policy_version
  if (
    state !== 'unknown' &&
    state !== 'granted' &&
    state !== 'denied' &&
    state !== 'withdrawn'
  ) {
    throw new Error('Invalid consent state')
  }
  if (policy !== undefined && typeof policy !== 'string') {
    throw new Error('Invalid consent policy')
  }
  return { state, policy_version: policy }
}

export function openAnalyticsConsentPreferences() {
  window.dispatchEvent(new Event(openAnalyticsConsentEvent))
}
