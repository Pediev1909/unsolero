import { afterEach, describe, expect, it, vi } from 'vitest'

import { apiRequest } from './client'

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
})
