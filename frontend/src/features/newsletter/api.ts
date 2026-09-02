import { apiRequest } from '../../lib/api/client'
import { newsletterReceiptSchema } from './schemas'

export function subscribeToNewsletter(email: string, source: string) {
  return apiRequest(
    '/newsletter/subscriptions',
    { method: 'POST', body: { email, source } },
    (value) => newsletterReceiptSchema.parse(value),
  )
}

export function confirmNewsletterSubscription(token: string) {
  return apiRequest(
    '/newsletter/confirmations',
    { method: 'POST', body: { token } },
    () => undefined,
  )
}

// Unsubscribe tokens do not expire and are not consumed: the server matches the
// row on the hash whatever its status, so a second click succeeds again. Only
// an unknown token fails, and it fails as `invalid_token`.
export function unsubscribeFromNewsletter(token: string) {
  return apiRequest(
    '/newsletter/unsubscriptions',
    { method: 'POST', body: { token } },
    () => undefined,
  )
}
