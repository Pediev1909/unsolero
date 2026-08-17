import { apiRequest, ApiError } from '../../lib/api/client'
import { parseAuthResponse, type AuthUser, type Credentials } from './schemas'

export async function getCurrentUser(): Promise<AuthUser | null> {
  try {
    return await apiRequest('/auth/me', { method: 'GET' }, parseAuthResponse)
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return null
    }
    throw error
  }
}

export function login(credentials: Credentials): Promise<AuthUser> {
  return apiRequest(
    '/auth/login',
    { method: 'POST', body: credentials },
    parseAuthResponse,
  )
}

export function register(credentials: Credentials): Promise<AuthUser> {
  return apiRequest(
    '/auth/register',
    { method: 'POST', body: credentials },
    parseAuthResponse,
  )
}

export function logout(): Promise<void> {
  return apiRequest('/auth/logout', { method: 'POST' }, () => undefined)
}
