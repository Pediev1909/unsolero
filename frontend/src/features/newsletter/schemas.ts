import { z } from 'zod'

export const newsletterSubscriptionSchema = z.object({
  email: z
    .string()
    .trim()
    .min(1, 'Enter your email address.')
    .max(254, 'Use no more than 254 characters.')
    .email('Enter a valid email address.'),
})

export type NewsletterSubscription = z.infer<
  typeof newsletterSubscriptionSchema
>

// The receipt is deliberately the same for a new, refreshed, or already
// confirmed address, so there is nothing else to parse.
export const newsletterReceiptSchema = z.object({
  recorded: z.literal(true),
})
