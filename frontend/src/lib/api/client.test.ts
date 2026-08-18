import { afterEach, describe, expect, it, vi } from 'vitest'

import { apiRequest, onAuthenticationExpired } from './client'

afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('apiRequest', () => {
  it('normalizes a malformed successful response', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response('not-json', {
          status: 200,
          headers: { 'Content-Type': 'text/plain' },
        }),
      ),
    )

    await expect(
      apiRequest('/health', { method: 'GET' }, (value) => value),
    ).rejects.toMatchObject({
      code: 'unexpected_response',
      status: 200,
    })
  })

  it('returns a safe timeout error for an unresponsive request', async () => {
    vi.useFakeTimers()
    vi.stubGlobal(
      'fetch',
      vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
        return new Promise<Response>((_resolve, reject) => {
          init?.signal?.addEventListener('abort', () => {
            reject(new DOMException('Aborted', 'AbortError'))
          })
        })
      }),
    )

    const request = apiRequest(
      '/health',
      { method: 'GET', timeoutMs: 100 },
      (value) => value,
    )
    const expectation = expect(request).rejects.toMatchObject({
      code: 'request_timeout',
      status: 0,
    })
    await vi.advanceTimersByTimeAsync(100)
    await expectation
  })

  it('notifies the application when an authenticated request expires', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: { code: 'authentication_required', message: 'Sign in.' },
          }),
          { status: 401, headers: { 'Content-Type': 'application/json' } },
        ),
      ),
    )
    const listener = vi.fn()
    const unsubscribe = onAuthenticationExpired(listener)
    await expect(
      apiRequest('/account/export', { method: 'GET' }, (value) => value),
    ).rejects.toMatchObject({ status: 401 })
    unsubscribe()
    expect(listener).toHaveBeenCalledTimes(1)
  })

  it.each([403, 429, 500])(
    'preserves a validated safe API error for HTTP %s',
    async (status) => {
      vi.stubGlobal(
        'fetch',
        vi.fn().mockResolvedValue(
          new Response(
            JSON.stringify({
              error: {
                code: `safe_${status}`,
                message: 'Safe public message.',
              },
            }),
            { status, headers: { 'Content-Type': 'application/json' } },
          ),
        ),
      )
      await expect(
        apiRequest('/validation', { method: 'GET' }, (value) => value),
      ).rejects.toMatchObject({ status, code: `safe_${status}` })
    },
  )

  it('preserves caller cancellation for query lifecycle handling', async () => {
    const controller = new AbortController()
    vi.stubGlobal(
      'fetch',
      vi.fn(
        (_input: RequestInfo | URL, init?: RequestInit) =>
          new Promise<Response>((_resolve, reject) => {
            init?.signal?.addEventListener('abort', () =>
              reject(new DOMException('Aborted', 'AbortError')),
            )
          }),
      ),
    )
    const request = apiRequest(
      '/health',
      { method: 'GET', signal: controller.signal },
      (value) => value,
    )
    controller.abort()
    await expect(request).rejects.toMatchObject({ name: 'AbortError' })
  })
})
