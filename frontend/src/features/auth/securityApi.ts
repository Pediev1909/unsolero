import { apiRequest } from '../../lib/api/client'
import { parseAuthResponse, type AuthUser } from './schemas'
import {
  mfaEnabledSchema,
  mfaEnrollmentSchema,
  requestReceiptSchema,
  sessionsResponseSchema,
  type MFAEnrollment,
  type SecuritySession,
} from './securitySchemas'

export function requestEmailVerification(email: string) {
  return apiRequest(
    '/auth/email-verification/request',
    { method: 'POST', body: { email } },
    (value) => requestReceiptSchema.parse(value),
  )
}

export function completeEmailVerification(token: string) {
  return apiRequest(
    '/auth/email-verification/complete',
    { method: 'POST', body: { token } },
    () => undefined,
  )
}

export function requestPasswordReset(email: string) {
  return apiRequest(
    '/auth/password-reset/request',
    { method: 'POST', body: { email } },
    (value) => requestReceiptSchema.parse(value),
  )
}

export function completePasswordReset(token: string, password: string) {
  return apiRequest(
    '/auth/password-reset/complete',
    { method: 'POST', body: { token, password } },
    () => undefined,
  )
}

export function completeMfaLogin(code: string): Promise<AuthUser> {
  return apiRequest(
    '/auth/mfa/complete',
    { method: 'POST', body: { code } },
    parseAuthResponse,
  )
}

export function changePassword(currentPassword: string, newPassword: string) {
  return apiRequest(
    '/account/security/password',
    {
      method: 'POST',
      body: {
        current_password: currentPassword,
        new_password: newPassword,
      },
    },
    () => undefined,
  )
}

export async function listSessions(): Promise<SecuritySession[]> {
  return apiRequest(
    '/account/security/sessions',
    { method: 'GET' },
    (value) => sessionsResponseSchema.parse(value).sessions,
  )
}

export function revokeSession(sessionID: string) {
  return apiRequest(
    `/account/security/sessions/${encodeURIComponent(sessionID)}`,
    { method: 'DELETE' },
    () => undefined,
  )
}

export function revokeOtherSessions() {
  return apiRequest(
    '/account/security/sessions',
    { method: 'DELETE' },
    () => undefined,
  )
}

export function beginMfaEnrollment(password: string): Promise<MFAEnrollment> {
  return apiRequest(
    '/account/security/mfa/enroll',
    { method: 'POST', body: { password } },
    (value) => mfaEnrollmentSchema.parse(value),
  )
}

export function verifyMfaEnrollment(code: string): Promise<string[]> {
  return apiRequest(
    '/account/security/mfa/verify',
    { method: 'POST', body: { code } },
    (value) => mfaEnabledSchema.parse(value).recovery_codes,
  )
}

export function regenerateRecoveryCodes(code: string): Promise<string[]> {
  return apiRequest(
    '/account/security/mfa/recovery-codes',
    { method: 'POST', body: { code } },
    (value) => mfaEnabledSchema.parse(value).recovery_codes,
  )
}

export function stepUpMfa(code: string) {
  return apiRequest(
    '/account/security/mfa/step-up',
    { method: 'POST', body: { code } },
    () => undefined,
  )
}

export function deleteAccount(password: string, confirmation: string) {
  return apiRequest(
    '/account',
    { method: 'DELETE', body: { password, confirmation } },
    () => undefined,
  )
}
