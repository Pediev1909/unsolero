import { z } from 'zod'

export const credentialsSchema = z.object({
  email: z
    .string()
    .trim()
    .min(1, 'Enter your email address.')
    .email('Enter a valid email address.'),
  password: z
    .string()
    .min(12, 'Use at least 12 characters.')
    .max(128, 'Use no more than 128 characters.'),
})

export type Credentials = z.infer<typeof credentialsSchema>

export const userSchema = z.object({
  id: z.string().min(1),
  email: z.string().email(),
  roles: z.array(z.string()),
  email_verified: z.boolean().default(false),
  mfa_enabled: z.boolean().default(false),
})

const authResponseSchema = z.object({
  user: userSchema,
})

const loginMfaSchema = z.object({
  mfa_required: z.literal(true),
  expires_at: z.string().datetime(),
})

const registrationReceiptSchema = z.object({
  recorded: z.literal(true),
  message: z.string(),
})

export type AuthUser = z.infer<typeof userSchema>
export type LoginResult = AuthUser | z.infer<typeof loginMfaSchema>
export type RegistrationReceipt = z.infer<typeof registrationReceiptSchema>

export function parseAuthResponse(value: unknown): AuthUser {
  return authResponseSchema.parse(value).user
}

export function parseLoginResponse(value: unknown): LoginResult {
  const mfa = loginMfaSchema.safeParse(value)
  return mfa.success ? mfa.data : parseAuthResponse(value)
}

export function parseRegistrationReceipt(value: unknown): RegistrationReceipt {
  return registrationReceiptSchema.parse(value)
}
