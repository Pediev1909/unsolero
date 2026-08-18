import { describe, expect, it } from 'vitest'

import { ApiError } from '../lib/api/client'
import { shouldRetryQuery } from './queryClient'

describe('query retry policy', () => {
  it('does not loop on authorization or validation failures', () => {
    expect(shouldRetryQuery(0, new ApiError(401, 'expired', 'expired'))).toBe(
      false,
    )
    expect(shouldRetryQuery(0, new ApiError(422, 'invalid', 'invalid'))).toBe(
      false,
    )
    expect(
      shouldRetryQuery(0, new ApiError(403, 'forbidden', 'forbidden')),
    ).toBe(false)
    expect(shouldRetryQuery(0, new ApiError(429, 'limited', 'limited'))).toBe(
      false,
    )
  })

  it('allows one retry for network and server failures only', () => {
    expect(shouldRetryQuery(0, new ApiError(0, 'network', 'network'))).toBe(
      true,
    )
    expect(shouldRetryQuery(0, new ApiError(503, 'down', 'down'))).toBe(true)
    expect(shouldRetryQuery(1, new ApiError(503, 'down', 'down'))).toBe(false)
  })
})
