import { z } from 'zod'

export const securitySessionSchema = z.object({
  id: z.string().uuid(),
  current: z.boolean(),
  created_at: z.string().datetime(),
  last_seen_at: z.string().datetime(),
  expires_at: z.string().datetime(),
  idle_expires_at: z.string().datetime(),
  mfa_authenticated_at: z.string().datetime().optional(),
  authentication_method: z.enum([
    'password',
    'password_mfa',
    'password_recovery',
  ]),
})

export const sessionsResponseSchema = z.object({
  sessions: z.array(securitySessionSchema),
})

export const requestReceiptSchema = z.object({
  recorded: z.literal(true),
  message: z.string(),
})

export const mfaEnrollmentSchema = z.object({
  secret: z.string().min(16),
  provisioning_uri: z.string().startsWith('otpauth://'),
})

export const mfaEnabledSchema = z.object({
  recovery_codes: z.array(z.string().min(10)).min(1),
})

export type SecuritySession = z.infer<typeof securitySessionSchema>
export type MFAEnrollment = z.infer<typeof mfaEnrollmentSchema>
