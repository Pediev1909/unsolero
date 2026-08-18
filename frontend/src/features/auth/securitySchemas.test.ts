import { describe, expect, it } from 'vitest'

import { parseLoginResponse } from './schemas'
import { sessionsResponseSchema } from './securitySchemas'

describe('account security response boundaries', () => {
  it('recognizes an MFA challenge without requiring a user claim', () => {
    expect(
      parseLoginResponse({
        mfa_required: true,
        expires_at: '2026-08-17T12:00:00Z',
      }),
    ).toEqual({
      mfa_required: true,
      expires_at: '2026-08-17T12:00:00Z',
    })
  })

  it('keeps authentication material outside the session view model', () => {
    const parsed = sessionsResponseSchema.parse({
      sessions: [
        {
          id: '47b28acd-b2ee-48ef-b856-31b740fe5aa7',
          current: true,
          created_at: '2026-08-17T12:00:00Z',
          last_seen_at: '2026-08-17T12:01:00Z',
          expires_at: '2026-08-18T12:00:00Z',
          idle_expires_at: '2026-08-17T13:00:00Z',
          authentication_method: 'password_mfa',
          raw_token: 'must-not-survive-parsing',
          token_hash: 'must-not-survive-parsing',
        },
      ],
    })
    expect(parsed.sessions[0]).not.toHaveProperty('raw_token')
    expect(parsed.sessions[0]).not.toHaveProperty('token_hash')
  })
})
