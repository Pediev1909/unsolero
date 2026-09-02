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
