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
})

const authResponseSchema = z.object({
  user: userSchema,
})

export type AuthUser = z.infer<typeof userSchema>

export function parseAuthResponse(value: unknown): AuthUser {
  return authResponseSchema.parse(value).user
}
