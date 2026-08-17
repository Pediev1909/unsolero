export type AnalyticsConsent = 'unknown' | 'granted' | 'denied'

const consentStorageKey = 'rigmark:analytics-consent:v1'
export const analyticsConsentChangedEvent = 'rigmark:analytics-consent-changed'
export const openAnalyticsConsentEvent = 'rigmark:analytics-consent-open'

export function getAnalyticsConsent(): AnalyticsConsent {
  try {
    const value = window.localStorage.getItem(consentStorageKey)
    return value === 'granted' || value === 'denied' ? value : 'unknown'
  } catch {
    return 'unknown'
  }
}

export function setAnalyticsConsent(
  consent: Exclude<AnalyticsConsent, 'unknown'>,
) {
  try {
    window.localStorage.setItem(consentStorageKey, consent)
  } catch {
    // A blocked preference store leaves analytics disabled by default.
  }
  window.dispatchEvent(
    new CustomEvent(analyticsConsentChangedEvent, { detail: consent }),
  )
}

export function openAnalyticsConsentPreferences() {
  window.dispatchEvent(new Event(openAnalyticsConsentEvent))
}
